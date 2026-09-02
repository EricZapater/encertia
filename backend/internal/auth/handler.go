package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Auth domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all auth endpoints onto the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.Refresh)
		authGroup.POST("/logout", authMiddleware, h.Logout)
		authGroup.GET("/me", authMiddleware, h.Me)
	}
}

// Register handles POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, appErr := h.service.Register(c.Request.Context(), req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	shared.RespondWithJSON(c, http.StatusCreated, res)
}

// Login handles POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, appErr := h.service.Login(c.Request.Context(), req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	shared.RespondOK(c, res)
}

// Refresh handles POST /auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, appErr := h.service.RefreshToken(c.Request.Context(), req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	shared.RespondOK(c, res)
}

// Logout handles POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	userIDStr, exists := c.Get(shared.CtxKeyUserID)
	if !exists {
		shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "No autoritzat."))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid al token."))
		return
	}

	accessToken := ""
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		accessToken = strings.TrimSpace(parts[1])
	}

	var req LogoutRequest
	// Logout request body is optional
	_ = c.ShouldBindJSON(&req)

	res, appErr := h.service.Logout(c.Request.Context(), userID, accessToken, req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	shared.RespondOK(c, res)
}

// Me handles GET /auth/me
func (h *Handler) Me(c *gin.Context) {
	userIDStr, exists := c.Get(shared.CtxKeyUserID)
	if !exists {
		shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "No autoritzat."))
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid al token."))
		return
	}

	res, appErr := h.service.GetCurrentUser(c.Request.Context(), userID)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	shared.RespondOK(c, res)
}

// TokenValidatorAdapter adapts auth.Service to shared.TokenValidator interface
func (h *Handler) TokenValidatorAdapter() shared.TokenValidator {
	return shared.TokenValidatorFunc(func(ctx context.Context, tokenString string) (string, string, string, *shared.AppError) {
		claims, err := h.service.ValidateAccessToken(ctx, tokenString)
		if err != nil {
			return "", "", "", err
		}
		if claims.UserID == "" {
			return "", "", "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Token sense identificador d'usuari.")
		}
		return claims.UserID, claims.Email, string(claims.Role), nil
	})
}
