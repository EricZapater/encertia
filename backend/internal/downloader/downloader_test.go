package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/encertia/backend/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockExecutor struct {
	runFunc func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (m *mockExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, name, args...)
	}
	return []byte{}, nil
}

type mockUserRepo struct {
	users map[uuid.UUID]*user.UserDB
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*user.UserDB, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func TestValidateURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "Buid",
			url:     "",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "Esquema no http",
			url:     "ftp://example.com/video.mp4",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "Fitxer local file://",
			url:     "file:///etc/passwd",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "Localhost",
			url:     "http://localhost:8080/video",
			wantErr: ErrDisallowedHost,
		},
		{
			name:    "Loopback 127.0.0.1",
			url:     "http://127.0.0.1:8080/video",
			wantErr: ErrDisallowedHost,
		},
		{
			name:    "IP Privada 192.168.1.1",
			url:     "http://192.168.1.1/video",
			wantErr: ErrDisallowedHost,
		},
		{
			name:    "IP Privada 10.0.0.1",
			url:     "http://10.0.0.1/video",
			wantErr: ErrDisallowedHost,
		},
		{
			name:    "URL Vàlida YouTube",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantErr: nil,
		},
		{
			name:    "URL Vàlida Instagram",
			url:     "https://www.instagram.com/reel/C123456789/",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateURL(tc.url)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("esperat error %v, obtingut: %v", tc.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("esperat sense error, obtingut: %v", err)
				}
			}
		})
	}
}

func TestGetVideoInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("Èxit retornant informació", func(t *testing.T) {
		mockJSON := `{
			"id": "dQw4w9WgXcQ",
			"title": "Rick Astley - Never Gonna Give You Up",
			"thumbnail": "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg",
			"duration": 213,
			"extractor": "youtube",
			"uploader": "Rick Astley",
			"description": "Official video"
		}`

		executor := &mockExecutor{
			runFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name != "yt-dlp" {
					t.Errorf("esperat yt-dlp, obtingut: %s", name)
				}
				// Comprova que conté -4 per forçar IPv4
				hasIPv4 := false
				hasNoCallHome := false
				hasMaxDownloads := false
				hasJSRuntimes := false
				for i, arg := range args {
					if arg == "-4" {
						hasIPv4 = true
					}
					if arg == "--no-call-home" {
						hasNoCallHome = true
					}
					if arg == "--max-downloads" {
						hasMaxDownloads = true
					}
					if arg == "--js-runtimes" && i+1 < len(args) && args[i+1] == "node" {
						hasJSRuntimes = true
					}
				}
				if !hasIPv4 {
					t.Errorf("s'esperava el flag -4 per forçar IPv4")
				}
				if hasNoCallHome {
					t.Errorf("--no-call-home està obsolet i no hauria de ser present")
				}
				if hasMaxDownloads {
					t.Errorf("--max-downloads causa exit code 101 i no hauria de ser present a GetVideoInfo")
				}
				if !hasJSRuntimes {
					t.Errorf("s'esperava --js-runtimes node per a resoldre reptes de YouTube")
				}
				return []byte(mockJSON), nil
			},
		}

		svc := NewService(executor)
		info, err := svc.GetVideoInfo(ctx, VideoInfoRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})
		if err != nil {
			t.Fatalf("no s'esperava error: %v", err)
		}
		if info.ID != "dQw4w9WgXcQ" {
			t.Errorf("esperat id dQw4w9WgXcQ, obtingut: %s", info.ID)
		}
		if info.Title != "Rick Astley - Never Gonna Give You Up" {
			t.Errorf("esperat títol correcte, obtingut: %s", info.Title)
		}
		if info.Duration != 213 {
			t.Errorf("esperada durada 213, obtingut: %d", info.Duration)
		}
		if info.Extractor != "youtube" {
			t.Errorf("esperat extractor youtube, obtingut: %s", info.Extractor)
		}
	})

	t.Run("yt-dlp no instal·lat", func(t *testing.T) {
		executor := &mockExecutor{
			runFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("executable file not found in $PATH")
			},
		}

		svc := NewService(executor)
		_, err := svc.GetVideoInfo(ctx, VideoInfoRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})
		if !errors.Is(err, ErrToolNotInstalled) {
			t.Fatalf("esperat ErrToolNotInstalled, obtingut: %v", err)
		}
	})

	t.Run("Amb cookies d'usuari", func(t *testing.T) {
		mockJSON := `{"id":"test","title":"Test with cookies","duration":10,"extractor":"test"}`
		cookiesReceived := false

		executor := &mockExecutor{
			runFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				for i, arg := range args {
					if arg == "--cookies" && i+1 < len(args) {
						cookiesReceived = true
						break
					}
				}
				return []byte(mockJSON), nil
			},
		}

		svc := NewService(executor)
		info, err := svc.GetVideoInfo(ctx, VideoInfoRequest{
			URL:     "https://www.youtube.com/watch?v=test",
			Cookies: "# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tTRUE\t0\tLOGIN_INFO\tval",
		})
		if err != nil {
			t.Fatalf("no s'esperava error: %v", err)
		}
		if !cookiesReceived {
			t.Errorf("s'esperava que el flag --cookies s'hagués passat a yt-dlp")
		}
		if info.Title != "Test with cookies" {
			t.Errorf("títol incorrecte: %s", info.Title)
		}
	})
}

func TestDownloadVideo(t *testing.T) {
	ctx := context.Background()

	t.Run("Èxit descarregant i netejant", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "downloader_test")
		if err != nil {
			t.Fatalf("error creant directori temporal: %v", err)
		}
		defer os.RemoveAll(tempDir)

		executor := &mockExecutor{
			runFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				var destFile string
				for i, arg := range args {
					if arg == "-o" && i+1 < len(args) {
						destPattern := args[i+1]
						destFile = filepath.Join(filepath.Dir(destPattern), filepath.Base(destPattern))
						destFile = destFile[:len(destFile)-len(".%(ext)s")] + ".mp4"
						break
					}
				}
				if destFile == "" {
					destFile = filepath.Join(tempDir, "sample.mp4")
				}
				_ = os.WriteFile(destFile, []byte("fake video content"), 0644)

				output := destFile + "\nMy Cool Video\n"
				return []byte(output), nil
			},
		}

		svc := &service{
			executor: executor,
			tempDir:  tempDir,
		}

		filePath, filename, cleanup, err := svc.DownloadVideo(ctx, DownloadRequest{URL: "https://www.youtube.com/watch?v=sample"})
		if err != nil {
			t.Fatalf("no s'esperava error: %v", err)
		}
		if _, statErr := os.Stat(filePath); statErr != nil {
			t.Fatalf("el fitxer no existeix: %s", filePath)
		}
		if !strings.Contains(filename, "My Cool Video.mp4") {
			t.Errorf("el nom del fitxer no és l'esperat: %s", filename)
		}

		cleanup()
		if _, statErr := os.Stat(filePath); !os.IsNotExist(statErr) {
			t.Errorf("s'esperava que el fitxer s'hagués eliminat en el cleanup: %s", filePath)
		}
	})
}

func TestDownloaderHandler_Permissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminID := uuid.New()
	permittedTeacherID := uuid.New()
	nonPermittedTeacherID := uuid.New()

	userRepo := &mockUserRepo{
		users: map[uuid.UUID]*user.UserDB{
			adminID: {
				ID:                adminID,
				Role:              "admin",
				CanDownloadVideos: true,
			},
			permittedTeacherID: {
				ID:                permittedTeacherID,
				Role:              "teacher",
				CanDownloadVideos: true,
			},
			nonPermittedTeacherID: {
				ID:                nonPermittedTeacherID,
				Role:              "teacher",
				CanDownloadVideos: false,
			},
		},
	}

	executor := &mockExecutor{
		runFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte(`{"id":"vid1","title":"Test Video","duration":10,"extractor":"test"}`), nil
		},
	}
	svc := NewService(executor)
	handler := NewHandler(svc, userRepo)

	setupRouter := func(userID *uuid.UUID, role string) *gin.Engine {
		r := gin.New()
		fakeAuth := func(c *gin.Context) {
			if userID != nil {
				c.Set(shared.CtxKeyUserID, userID.String())
				c.Set(shared.CtxKeyUserRole, role)
			}
			c.Next()
		}
		handler.RegisterRoutes(r.Group(""), fakeAuth)
		return r
	}

	body, _ := json.Marshal(VideoInfoRequest{URL: "https://www.youtube.com/watch?v=vid1"})

	t.Run("Admin té accés", func(t *testing.T) {
		r := setupRouter(&adminID, "admin")
		req, _ := http.NewRequest(http.MethodPost, "/downloader/info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("esperat status 200, obtingut: %d", w.Code)
		}
	})

	t.Run("Usuari amb can_download_videos=true té accés", func(t *testing.T) {
		r := setupRouter(&permittedTeacherID, "teacher")
		req, _ := http.NewRequest(http.MethodPost, "/downloader/info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("esperat status 200, obtingut: %d", w.Code)
		}
	})

	t.Run("Usuari sense permís rep 403 Forbidden", func(t *testing.T) {
		r := setupRouter(&nonPermittedTeacherID, "teacher")
		req, _ := http.NewRequest(http.MethodPost, "/downloader/info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("esperat status 403, obtingut: %d", w.Code)
		}
	})

	t.Run("Sense autenticació rep 401 Unauthorized", func(t *testing.T) {
		r := setupRouter(nil, "")
		req, _ := http.NewRequest(http.MethodPost, "/downloader/info", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("esperat status 401, obtingut: %d", w.Code)
		}
	})
}

func TestSanitizeVideoURL(t *testing.T) {
	parsed1, _ := ValidateURL("https://www.youtube.com/watch?v=jVOluNDjc5s&list=RDjVOluNDjc5s&index=1")
	sanitized1 := sanitizeVideoURL(parsed1)
	if strings.Contains(sanitized1, "list=") || strings.Contains(sanitized1, "index=") {
		t.Errorf("s'esperava eliminar list i index de la URL de YouTube: %s", sanitized1)
	}
	if !strings.Contains(sanitized1, "v=jVOluNDjc5s") {
		t.Errorf("s'hauria de mantenir el paràmetre v: %s", sanitized1)
	}

	// Altres plataformes no s'han de modificar
	parsed2, _ := ValidateURL("https://www.linkedin.com/posts/activity-123456?list=sample")
	sanitized2 := sanitizeVideoURL(parsed2)
	if !strings.Contains(sanitized2, "list=sample") {
		t.Errorf("no s'haurien de tocar paràmetres de LinkedIn: %s", sanitized2)
	}
}

func TestCleanStderr(t *testing.T) {
	raw := "Deprecated Feature: The following options have been deprecated: --no-call-home\n" +
		"Please remove them from your command/configuration to avoid future errors.\n" +
		"WARNING: [youtube] Some warning\n" +
		"ERROR: [youtube] video unavailable\n"

	cleaned := cleanStderr(raw)
	if strings.Contains(cleaned, "Deprecated Feature") {
		t.Errorf("cleanStderr hauria de filtrar avisos de deprecació")
	}
	if strings.Contains(cleaned, "WARNING:") {
		t.Errorf("cleanStderr hauria de filtrar warnings comuns")
	}
	if !strings.Contains(cleaned, "ERROR: [youtube] video unavailable") {
		t.Errorf("cleanStderr hauria de conservar l'error real, obtingut: %s", cleaned)
	}
}
