package course

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Course domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all course routes onto the router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	coursesGroup := rg.Group("/courses")
	coursesGroup.Use(authMiddleware)
	{
		coursesGroup.GET("", h.ListCourses)
		coursesGroup.POST("", shared.RequireRole("teacher", "admin"), h.CreateCourse)
		coursesGroup.GET("/:id", h.GetCourseByID)
		coursesGroup.PUT("/:id", shared.RequireRole("teacher", "admin"), h.UpdateCourse)
		coursesGroup.DELETE("/:id", shared.RequireRole("teacher", "admin"), h.DeleteCourse)

		coursesGroup.GET("/:id/students", shared.RequireRole("teacher", "admin"), h.GetCourseStudents)
		coursesGroup.POST("/:id/students", shared.RequireRole("teacher", "admin"), h.EnrollStudents)
		coursesGroup.DELETE("/:id/students/:studentId", shared.RequireRole("teacher", "admin"), h.UnenrollStudent)

		coursesGroup.GET("/:id/units", h.ListCourseUnits)
		coursesGroup.POST("/:id/units", shared.RequireRole("teacher", "admin"), h.CreateCourseUnit)
		coursesGroup.PUT("/:id/units/reorder", shared.RequireRole("teacher", "admin"), h.ReorderCourseUnits)
		coursesGroup.GET("/:id/units/:unitId", h.GetCourseUnit)
		coursesGroup.PUT("/:id/units/:unitId", shared.RequireRole("teacher", "admin"), h.UpdateCourseUnit)
		coursesGroup.DELETE("/:id/units/:unitId", shared.RequireRole("teacher", "admin"), h.DeleteCourseUnit)

		coursesGroup.POST("/:id/units/:unitId/quizzes", shared.RequireRole("teacher", "admin"), h.LinkQuizToUnit)
		coursesGroup.DELETE("/:id/units/:unitId/quizzes/:quizId", shared.RequireRole("teacher", "admin"), h.UnlinkQuizFromUnit)

		coursesGroup.GET("/:id/units/:unitId/script", h.GetUnitScript)
		coursesGroup.PUT("/:id/units/:unitId/script", shared.RequireRole("teacher", "admin"), h.UpdateUnitScript)
	}
}

func (h *Handler) ListCourses(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")
	status := c.Query("status")

	filters := CourseListFilters{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Status:   status,
	}

	res, svcErr := h.service.ListCourses(c.Request.Context(), actorID, actorRole, filters)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) CreateCourse(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.CreateCourse(c.Request.Context(), actorID, actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

func (h *Handler) GetCourseByID(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.GetCourseByID(c.Request.Context(), actorID, actorRole, courseID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) UpdateCourse(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateCourse(c.Request.Context(), actorID, actorRole, courseID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) DeleteCourse(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	if svcErr := h.service.DeleteCourse(c.Request.Context(), actorID, actorRole, courseID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetCourseStudents(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.GetCourseStudents(c.Request.Context(), actorID, actorRole, courseID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) EnrollStudents(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req EnrollStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.EnrollStudents(c.Request.Context(), actorID, actorRole, courseID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) UnenrollStudent(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	studentID, err := uuid.Parse(c.Param("studentId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador d'alumne (UUID) invàlid.", map[string]interface{}{"field": "studentId"}))
		return
	}

	if svcErr := h.service.UnenrollStudent(c.Request.Context(), actorID, actorRole, courseID, studentID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListCourseUnits(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.ListCourseUnits(c.Request.Context(), actorID, actorRole, courseID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) CreateCourseUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req CreateCourseUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.CreateCourseUnit(c.Request.Context(), actorID, actorRole, courseID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

func (h *Handler) ReorderCourseUnits(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var unitIDs []uuid.UUID
	if err := c.ShouldBindJSON(&unitIDs); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format d'array d'UUIDs vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.ReorderCourseUnits(c.Request.Context(), actorID, actorRole, courseID, unitIDs)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) GetCourseUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	res, svcErr := h.service.GetCourseUnit(c.Request.Context(), actorID, actorRole, courseID, unitID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) UpdateCourseUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	var req UpdateCourseUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateCourseUnit(c.Request.Context(), actorID, actorRole, courseID, unitID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) DeleteCourseUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	if svcErr := h.service.DeleteCourseUnit(c.Request.Context(), actorID, actorRole, courseID, unitID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) LinkQuizToUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	var req LinkQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	if svcErr := h.service.LinkQuizToUnit(c.Request.Context(), actorID, actorRole, courseID, unitID, req.QuizID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, gin.H{"message": "Qüestionari vinculat a la unitat."})
}

func (h *Handler) UnlinkQuizFromUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	quizID, err := uuid.Parse(c.Param("quizId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de qüestionari (UUID) invàlid.", map[string]interface{}{"field": "quizId"}))
		return
	}

	if svcErr := h.service.UnlinkQuizFromUnit(c.Request.Context(), actorID, actorRole, courseID, unitID, quizID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetUnitScript(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	res, svcErr := h.service.GetUnitScript(c.Request.Context(), actorID, actorRole, courseID, unitID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

func (h *Handler) UpdateUnitScript(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de curs (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de unitat (UUID) invàlid.", map[string]interface{}{"field": "unitId"}))
		return
	}

	var req []CreateScriptBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format d'array JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateUnitScript(c.Request.Context(), actorID, actorRole, courseID, unitID, req)
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
