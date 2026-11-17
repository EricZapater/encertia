package downloader

import (
	"runtime"
	"strings"
	"testing"
)

func TestParseProtocolURI(t *testing.T) {
	tests := []struct {
		name         string
		uri          string
		wantAction   string
		wantVideoURL string
		wantErr      bool
	}{
		{
			name:         "URI estàndard amb query param",
			uri:          "encertia://download?url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DdQw4w9WgXcQ",
			wantAction:   "download",
			wantVideoURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantErr:      false,
		},
		{
			name:         "URI amb cometes de registre de Windows",
			uri:          `"encertia://download?url=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DdQw4w9WgXcQ"`,
			wantAction:   "download",
			wantVideoURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantErr:      false,
		},
		{
			name:         "URI Instagram Reel",
			uri:          "encertia://download?url=https%3A%2F%2Fwww.instagram.com%2Freel%2FC1234567%2F",
			wantAction:   "download",
			wantVideoURL: "https://www.instagram.com/reel/C1234567/",
			wantErr:      false,
		},
		{
			name:         "Esquema no vàlid",
			uri:          "http://example.com/download",
			wantAction:   "",
			wantVideoURL: "",
			wantErr:      true,
		},
		{
			name:         "Sense paràmetre url",
			uri:          "encertia://download",
			wantAction:   "download",
			wantVideoURL: "",
			wantErr:      true,
		},
		{
			name:         "URL no http",
			uri:          "encertia://download?url=ftp%3A%2F%2Fexample.com",
			wantAction:   "download",
			wantVideoURL: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, videoURL, err := ParseProtocolURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseProtocolURI() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if action != tt.wantAction {
					t.Errorf("action = %v, want %v", action, tt.wantAction)
				}
				if videoURL != tt.wantVideoURL {
					t.Errorf("videoURL = %v, want %v", videoURL, tt.wantVideoURL)
				}
			}
		})
	}
}

func TestGetDownloadsDir(t *testing.T) {
	dir := GetDownloadsDir()
	if strings.TrimSpace(dir) == "" {
		t.Errorf("GetDownloadsDir() ha retornat una cadena buida")
	}
}

func TestGetOfficialYtDlpURL(t *testing.T) {
	url := GetOfficialYtDlpURL()
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(url, "yt-dlp.exe") {
			t.Errorf("esperat yt-dlp.exe a Windows, obtingut: %s", url)
		}
	} else if runtime.GOOS == "darwin" {
		if !strings.HasSuffix(url, "yt-dlp_macos") {
			t.Errorf("esperat yt-dlp_macos a macOS, obtingut: %s", url)
		}
	} else if runtime.GOOS == "linux" {
		if !strings.HasSuffix(url, "yt-dlp") {
			t.Errorf("esperat yt-dlp a Linux, obtingut: %s", url)
		}
	}
}
