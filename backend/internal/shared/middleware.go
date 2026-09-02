package shared

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxKeyUserID    = "auth_user_id"
	CtxKeyUserEmail = "auth_user_email"
	CtxKeyUserRole  = "auth_user_role"
)

// TokenValidator defines the contract for validating an access token.
type TokenValidator interface {
	ValidateAccessToken(ctx context.Context, tokenString string) (userID string, email string, role string, appErr *AppError)
}

// TokenValidatorFunc allows using a function as TokenValidator.
type TokenValidatorFunc func(ctx context.Context, tokenString string) (userID string, email string, role string, appErr *AppError)

func (f TokenValidatorFunc) ValidateAccessToken(ctx context.Context, tokenString string) (string, string, string, *AppError) {
	return f(ctx, tokenString)
}

// AuthMiddleware creates a Gin middleware that requires a valid Bearer JWT.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			RespondWithError(c, ErrUnauthorized(ErrCodeUnauthorized, "Manca la capçalera d'autorització."))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			RespondWithError(c, ErrUnauthorized(ErrCodeInvalidToken, "Format de capçalera Authorization invàlid. S'espera 'Bearer <token>'."))
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			RespondWithError(c, ErrUnauthorized(ErrCodeInvalidToken, "Token no proporcionat."))
			c.Abort()
			return
		}

		userID, email, role, err := validator.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			RespondWithError(c, err)
			c.Abort()
			return
		}

		c.Set(CtxKeyUserID, userID)
		c.Set(CtxKeyUserEmail, email)
		c.Set(CtxKeyUserRole, role)

		c.Next()
	}
}

// RequireRole checks that the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}
	return func(c *gin.Context) {
		userRole, exists := c.Get(CtxKeyUserRole)
		if !exists {
			RespondWithError(c, ErrUnauthorized(ErrCodeUnauthorized, "No autenticat."))
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok || !roleSet[roleStr] {
			RespondWithError(c, ErrForbidden(ErrCodeForbidden, "No tens permisos suficients per realitzar aquesta acció."))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware provides CORS headers.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
