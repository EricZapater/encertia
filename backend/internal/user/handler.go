package user

import (
	"strconv"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the User domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance for the User domain.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all user domain endpoints onto the router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	usersGroup := rg.Group("/users")
	usersGroup.Use(authMiddleware)
	{
		usersGroup.GET("", shared.RequireRole(string(RoleAdmin), string(RoleTeacher)), h.ListUsers)
		usersGroup.POST("", shared.RequireRole(string(RoleAdmin), string(RoleTeacher)), h.CreateUser)
		usersGroup.POST("/batch", shared.RequireRole(string(RoleAdmin), string(RoleTeacher)), h.BatchCreateUsers)
		usersGroup.GET("/:id", h.GetUserByID)
		usersGroup.PUT("/:id", h.UpdateUser)
		usersGroup.POST("/:id/password", shared.RequireRole(string(RoleAdmin), string(RoleTeacher)), h.ResetPassword)
		usersGroup.DELETE("/:id", shared.RequireRole(string(RoleAdmin)), h.DeleteUser)
	}
}

// ListUsers handles GET /users
func (h *Handler) ListUsers(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.DefaultQuery("status", "active")

	filter := ListUsersFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Role:     role,
		Status:   status,
	}

	res, svcErr := h.service.ListUsers(c.Request.Context(), actorID, actorRole, filter)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.CreateUser(c.Request.Context(), actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

// BatchCreateUsers handles POST /users/batch
func (h *Handler) BatchCreateUsers(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req BatchCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El format del cos de la petició és invàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.BatchCreateUsers(c.Request.Context(), actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// GetUserByID handles GET /users/:id
func (h *Handler) GetUserByID(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'usuari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.GetUserByID(c.Request.Context(), actorID, actorRole, targetID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'usuari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req UpdateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El format de les dades no és vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateUser(c.Request.Context(), actorID, actorRole, targetID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// ResetPassword handles POST /users/:id/password
func (h *Handler) ResetPassword(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'usuari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req ResetPasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El format de la contrasenya no és vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.ResetPassword(c.Request.Context(), actorID, actorRole, targetID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'usuari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.DeleteUser(c.Request.Context(), actorID, actorRole, targetID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
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
