package shared_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/encertia/backend/internal/shared"
)

func TestStorage_LocalUploadValidation(t *testing.T) {
	tempDir := t.TempDir()
	svc := shared.NewStorageService(shared.StorageConfig{
		LocalUploadDir: tempDir,
	})
	ctx := context.Background()

	// 1. Empty data
	_, appErrEmpty := svc.UploadImageBytes(ctx, []byte{}, "empty.png", "image/png")
	if appErrEmpty == nil || appErrEmpty.StatusCode != 400 {
		t.Errorf("expected 400 for empty data, got %v", appErrEmpty)
	}

	// 2. Disallowed format (text/plain)
	_, appErrInvalidMIME := svc.UploadImageBytes(ctx, []byte("plain text"), "test.txt", "text/plain")
	if appErrInvalidMIME == nil || appErrInvalidMIME.StatusCode != 400 {
		t.Errorf("expected 400 for text/plain, got %v", appErrInvalidMIME)
	}

	// 3. Exceeding max size (> 5MB)
	tooLarge := make([]byte, 5*1024*1024+1)
	_, appErrTooLarge := svc.UploadImageBytes(ctx, tooLarge, "large.png", "image/png")
	if appErrTooLarge == nil || appErrTooLarge.StatusCode != 400 {
		t.Errorf("expected 400 for file exceeding 5MB, got %v", appErrTooLarge)
	}

	// 4. Valid PNG
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	res, appErrValid := svc.UploadImageBytes(ctx, pngHeader, "test.png", "image/png")
	if appErrValid != nil {
		t.Fatalf("unexpected error uploading valid PNG: %v", appErrValid)
	}
	if res.URL == "" || res.Key == "" {
		t.Errorf("expected non-empty url and key, got url=%s, key=%s", res.URL, res.Key)
	}
}

func TestStorage_UploadImageMultipart(t *testing.T) {
	tempDir := t.TempDir()
	svc := shared.NewStorageService(shared.StorageConfig{
		LocalUploadDir: tempDir,
	})
	ctx := context.Background()

	// 1. Nil header
	_, appErrNil := svc.UploadImage(ctx, nil)
	if appErrNil == nil || appErrNil.StatusCode != 400 {
		t.Errorf("expected 400 for nil fileHeader, got %v", appErrNil)
	}

	// 2. Valid multipart Header
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="sample.jpg"`)
	header.Set("Content-Type", "image/jpeg")
	part, err := mw.CreatePart(header)
	if err != nil {
		t.Fatalf("failed to create part: %v", err)
	}
	// JPEG SOI marker
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
	_, _ = part.Write(jpegBytes)
	mw.Close()

	reader := multipart.NewReader(&buf, mw.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	if err != nil {
		t.Fatalf("failed to read form: %v", err)
	}
	files := form.File["file"]
	if len(files) == 0 {
		t.Fatal("no file in multipart form")
	}

	res, appErr := svc.UploadImage(ctx, files[0])
	if appErr != nil {
		t.Fatalf("unexpected error uploading multipart image: %v", appErr)
	}
	if res.URL == "" || res.Key == "" {
		t.Errorf("expected valid url and key, got url=%s, key=%s", res.URL, res.Key)
	}
}
