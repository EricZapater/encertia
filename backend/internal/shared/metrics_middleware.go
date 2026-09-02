package shared

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogInput conté les dades d'una petició HTTP per ser registrades a l'auditoria.
type AuditLogInput struct {
	UserID     *string
	UserEmail  *string
	UserRole   *string
	Action     string
	Module     string
	Endpoint   string
	Method     string
	StatusCode int
	DurationMs int
	IPAddress  *string
	UserAgent  *string
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

// MetricsMiddleware mesura el temps de resposta duration_ms de cada petició HTTP
// i registra asíncronament l'audit log a PostgreSQL via el MetricsRecorder passat.
func MetricsMiddleware(recorder MetricsRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		durationMs := int(time.Since(start).Milliseconds())

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
				UserID:     userIDPtr,
				UserEmail:  userEmailPtr,
				UserRole:   userRolePtr,
				Action:     action,
				Module:     module,
				Endpoint:   path,
				Method:     method,
				StatusCode: cCopy.Writer.Status(),
				DurationMs: durationMs,
				IPAddress:  ipPtr,
				UserAgent:  uaPtr,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = recorder.RecordAuditLog(ctx, entry)
		}()
	}
}
