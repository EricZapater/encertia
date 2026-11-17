package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ParseProtocolURI analitza una URI del tipus encertia://download?url=...
func ParseProtocolURI(uriStr string) (action string, videoURL string, err error) {
	trimmed := strings.TrimSpace(uriStr)
	// Elimina cometes si vénen del registre de Windows ("%1")
	trimmed = strings.Trim(trimmed, `"'`)

	if !strings.HasPrefix(strings.ToLower(trimmed), "encertia://") {
		return "", "", fmt.Errorf("esquema de protocol no vàlid: %s", uriStr)
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("error analitzant la URI: %w", err)
	}

	action = strings.ToLower(u.Host)
	if action == "" {
		action = strings.TrimPrefix(strings.ToLower(u.Path), "/")
	}

	// 1. Cercar paràmetre query ?url=
	videoURL = u.Query().Get("url")

	// 2. Fallback: url codificada directament al path: encertia://download/https://...
	if videoURL == "" {
		path := strings.TrimPrefix(u.Path, "/")
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			videoURL = path
		} else if unescaped, err := url.QueryUnescape(path); err == nil && (strings.HasPrefix(unescaped, "http://") || strings.HasPrefix(unescaped, "https://")) {
			videoURL = unescaped
		}
	}

	videoURL = strings.TrimSpace(videoURL)
	if videoURL == "" {
		return action, "", errors.New("no s'ha proporcionat cap adreça URL de vídeo a la petició")
	}

	if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
		return action, "", fmt.Errorf("la URL proporcionada no és vàlida (ha de començar per http:// o https://): %s", videoURL)
	}

	return action, videoURL, nil
}

// GetDownloadsDir obté la ruta absoluta de la carpeta de baixades de l'usuari actual.
func GetDownloadsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	candidates := []string{"Downloads", "Baixades", "Descargas"}
	for _, c := range candidates {
		p := filepath.Join(home, c)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}

	defaultDownloads := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(defaultDownloads, 0755)
	return defaultDownloads
}

// GetOfficialYtDlpURL retorna la URL oficial de descàrrega del binari autònom de yt-dlp per al sistema operatiu actual.
func GetOfficialYtDlpURL() string {
	switch runtime.GOOS {
	case "windows":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case "darwin":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	case "linux":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	default:
		return ""
	}
}

// EnsureYtDlp assegura que el motor de descàrrega autònom yt-dlp estigui disponible a l'equip.
func EnsureYtDlp() (string, error) {
	// 1. Cercar al PATH global del sistema
	binName := "yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
	}

	if path, err := exec.LookPath(binName); err == nil {
		return path, nil
	}

	// 2. Directori local de l'agent: ~/.encertia/bin/
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obtenint directori d'usuari: %w", err)
	}
	toolsDir := filepath.Join(home, ".encertia", "bin")
	_ = os.MkdirAll(toolsDir, 0755)

	targetPath := filepath.Join(toolsDir, binName)
	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() && info.Size() > 0 {
		return targetPath, nil
	}

	// 3. Descarregar des de GitHub Releases
	downloadURL := GetOfficialYtDlpURL()
	if downloadURL == "" {
		return "", fmt.Errorf("sistema operatiu no suportat per a descàrrega automàtica: %s", runtime.GOOS)
	}

	fmt.Println("📥 Preparant el motor de descàrrega autònom per primer cop...")
	fmt.Println("   (Aquest pas només es realitza una sola vegada)")
	fmt.Printf("   Descàrrega des de: %s\n", downloadURL)

	tempFile := targetPath + ".download"
	if err := downloadFile(downloadURL, tempFile); err != nil {
		_ = os.Remove(tempFile)
		return "", fmt.Errorf("error descarregant el motor yt-dlp: %w", err)
	}

	if err := os.Rename(tempFile, targetPath); err != nil {
		return "", fmt.Errorf("error desant el motor yt-dlp: %w", err)
	}

	_ = os.Chmod(targetPath, 0755)
	fmt.Println("   ✔️ Motor preparat i a punt!")
	return targetPath, nil
}

// downloadFile descarrega un fitxer des d'una URL mostrant el progrés a la consola.
func downloadFile(urlStr, destPath string) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("el servidor ha retornat codi HTTP %d", resp.StatusCode)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	totalBytes := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	lastPrint := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if time.Since(lastPrint) > 500*time.Millisecond {
				if totalBytes > 0 {
					pct := float64(downloaded) / float64(totalBytes) * 100
					fmt.Printf("\r   Progrés: %.1f%% (%.1f / %.1f MB)", pct, float64(downloaded)/1024/1024, float64(totalBytes)/1024/1024)
				} else {
					fmt.Printf("\r   Descarregats: %.1f MB", float64(downloaded)/1024/1024)
				}
				lastPrint = time.Now()
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
	}

	fmt.Println("\r   Progrés: 100% completat!                        ")
	return nil
}

// RunDownload executa la descàrrega d'un vídeo mitjançant yt-dlp i el desa a la carpeta de descàrregues de l'equip.
func RunDownload(videoURL string) error {
	ytDlpPath, err := EnsureYtDlp()
	if err != nil {
		return err
	}

	downloadsDir := GetDownloadsDir()
	outputPattern := filepath.Join(downloadsDir, "%(title)s [%(id)s].%(ext)s")

	fmt.Println()
	fmt.Printf("[%s] 🚀 Iniciant descàrrega de: %s\n", time.Now().Format("15:04:05"), videoURL)
	fmt.Printf("[%s] 📂 Carpeta de destí: %s\n\n", time.Now().Format("15:04:05"), downloadsDir)

	args := []string{
		"-4",
		"--no-playlist",
		"--no-warnings",
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format", "mp4",
		"-o", outputPattern,
		videoURL,
	}

	cmd := exec.Command(ytDlpPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("la descàrrega ha fallat: %w", err)
	}

	fmt.Println("\n==================================================")
	fmt.Println("✅ VÍDEO DESCARREGAT CORRECTAMENT A L'EQUIP!")
	fmt.Printf("   Ubicació: %s\n", downloadsDir)
	fmt.Println("==================================================")

	openFolder(downloadsDir)
	return nil
}

// openFolder obre la carpeta al navegador d'arxius del sistema operatiu.
func openFolder(dir string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("explorer.exe", dir).Start()
	case "darwin":
		_ = exec.Command("open", dir).Start()
	case "linux":
		_ = exec.Command("xdg-open", dir).Start()
	}
}

// HandleProtocolLaunch gestiona la invocació de l'agent a través del protocol encertia://
func HandleProtocolLaunch(uriStr string) {
	fmt.Println("==================================================")
	fmt.Println("        ENCERTIA — DESCARREGADOR DE VÍDEOS        ")
	fmt.Println("==================================================")

	action, videoURL, err := ParseProtocolURI(uriStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error en el protocol: %v\n", err)
		waitEnterToExit()
		os.Exit(1)
	}

	if action != "download" && action != "" {
		fmt.Fprintf(os.Stderr, "❌ Acció desconeguda: %s\n", action)
		waitEnterToExit()
		os.Exit(1)
	}

	fmt.Printf("[%s] 🔗 Ordre de descàrrega rebuda per a: %s\n", time.Now().Format("15:04:05"), videoURL)

	if err := RunDownload(videoURL); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		waitEnterToExit()
		os.Exit(1)
	}

	waitBeforeAutoClose(5)
}

// HandleDirectDownload gestiona la invocació per línia de comandes: encertia-agent download <url>
func HandleDirectDownload(videoURL string) {
	fmt.Println("==================================================")
	fmt.Println("        ENCERTIA — DESCARREGADOR DE VÍDEOS        ")
	fmt.Println("==================================================")

	trimmed := strings.TrimSpace(videoURL)
	if trimmed == "" {
		fmt.Fprintln(os.Stderr, "❌ Cal especificar una URL vàlida.")
		os.Exit(1)
	}

	if err := RunDownload(trimmed); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		waitEnterToExit()
		os.Exit(1)
	}

	waitBeforeAutoClose(3)
}

func waitEnterToExit() {
	fmt.Println("\nPrem Enter per sortir d'aquesta finestra...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func waitBeforeAutoClose(seconds int) {
	fmt.Printf("\nAquesta finestra es tancarà en %d segons (o prem Enter)...", seconds)
	done := make(chan struct{})
	go func() {
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Duration(seconds) * time.Second):
		fmt.Println()
	}
}

// RegisterProtocolSilently registra de manera transparent el protocol encertia:// al sistema operatiu.
func RegisterProtocolSilently() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return
	}

	switch runtime.GOOS {
	case "windows":
		registerWindowsProtocol(exePath)
	case "linux":
		registerLinuxProtocol(exePath)
	}
}

func registerWindowsProtocol(exePath string) {
	// Registre a HKCU (Current User) no requereix privilegis d'administrador
	_ = exec.Command("reg", "add", `HKCU\Software\Classes\encertia`, "/ve", "/d", "URL:Encertia Protocol", "/f").Run()
	_ = exec.Command("reg", "add", `HKCU\Software\Classes\encertia`, "/v", "URL Protocol", "/d", "", "/f").Run()
	_ = exec.Command("reg", "add", `HKCU\Software\Classes\encertia\shell\open\command`, "/ve", "/d", fmt.Sprintf(`"%s" "%%1"`, exePath), "/f").Run()
}

func registerLinuxProtocol(exePath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	appsDir := filepath.Join(home, ".local", "share", "applications")
	_ = os.MkdirAll(appsDir, 0755)

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=Encertia Agent
Exec="%s" "%%u"
Type=Application
Terminal=true
MimeType=x-scheme-handler/encertia;
NoDisplay=true
`, exePath)

	desktopFile := filepath.Join(appsDir, "encertia-agent.desktop")
	if err := os.WriteFile(desktopFile, []byte(desktopContent), 0644); err == nil {
		_ = exec.Command("xdg-mime", "default", "encertia-agent.desktop", "x-scheme-handler/encertia").Run()
	}
}
