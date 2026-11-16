package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidURL       = errors.New("l'adreça URL proporcionada no és vàlida")
	ErrDisallowedHost   = errors.New("l'adreça URL no està permesa (adreces privades o locals)")
	ErrToolNotInstalled = errors.New("el motor de descàrrega yt-dlp no està instal·lat al servidor")
	ErrDownloadFailed   = errors.New("no s'ha pogut obtenir o descarregar el vídeo de la plataforma")
)

// CommandExecutor permet executar processos del sistema operatiu (fàcilment mockejable per a tests).
type CommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// defaultExecutor executa ordres reals del sistema operatiu mitjançant exec.CommandContext.
type defaultExecutor struct{}

func (e *defaultExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("%w: %s", err, stderrStr)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// Service defineix les operacions de negoci del descarregador de vídeos.
type Service interface {
	GetVideoInfo(ctx context.Context, rawURL string) (*VideoInfoResponse, error)
	DownloadVideo(ctx context.Context, rawURL string) (filePath string, filename string, cleanup func(), err error)
}

type service struct {
	executor CommandExecutor
	tempDir  string
}

// NewService crea una nova instància del servei amb l'executor per defecte.
func NewService(customExecutor ...CommandExecutor) Service {
	var exec CommandExecutor = &defaultExecutor{}
	if len(customExecutor) > 0 && customExecutor[0] != nil {
		exec = customExecutor[0]
	}

	tempDir := filepath.Join(os.TempDir(), "encertia_downloads")
	_ = os.MkdirAll(tempDir, 0755)

	return &service{
		executor: exec,
		tempDir:  tempDir,
	}
}

// ValidateURL comprova que la URL sigui vàlida i segura (protecció anti-SSRF).
func ValidateURL(rawURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, ErrInvalidURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, ErrInvalidURL
	}

	host := parsed.Hostname()
	if host == "" {
		return nil, ErrInvalidURL
	}

	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || lowerHost == "127.0.0.1" || lowerHost == "::1" || lowerHost == "0.0.0.0" {
		return nil, ErrDisallowedHost
	}

	// Comprovació de rangs privats d'IP
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return nil, ErrDisallowedHost
		}
	}

	return parsed, nil
}

// rawYtDlpJSON estructura per deserialitzar la sortida de yt-dlp --dump-single-json.
type rawYtDlpJSON struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Thumbnail   string  `json:"thumbnail"`
	Duration    float64 `json:"duration"`
	Extractor   string  `json:"extractor"`
	Uploader    string  `json:"uploader"`
	Description string  `json:"description"`
}

func (s *service) GetVideoInfo(ctx context.Context, rawURL string) (*VideoInfoResponse, error) {
	parsedURL, err := ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Timeout de seguretat de 30 segons per obtenir la informació
	infoCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	args := []string{
		"--dump-single-json",
		"--no-playlist",
		"--no-warnings",
		"--no-call-home",
		"--max-downloads", "1",
		parsedURL.String(),
	}

	out, err := s.executor.Run(infoCtx, "yt-dlp", args...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
			return nil, ErrToolNotInstalled
		}
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	var raw rawYtDlpJSON
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("error processant dades del vídeo: %w", err)
	}

	if raw.Title == "" {
		raw.Title = "Vídeo sense títol"
	}

	extractor := raw.Extractor
	if extractor == "" {
		extractor = parsedURL.Hostname()
	}

	return &VideoInfoResponse{
		ID:          raw.ID,
		Title:       raw.Title,
		Thumbnail:   raw.Thumbnail,
		Duration:    int(raw.Duration),
		Extractor:   extractor,
		Uploader:    raw.Uploader,
		Description: raw.Description,
	}, nil
}

func (s *service) DownloadVideo(ctx context.Context, rawURL string) (string, string, func(), error) {
	parsedURL, err := ValidateURL(rawURL)
	if err != nil {
		return "", "", nil, err
	}

	// Timeout de 10 minuts per a la descàrrega
	dlCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Assegura directori temporal
	_ = os.MkdirAll(s.tempDir, 0755)

	downloadID := uuid.New().String()
	outputPathPattern := filepath.Join(s.tempDir, fmt.Sprintf("%s.%%(ext)s", downloadID))

	args := []string{
		"--no-playlist",
		"--no-warnings",
		"--no-call-home",
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
		"--print", "after_move:filepath",
		"--print", "title",
		"-o", outputPathPattern,
		parsedURL.String(),
	}

	out, err := s.executor.Run(dlCtx, "yt-dlp", args...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
			return "", "", nil, ErrToolNotInstalled
		}
		return "", "", nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	// Parsejar sortida: primer línia és el filepath, segona és el títol
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var generatedFile string
	var title string

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}
		if generatedFile == "" && (strings.HasSuffix(trimmed, ".mp4") || strings.HasPrefix(trimmed, s.tempDir)) {
			generatedFile = trimmed
		} else if title == "" {
			title = trimmed
		}
	}

	// Fallback si no troba el path per la sortida impresa: busca el fitxer generat amb el downloadID
	if generatedFile == "" || !fileExists(generatedFile) {
		patternMatches, _ := filepath.Glob(filepath.Join(s.tempDir, downloadID+".*"))
		if len(patternMatches) > 0 {
			generatedFile = patternMatches[0]
		}
	}

	if generatedFile == "" || !fileExists(generatedFile) {
		return "", "", nil, fmt.Errorf("%w: no s'ha generat cap fitxer de vídeo", ErrDownloadFailed)
	}

	if title == "" {
		title = "video"
	}

	safeTitle := sanitizeFilename(title)
	ext := filepath.Ext(generatedFile)
	if ext == "" {
		ext = ".mp4"
	}
	cleanFilename := safeTitle + ext

	cleanup := func() {
		_ = os.Remove(generatedFile)
	}

	return generatedFile, cleanFilename, cleanup, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

var invalidFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9_\-\. \p{L}]`)

func sanitizeFilename(name string) string {
	cleaned := invalidFilenameChars.ReplaceAllString(name, "")
	cleaned = strings.TrimSpace(cleaned)
	if len(cleaned) > 80 {
		cleaned = cleaned[:80]
	}
	if cleaned == "" {
		cleaned = "video"
	}
	return cleaned
}
