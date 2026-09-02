package metrics

import (
	"context"
	"net/http"
	"strconv"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RecorderAdapter retorna un MetricsRecorder apte per passar a shared.MetricsMiddleware.
func (h *Handler) RecorderAdapter() shared.MetricsRecorder {
	return shared.MetricsRecorderFunc(func(ctx context.Context, entry shared.AuditLogInput) error {
		return h.service.RecordAuditLog(ctx, entry)
	})
}

// RegisterRoutes registra els endpoints protegits per admin.
func (h *Handler) RegisterRoutes(router gin.IRouter, authMiddleware gin.HandlerFunc) {
	metricsGroup := router.Group("/metrics", authMiddleware, shared.RequireRole("admin"))
	{
		metricsGroup.GET("/summary", h.GetSummary)
		metricsGroup.GET("/api-latency", h.GetApiLatency)
		metricsGroup.GET("/audit-logs", h.GetAuditLogs)
		metricsGroup.GET("/audit-logs/export", h.ExportAuditLogsCSV)
	}
}

// GetSummary (GET /metrics/summary)
func (h *Handler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary(c.Request.Context())
	if err != nil {
		shared.RespondWithError(c, err)
		return
	}
	shared.RespondOK(c, summary)
}

// GetApiLatency (GET /metrics/api-latency)
func (h *Handler) GetApiLatency(c *gin.Context) {
	latency, err := h.service.GetApiLatency(c.Request.Context())
	if err != nil {
		shared.RespondWithError(c, err)
		return
	}
	shared.RespondOK(c, latency)
}

// GetAuditLogs (GET /metrics/audit-logs)
func (h *Handler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")
	module := c.Query("module")
	userID := c.Query("userId")

	filter := AuditLogFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Module:   module,
		UserID:   userID,
	}

	result, err := h.service.GetAuditLogs(c.Request.Context(), filter)
	if err != nil {
		shared.RespondWithError(c, err)
		return
	}
	shared.RespondOK(c, result)
}

// ExportAuditLogsCSV (GET /metrics/audit-logs/export)
func (h *Handler) ExportAuditLogsCSV(c *gin.Context) {
	csvBytes, err := h.service.ExportAuditLogsCSV(c.Request.Context())
	if err != nil {
		shared.RespondWithError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="audit_logs.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}
