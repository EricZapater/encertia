package material

import (
	"net/http"
	"strconv"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Material domain.
type Handler struct {
	service        Service
	storageService shared.StorageService
}

// NewHandler creates a new Handler instance.
func NewHandler(service Service, storageService shared.StorageService) *Handler {
	return &Handler{
		service:        service,
		storageService: storageService,
	}
}

// RegisterRoutes registers all material endpoints onto the router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	materialsGroup := rg.Group("/materials")
	materialsGroup.Use(authMiddleware)
	{
		materialsGroup.GET("", h.ListMaterials)
		materialsGroup.POST("", shared.RequireRole("teacher", "admin"), h.CreateMaterial)
		materialsGroup.POST("/upload", shared.RequireRole("teacher", "admin"), h.UploadMaterialFile)
		materialsGroup.GET("/:id", h.GetMaterial)
		materialsGroup.PUT("/:id", shared.RequireRole("teacher", "admin"), h.UpdateMaterial)
		materialsGroup.DELETE("/:id", shared.RequireRole("teacher", "admin"), h.DeleteMaterial)

		materialsGroup.POST("/:id/views", h.RecordMaterialView)
		materialsGroup.GET("/:id/views", shared.RequireRole("teacher", "admin"), h.GetMaterialViewsReport)
	}

	coursesGroup := rg.Group("/courses")
	coursesGroup.Use(authMiddleware)
	{
		coursesGroup.GET("/:id/units/:unitId/materials", h.ListUnitMaterials)
		coursesGroup.POST("/:id/units/:unitId/materials", shared.RequireRole("teacher", "admin"), h.LinkMaterialToUnit)
		coursesGroup.DELETE("/:id/units/:unitId/materials/:materialId", shared.RequireRole("teacher", "admin"), h.UnlinkMaterialFromUnit)
	}
}

func getActorFromContext(c *gin.Context) (uuid.UUID, string, *shared.AppError) {
	userIDVal, exists := c.Get(shared.CtxKeyUserID)
	if !exists {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Usuari no autenticat.")
	}

	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "ID d'usuari invàlid.")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Format d'ID d'usuari invàlid.")
	}

	roleVal, _ := c.Get(shared.CtxKeyUserRole)
	roleStr, _ := roleVal.(string)

	return userID, roleStr, nil
}

func (h *Handler) ListMaterials(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")
	materialType := c.Query("materialType")

	filters := MaterialListFilters{
		Page:         page,
		PageSize:     pageSize,
		Search:       search,
		MaterialType: materialType,
	}

	resp, appErr := h.service.ListMaterials(c.Request.Context(), actorID, actorRole, filters)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateMaterial(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format de petició JSON invàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	resp, appErr := h.service.CreateMaterial(c.Request.Context(), actorID, actorRole, req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) UploadMaterialFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Manca el fitxer multipart 'file'.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, appErr := h.storageService.UploadDocument(c.Request.Context(), fileHeader)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, UploadFileResponse{
		FileURL:       res.FileURL,
		FileName:      res.FileName,
		FileSizeBytes: res.FileSizeBytes,
		MIMEType:      res.MIMEType,
		PageCount:     res.PageCount,
	})
}

func (h *Handler) GetMaterial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	resp, appErr := h.service.GetMaterialByID(c.Request.Context(), id)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateMaterial(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	var req UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format de petició JSON invàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	resp, appErr := h.service.UpdateMaterial(c.Request.Context(), actorID, actorRole, id, req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteMaterial(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	appErr = h.service.DeleteMaterial(c.Request.Context(), actorID, actorRole, id)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) RecordMaterialView(c *gin.Context) {
	actorID, _, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	res, appErr := h.service.RecordMaterialView(c.Request.Context(), actorID, id)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) GetMaterialViewsReport(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	resp, appErr := h.service.GetMaterialViewsReport(c.Request.Context(), actorID, actorRole, id)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListUnitMaterials(c *gin.Context) {
	unitIDStr := c.Param("unitId")
	unitID, err := uuid.Parse(unitIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de unitat invàlid.", map[string]interface{}{"param": "unitId"}))
		return
	}

	resp, appErr := h.service.ListUnitMaterials(c.Request.Context(), unitID)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) LinkMaterialToUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de curs invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	unitIDStr := c.Param("unitId")
	unitID, err := uuid.Parse(unitIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de unitat invàlid.", map[string]interface{}{"param": "unitId"}))
		return
	}

	var req LinkMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format de petició JSON invàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	appErr = h.service.LinkMaterialToUnit(c.Request.Context(), actorID, actorRole, courseID, unitID, req)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UnlinkMaterialFromUnit(c *gin.Context) {
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de curs invàlid.", map[string]interface{}{"param": "id"}))
		return
	}

	unitIDStr := c.Param("unitId")
	unitID, err := uuid.Parse(unitIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de unitat invàlid.", map[string]interface{}{"param": "unitId"}))
		return
	}

	materialIDStr := c.Param("materialId")
	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "Format d'ID de material invàlid.", map[string]interface{}{"param": "materialId"}))
		return
	}

	appErr = h.service.UnlinkMaterialFromUnit(c.Request.Context(), actorID, actorRole, courseID, unitID, materialID)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	c.Status(http.StatusNoContent)
}
