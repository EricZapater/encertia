package device

import (
	"net/http"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Device domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance for devices.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all device routes onto the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	rg.POST("/devices/verify", h.VerifyDevice)

	adminDevices := rg.Group("/admin/devices")
	adminDevices.Use(authMiddleware, shared.RequireRole("admin"))
	{
		adminDevices.GET("", h.ListDevices)
		adminDevices.POST("", h.RegisterDevice)
		adminDevices.DELETE("/:id", h.RevokeDevice)
	}
}

// VerifyDevice handles POST /devices/verify
func (h *Handler) VerifyDevice(c *gin.Context) {
	var req VerifyDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El cos de la petició no té un format JSON vàlid.",
			map[string]interface{}{"raw_error": err.Error()},
		))
		return
	}

	res, svcErr := h.service.VerifyDevice(c.Request.Context(), req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// ListDevices handles GET /admin/devices
func (h *Handler) ListDevices(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	res, svcErr := h.service.ListDevices(c.Request.Context(), actorRole)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// RegisterDevice handles POST /admin/devices
func (h *Handler) RegisterDevice(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El cos de la petició no té un format JSON vàlid.",
			map[string]interface{}{"raw_error": err.Error()},
		))
		return
	}

	res, svcErr := h.service.RegisterDevice(c.Request.Context(), actorID, actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

// RevokeDevice handles DELETE /admin/devices/:id
func (h *Handler) RevokeDevice(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"Format d'identificador de dispositiu (UUID) invàlid.",
			map[string]interface{}{"field": "id"},
		))
		return
	}

	if svcErr := h.service.RevokeDevice(c.Request.Context(), actorRole, deviceID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func getActorFromContext(c *gin.Context) (uuid.UUID, string, *shared.AppError) {
	userIDVal, exists := c.Get(shared.CtxKeyUserID)
	if !exists {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "No autenticat.")
	}

	userIDStr, ok := userIDVal.(string)
	if !ok || strings.TrimSpace(userIDStr) == "" {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid.")
	}

	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid.")
	}

	userRoleVal, exists := c.Get(shared.CtxKeyUserRole)
	if !exists {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Rol d'usuari no especificat al token.")
	}

	actorRole, ok := userRoleVal.(string)
	if !ok || strings.TrimSpace(actorRole) == "" {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Rol d'usuari invàlid.")
	}

	return actorID, actorRole, nil
}
