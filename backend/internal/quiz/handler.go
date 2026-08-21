package quiz

import (
	"strconv"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Quiz and Uploads domains.
type Handler struct {
	service        Service
	storageService shared.StorageService
}

// NewHandler creates a new Handler instance for the Quiz domain.
func NewHandler(service Service, storageService shared.StorageService) *Handler {
	return &Handler{
		service:        service,
		storageService: storageService,
	}
}

// RegisterRoutes registers all quiz and upload endpoints onto the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	quizzesGroup := rg.Group("/quizzes")
	quizzesGroup.Use(authMiddleware)
	{
		quizzesGroup.GET("", h.ListQuizzes)
		quizzesGroup.POST("", h.CreateQuiz)
		quizzesGroup.GET("/:id", h.GetQuizByID)
		quizzesGroup.PUT("/:id", h.UpdateQuiz)
		quizzesGroup.DELETE("/:id", h.DeleteQuiz)
		quizzesGroup.POST("/:id/duplicate", h.DuplicateQuiz)
	}

	uploadsGroup := rg.Group("/uploads")
	uploadsGroup.Use(authMiddleware)
	{
		uploadsGroup.POST("/images", h.UploadImage)
	}
}

// ListQuizzes handles GET /quizzes
func (h *Handler) ListQuizzes(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "12"))
	search := c.Query("search")
	status := c.Query("status")
	tag := c.Query("tag")

	filter := QuizListFilters{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Status:   status,
		Tag:      tag,
	}

	res, svcErr := h.service.ListQuizzes(c.Request.Context(), actorID, actorRole, filter)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// CreateQuiz handles POST /quizzes
func (h *Handler) CreateQuiz(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateQuizInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.CreateQuiz(c.Request.Context(), actorID, actorRole, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

// GetQuizByID handles GET /quizzes/:id
func (h *Handler) GetQuizByID(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	quizIDStr := c.Param("id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de qüestionari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.GetQuizByID(c.Request.Context(), actorID, actorRole, quizID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// UpdateQuiz handles PUT /quizzes/:id
func (h *Handler) UpdateQuiz(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	quizIDStr := c.Param("id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de qüestionari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req UpdateQuizInput
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El cos de la petició no té un format JSON vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, svcErr := h.service.UpdateQuiz(c.Request.Context(), actorID, actorRole, quizID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// DeleteQuiz handles DELETE /quizzes/:id
func (h *Handler) DeleteQuiz(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	quizIDStr := c.Param("id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de qüestionari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	res, svcErr := h.service.DeleteQuiz(c.Request.Context(), actorID, actorRole, quizID)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// DuplicateQuiz handles POST /quizzes/:id/duplicate
func (h *Handler) DuplicateQuiz(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	quizIDStr := c.Param("id")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'identificador de qüestionari (UUID) invàlid.", map[string]interface{}{"field": "id"}))
		return
	}

	var req DuplicateQuizInput
	// Body is optional for duplicate endpoint
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El format del cos de la petició és invàlid.", map[string]interface{}{"raw_error": err.Error()}))
			return
		}
	}

	res, svcErr := h.service.DuplicateQuiz(c.Request.Context(), actorID, actorRole, quizID, req)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondCreated(c, res)
}

// UploadImage handles POST /uploads/images
func (h *Handler) UploadImage(c *gin.Context) {
	_, _, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "No s'ha rebut cap fitxer 'file' al formulari multipart.", map[string]interface{}{"field": "file", "raw_error": err.Error()}))
		return
	}

	res, uploadErr := h.storageService.UploadImage(c.Request.Context(), fileHeader)
	if uploadErr != nil {
		shared.RespondWithError(c, uploadErr)
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
