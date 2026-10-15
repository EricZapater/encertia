package shared

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogInput conté les dades d'una petició HTTP per ser registrades a l'auditoria.
type AuditLogInput struct {
	UserID       *string
	UserEmail    *string
	UserRole     *string
	Action       string
	Module       string
	Endpoint     string
	Method       string
	StatusCode   int
	DurationMs   int
	Payload      *string
	ResponseBody *string
	IPAddress    *string
	UserAgent    *string
}

// sensitivePattern detecta camps sensibles en formats JSON i form-data per emmascarar-los.
var (
	sensitiveJsonPattern = regexp.MustCompile(`(?i)"(password|currentPassword|newPassword|confirmPassword|token|accessToken|refreshToken|secret|deviceToken)"\s*:\s*"[^"]*"`)
	sensitiveFormPattern = regexp.MustCompile(`(?i)(password|currentPassword|newPassword|token|secret|deviceToken)=[^&]*`)
)

func sanitizeBody(content string) string {
	if content == "" {
		return ""
	}
	sanitized := sensitiveJsonPattern.ReplaceAllString(content, `"$1":"[REDACTED]"`)
	sanitized = sensitiveFormPattern.ReplaceAllString(sanitized, `$1=[REDACTED]`)
	return sanitized
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	max  int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.body.Len() < w.max {
		remaining := w.max - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if w.body != nil && w.body.Len() < w.max {
		remaining := w.max - w.body.Len()
		if len(s) > remaining {
			w.body.WriteString(s[:remaining])
		} else {
			w.body.WriteString(s)
		}
	}
	return w.ResponseWriter.WriteString(s)
}

// MetricsRecorder defineix el contracte per enregistrar logs d'auditoria.
type MetricsRecorder interface {
	RecordAuditLog(ctx context.Context, entry AuditLogInput) error
}

// MetricsRecorderFunc permet utilitzar una funció com a MetricsRecorder.
type MetricsRecorderFunc func(ctx context.Context, entry AuditLogInput) error

func (f MetricsRecorderFunc) RecordAuditLog(ctx context.Context, entry AuditLogInput) error {
	return f(ctx, entry)
}

// MetricsMiddleware mesura el temps de resposta duration_ms de cada petició HTTP,
// captura el payload i response_body (sanititzant dades sensibles) i registra asíncronament l'audit log.
func MetricsMiddleware(recorder MetricsRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Captura del payload de la petició (abans de c.Next())
		var payloadPtr *string
		if c.Request.Body != nil {
			contentType := c.ContentType()
			if strings.HasPrefix(contentType, "multipart/form-data") {
				msg := "[Multipart Form Data]"
				payloadPtr = &msg
			} else {
				// Llegim fins a 32 KB per no saturar memòria
				bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, 32768))
				if err == nil {
					// Restaurem el body perquè els handlers subsegüents el puguin consumir
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					if len(bodyBytes) > 0 {
						sanitized := sanitizeBody(string(bodyBytes))
						payloadPtr = &sanitized
					}
				}
			}
		}

		if payloadPtr == nil && c.Request.URL.RawQuery != "" {
			query := "?" + c.Request.URL.RawQuery
			payloadPtr = &query
		}

		// 2. Intercepció del cos de resposta
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
			max:            32768,
		}
		c.Writer = blw

		c.Next()

		durationMs := int(time.Since(start).Milliseconds())

		// 3. Captura del cos de resposta
		var responseBodyPtr *string
		respContentType := c.Writer.Header().Get("Content-Type")
		isBinary := strings.HasPrefix(respContentType, "application/octet-stream") ||
			strings.HasPrefix(respContentType, "application/zip") ||
			strings.HasPrefix(respContentType, "image/") ||
			strings.HasPrefix(respContentType, "audio/") ||
			strings.HasPrefix(respContentType, "video/") ||
			strings.HasPrefix(respContentType, "application/pdf")

		if isBinary {
			msg := "[Binary Content]"
			responseBodyPtr = &msg
		} else if blw.body != nil && blw.body.Len() > 0 {
			sanitized := sanitizeBody(blw.body.String())
			responseBodyPtr = &sanitized
		}

		// Fem una còpia del context de Gin per llegir dades de forma segura dins la goroutine
		cCopy := c.Copy()

		go func() {
			path := cCopy.FullPath()
			if path == "" {
				path = cCopy.Request.URL.Path
			}

			// Infecció del mòdul a partir de la ruta
			cleanPath := strings.TrimPrefix(path, "/api/v1")
			cleanPath = strings.TrimPrefix(cleanPath, "/api")
			parts := strings.Split(strings.Trim(cleanPath, "/"), "/")

			module := "system"
			if len(parts) > 0 && parts[0] != "" {
				module = parts[0]
			}

			method := cCopy.Request.Method
			action := method + " " + path

			var userIDPtr *string
			if val, exists := cCopy.Get(CtxKeyUserID); exists {
				if uid, ok := val.(string); ok && uid != "" {
					userIDPtr = &uid
				}
			}

			var userEmailPtr *string
			if val, exists := cCopy.Get(CtxKeyUserEmail); exists {
				if email, ok := val.(string); ok && email != "" {
					userEmailPtr = &email
				}
			}

			var userRolePtr *string
			if val, exists := cCopy.Get(CtxKeyUserRole); exists {
				if role, ok := val.(string); ok && role != "" {
					userRolePtr = &role
				}
			}

			ip := cCopy.ClientIP()
			var ipPtr *string
			if ip != "" {
				ipPtr = &ip
			}

			ua := cCopy.Request.UserAgent()
			var uaPtr *string
			if ua != "" {
				uaPtr = &ua
			}

			entry := AuditLogInput{
				UserID:       userIDPtr,
				UserEmail:    userEmailPtr,
				UserRole:     userRolePtr,
				Action:       action,
				Module:       module,
				Endpoint:     path,
				Method:       method,
				StatusCode:   cCopy.Writer.Status(),
				DurationMs:   durationMs,
				Payload:      payloadPtr,
				ResponseBody: responseBodyPtr,
				IPAddress:    ipPtr,
				UserAgent:    uaPtr,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = recorder.RecordAuditLog(ctx, entry)
		}()
	}
}
