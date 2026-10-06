package group

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Group domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance for groups.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all group routes onto the router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	groupsGroup := rg.Group("/groups")
	groupsGroup.Use(authMiddleware)
	{
		groupsGroup.GET("", h.ListGroups)
		groupsGroup.POST("", shared.RequireRole("teacher", "admin"), h.CreateGroup)
		groupsGroup.GET("/:id", h.GetGroupByID)
		groupsGroup.PUT("/:id", shared.RequireRole("teacher", "admin"), h.UpdateGroup)
		groupsGroup.DELETE("/:id", shared.RequireRole("teacher", "admin"), h.DeleteGroup)

		groupsGroup.GET("/:id/students", h.ListGroupStudents)
		groupsGroup.POST("/:id/students", shared.RequireRole("teacher", "admin"), h.AddStudentsToGroup)
		groupsGroup.DELETE("/:id/students/:studentId", shared.RequireRole("teacher", "admin"), h.RemoveStudentFromGroup)
	}
}

func (h *Handler) ListGroups(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")
	academicYear := c.Query("academicYear")
	courseIDStr := c.Query("courseId")

	var courseID *uuid.UUID
	if strings.TrimSpace(courseIDStr) != "" {
		parsed, err := uuid.Parse(courseIDStr)
		if err != nil {
			shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "courseId"}))
			return
		}
		courseID = &parsed
	}

	filters := GroupListFilters{
		Page:         page,
		PageSize:     pageSize,
		Search:       search,
		AcademicYear: academicYear,
		CourseID:     courseID,
	}

	res, svcErr := h.service.ListGroups(c.Request.Context(), actorID, actorRole, filters)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) CreateGroup(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.CreateGroup(c.Request.Context(), actorID, actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

func (h *Handler) GetGroupByID(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.GetGroupByID(c.Request.Context(), actorID, actorRole, groupID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateGroup(c.Request.Context(), actorID, actorRole, groupID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	if svcErr := h.service.DeleteGroup(c.Request.Context(), actorID, actorRole, groupID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListGroupStudents(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.ListGroupStudents(c.Request.Context(), actorID, actorRole, groupID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) AddStudentsToGroup(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req AddStudentsToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.AddStudentsToGroup(c.Request.Context(), actorID, actorRole, groupID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) RemoveStudentFromGroup(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de grup (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	studentID, err := uuid.Parse(c.Param("studentId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'alumne (UUID) invàlid.", map[string]interface{}{"field": "studentId"}))
		return
	}

	if svcErr := h.service.RemoveStudentFromGroup(c.Request.Context(), actorID, actorRole, groupID, studentID); svcErr != nil {
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
