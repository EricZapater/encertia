package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/encertia/backend/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserPermissionChecker defineix la interfície per verificar permisos d'usuari.
type UserPermissionChecker interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.UserDB, error)
}

// Handler gestiona les peticions HTTP relacionades amb el descarregador de vídeos.
type Handler struct {
	service  Service
	userRepo UserPermissionChecker
}

// NewHandler crea una nova instància del Handler.
func NewHandler(service Service, userRepo UserPermissionChecker) *Handler {
	return &Handler{
		service:  service,
		userRepo: userRepo,
	}
}

// RegisterRoutes registra les rutes del mòdul de descàrrega de vídeos.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	dlGroup := rg.Group("/downloader")
	dlGroup.Use(authMiddleware, h.requireVideoDownloaderAccess())
	{
		dlGroup.POST("/info", h.GetVideoInfo)
		dlGroup.POST("/download", h.DownloadVideo)
	}
}

// requireVideoDownloaderAccess comprova que l'usuari sigui admin o tingui permís explícit can_download_videos.
func (h *Handler) requireVideoDownloaderAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(shared.CtxKeyUserRole)
		if !exists {
			shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Usuari no autenticat."))
			c.Abort()
			return
		}

		role, _ := roleVal.(string)
		// Els administradors tenen sempre accés
		if role == "admin" {
			c.Next()
			return
		}

		// Per a altres rols (professor, alumne), comprovem el permís can_download_videos
		userIDVal, exists := c.Get(shared.CtxKeyUserID)
		if !exists {
			shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Usuari no autenticat."))
			c.Abort()
			return
		}

		userIDStr, _ := userIDVal.(string)
		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			shared.RespondWithError(c, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari no vàlid."))
			c.Abort()
			return
		}

		if h.userRepo == nil {
			shared.RespondWithError(c, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per utilitzar el descarregador de vídeos."))
			c.Abort()
			return
		}

		u, err := h.userRepo.GetUserByID(c.Request.Context(), userUUID)
		if err != nil || u == nil || !u.CanDownloadVideos {
			shared.RespondWithError(c, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per utilitzar el descarregador de vídeos."))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetVideoInfo obté les metadades d'un vídeo donada la seva URL.
// POST /downloader/info
func (h *Handler) GetVideoInfo(c *gin.Context) {
	var req VideoInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Cal indicar l'adreça URL del vídeo.", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	info, err := h.service.GetVideoInfo(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, info)
}

// DownloadVideo processa la descàrrega del vídeo i retorna el fitxer MP4 com a stream/fitxer adjunt.
// POST /downloader/download
func (h *Handler) DownloadVideo(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Cal indicar l'adreça URL del vídeo a descarregar.", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	filePath, filename, cleanup, err := h.service.DownloadVideo(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	defer cleanup()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "video/mp4")
	c.File(filePath)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, ErrInvalidURL) {
		shared.RespondWithError(c, shared.ErrBadRequest("INVALID_URL", err.Error(), nil))
		return
	}
	if errors.Is(err, ErrDisallowedHost) {
		shared.RespondWithError(c, shared.ErrForbidden("DISALLOWED_HOST", err.Error()))
		return
	}
	if errors.Is(err, ErrToolNotInstalled) {
		shared.RespondWithError(c, shared.NewAppError(
			http.StatusServiceUnavailable,
			"DOWNLOADER_UNAVAILABLE",
			"El motor de descàrrega no està disponible temporalment.",
			nil,
			err,
		))
		return
	}
	if errors.Is(err, ErrDownloadFailed) {
		errMsg := err.Error()
		friendlyMsg := errMsg
		if strings.Contains(errMsg, "Sign in to confirm you’re not a bot") || strings.Contains(errMsg, "confirm you're not a bot") {
			friendlyMsg = "YouTube demana confirmació de bot. Si us plau, enganxa les teves cookies de YouTube a 'Opcions avançades'."
		} else if strings.Contains(errMsg, "login") || strings.Contains(errMsg, "private") || strings.Contains(errMsg, "account") {
			friendlyMsg = "Aquest vídeo requereix iniciar sessió a la plataforma (ex: publicació privada o LinkedIn). Pots indicar les teves cookies a 'Opcions avançades'."
		}
		shared.RespondWithError(c, shared.ErrBadRequest("DOWNLOAD_FAILED", friendlyMsg, map[string]interface{}{
			"raw": errMsg,
		}))
		return
	}

	shared.RespondWithError(c, shared.ErrInternal(err))
}
