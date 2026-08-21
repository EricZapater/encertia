package shared

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxImageSizeBytes = 5 * 1024 * 1024 // 5 MB
)

// AllowedImageMIMEs maps accepted MIME types to file extensions.
var AllowedImageMIMEs = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// UploadResult represents the result of a successful file upload.
type UploadResult struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// StorageConfig holds configuration for Cloudflare R2 or local storage fallback.
type StorageConfig struct {
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string
	LocalUploadDir    string
	BaseURL           string
}

// StorageService defines the contract for image storage.
type StorageService interface {
	UploadImage(ctx context.Context, fileHeader *multipart.FileHeader) (*UploadResult, *AppError)
	UploadImageBytes(ctx context.Context, data []byte, originalFilename, contentType string) (*UploadResult, *AppError)
}

type storageService struct {
	cfg StorageConfig
}

// NewStorageService initializes a new StorageService instance.
func NewStorageService(cfg StorageConfig) StorageService {
	if cfg.LocalUploadDir == "" {
		cfg.LocalUploadDir = "./uploads"
	}
	cfg.R2PublicURL = strings.TrimRight(cfg.R2PublicURL, "/")
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &storageService{cfg: cfg}
}

// isR2Configured checks if all required R2 environment credentials are provided.
func (s *storageService) isR2Configured() bool {
	return s.cfg.R2AccountID != "" &&
		s.cfg.R2AccessKeyID != "" &&
		s.cfg.R2SecretAccessKey != "" &&
		s.cfg.R2BucketName != ""
}

// UploadImage processes and validates a multipart image file.
func (s *storageService) UploadImage(ctx context.Context, fileHeader *multipart.FileHeader) (*UploadResult, *AppError) {
	if fileHeader == nil {
		return nil, ErrBadRequest(ErrCodeValidation, "No s'ha proporcionat cap fitxer per pujar.", map[string]interface{}{"field": "file"})
	}

	if fileHeader.Size > MaxImageSizeBytes {
		return nil, ErrBadRequest(ErrCodeValidation, fmt.Sprintf("La mida del fitxer supera el límit màxim permès de 5MB (mida actual: %d bytes).", fileHeader.Size), map[string]interface{}{"field": "file", "size": fileHeader.Size, "maxSize": MaxImageSizeBytes})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, ErrBadRequest(ErrCodeValidation, "No s'ha pogut obrir el fitxer proporcionat.", map[string]interface{}{"raw_error": err.Error()})
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, ErrInternal(fmt.Errorf("error llegint dades del fitxer: %w", err))
	}

	contentType := fileHeader.Header.Get("Content-Type")
	return s.UploadImageBytes(ctx, data, fileHeader.Filename, contentType)
}

// UploadImageBytes validates image content and stores it in R2 or local storage.
func (s *storageService) UploadImageBytes(ctx context.Context, data []byte, originalFilename, contentType string) (*UploadResult, *AppError) {
	if len(data) == 0 {
		return nil, ErrBadRequest(ErrCodeValidation, "El fitxer està buit.", map[string]interface{}{"field": "file"})
	}

	if len(data) > MaxImageSizeBytes {
		return nil, ErrBadRequest(ErrCodeValidation, fmt.Sprintf("La mida del fitxer supera el límit màxim permès de 5MB (mida actual: %d bytes).", len(data)), map[string]interface{}{"field": "file", "size": len(data), "maxSize": MaxImageSizeBytes})
	}

	// Detect MIME type from file header bytes if needed
	detectedMIME := http.DetectContentType(data)
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedMIME
	}

	ext, allowed := AllowedImageMIMEs[strings.ToLower(contentType)]
	if !allowed {
		// Fallback check by detected MIME
		ext, allowed = AllowedImageMIMEs[strings.ToLower(detectedMIME)]
		if !allowed {
			// Fallback check by extension if MIME detection produced octet-stream
			fileExt := strings.ToLower(filepath.Ext(originalFilename))
			switch fileExt {
			case ".png":
				contentType = "image/png"
				ext = ".png"
				allowed = true
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
				ext = ".jpg"
				allowed = true
			case ".webp":
				contentType = "image/webp"
				ext = ".webp"
				allowed = true
			case ".gif":
				contentType = "image/gif"
				ext = ".gif"
				allowed = true
			}
		}
	}

	if !allowed {
		return nil, ErrBadRequest(ErrCodeValidation, "Format de fitxer no permès. Només s'accepten imatges PNG, JPG, JPEG, WEBP i GIF.", map[string]interface{}{"contentType": contentType, "detectedMIME": detectedMIME})
	}

	now := time.Now().UTC()
	key := fmt.Sprintf("uploads/%04d/%02d/%s%s", now.Year(), int(now.Month()), uuid.New().String(), ext)

	if s.isR2Configured() {
		res, err := s.uploadToR2(ctx, key, data, contentType)
		if err != nil {
			return nil, ErrInternal(fmt.Errorf("error pujant imatge a Cloudflare R2: %w", err))
		}
		return res, nil
	}

	res, err := s.uploadToLocalStorage(key, data)
	if err != nil {
		return nil, ErrInternal(fmt.Errorf("error desant fitxer localment: %w", err))
	}
	return res, nil
}

// uploadToLocalStorage saves the image to the local filesystem.
func (s *storageService) uploadToLocalStorage(key string, data []byte) (*UploadResult, error) {
	fullPath := filepath.Join(s.cfg.LocalUploadDir, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creant directori %s: %w", dir, err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("error escrivint fitxer a %s: %w", fullPath, err)
	}

	var url string
	if s.cfg.R2PublicURL != "" {
		url = fmt.Sprintf("%s/%s", s.cfg.R2PublicURL, key)
	} else if s.cfg.BaseURL != "" {
		url = fmt.Sprintf("%s/%s", s.cfg.BaseURL, key)
	} else {
		url = fmt.Sprintf("/%s", key)
	}

	return &UploadResult{
		URL: url,
		Key: key,
	}, nil
}

// uploadToR2 uploads data to Cloudflare R2 using standard S3 AWS Signature Version 4.
func (s *storageService) uploadToR2(ctx context.Context, key string, data []byte, contentType string) (*UploadResult, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", s.cfg.R2AccountID, s.cfg.R2BucketName, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creant petició HTTP: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	region := "auto"
	service := "s3"

	payloadHash := sha256Hex(data)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	host := fmt.Sprintf("%s.r2.cloudflarestorage.com", s.cfg.R2AccountID)
	req.Host = host

	canonicalURI := fmt.Sprintf("/%s/%s", s.cfg.R2BucketName, key)
	canonicalQuery := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		contentType, host, payloadHash, amzDate)
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := getSignatureKey(s.cfg.R2SecretAccessKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.cfg.R2AccessKeyID, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executant petició HTTP a R2: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("r2 va respondre amb codi %d: %s", resp.StatusCode, string(body))
	}

	publicURL := s.cfg.R2PublicURL
	if publicURL == "" {
		publicURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", s.cfg.R2AccountID, s.cfg.R2BucketName)
	}
	finalURL := fmt.Sprintf("%s/%s", publicURL, key)

	return &UploadResult{
		URL: finalURL,
		Key: key,
	}, nil
}

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(secret, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
