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

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

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
	GetVideoInfo(ctx context.Context, req VideoInfoRequest) (*VideoInfoResponse, error)
	DownloadVideo(ctx context.Context, req DownloadRequest) (filePath string, filename string, cleanup func(), err error)
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

// prepareCookies desa les cookies de l'usuari en un fitxer temporal o utilitza les cookies del servidor si estan configurades.
func (s *service) prepareCookies(userCookies string) (string, func(), error) {
	trimmed := strings.TrimSpace(userCookies)
	if trimmed != "" {
		cookieFile := filepath.Join(s.tempDir, fmt.Sprintf("cookies_%s.txt", uuid.New().String()))
		if err := os.WriteFile(cookieFile, []byte(trimmed), 0600); err != nil {
			return "", nil, err
		}
		return cookieFile, func() { _ = os.Remove(cookieFile) }, nil
	}

	if envPath := os.Getenv("YTDLP_COOKIES_PATH"); envPath != "" && fileExists(envPath) {
		return envPath, func() {}, nil
	}
	if fileExists("/app/cookies.txt") {
		return "/app/cookies.txt", func() {}, nil
	}
	if fileExists("/tmp/cookies.txt") {
		return "/tmp/cookies.txt", func() {}, nil
	}

	return "", func() {}, nil
}

// buildBaseArgs construeix els arguments bàsics recomanats per a robustesa en servidors i xarxes.
func (s *service) buildBaseArgs() []string {
	return []string{
		"-4", // Força IPv4 per evitar errors ENETUNREACH / Errno 101 en contenidors sense sortida IPv6
		"--no-playlist",
		"--no-warnings",
		"--no-call-home",
		"--user-agent", defaultUserAgent,
		"--extractor-args", "youtube:player_client=ios,android,mweb",
	}
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

func (s *service) GetVideoInfo(ctx context.Context, req VideoInfoRequest) (*VideoInfoResponse, error) {
	parsedURL, err := ValidateURL(req.URL)
	if err != nil {
		return nil, err
	}

	// Timeout de seguretat de 30 segons per obtenir la informació
	infoCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cookiePath, cookieCleanup, err := s.prepareCookies(req.Cookies)
	if err != nil {
		return nil, fmt.Errorf("error preparant cookies: %w", err)
	}
	defer cookieCleanup()

	args := s.buildBaseArgs()
	if cookiePath != "" {
		args = append(args, "--cookies", cookiePath)
	}

	args = append(args,
		"--dump-single-json",
		"--playlist-items", "1",
		"--max-downloads", "1",
		parsedURL.String(),
	)

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

func (s *service) DownloadVideo(ctx context.Context, req DownloadRequest) (string, string, func(), error) {
	parsedURL, err := ValidateURL(req.URL)
	if err != nil {
		return "", "", nil, err
	}

	// Timeout de 10 minuts per a la descàrrega
	dlCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Assegura directori temporal
	_ = os.MkdirAll(s.tempDir, 0755)

	cookiePath, cookieCleanup, err := s.prepareCookies(req.Cookies)
	if err != nil {
		return "", "", nil, fmt.Errorf("error preparant cookies: %w", err)
	}
	defer cookieCleanup()

	downloadID := uuid.New().String()
	outputPathPattern := filepath.Join(s.tempDir, fmt.Sprintf("%s.%%(ext)s", downloadID))

	args := s.buildBaseArgs()
	if cookiePath != "" {
		args = append(args, "--cookies", cookiePath)
	}

	args = append(args,
		"--playlist-items", "1",
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
		"--print", "after_move:filepath",
		"--print", "title",
		"-o", outputPathPattern,
		parsedURL.String(),
	)

	out, err := s.executor.Run(dlCtx, "yt-dlp", args...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
			return "", "", nil, ErrToolNotInstalled
		}
		return "", "", nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	// Parsejar sortida: primera línia és el filepath, segona és el títol
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
