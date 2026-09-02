package metrics

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"math"
	"runtime"
	"strconv"
	"time"

	"github.com/encertia/backend/internal/shared"
)

type Service interface {
	RecordAuditLog(ctx context.Context, entry shared.AuditLogInput) error
	GetSummary(ctx context.Context) (*MetricsSummary, *shared.AppError)
	GetApiLatency(ctx context.Context) (*ApiLatencyMetricsResponse, *shared.AppError)
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) (*AuditLogsPaginatedResponse, *shared.AppError)
	ExportAuditLogsCSV(ctx context.Context) ([]byte, *shared.AppError)
}

type service struct {
	repo      Repository
	db        *sql.DB
	startTime time.Time
}

func NewService(repo Repository, db *sql.DB) Service {
	return &service{
		repo:      repo,
		db:        db,
		startTime: time.Now(),
	}
}

func (s *service) RecordAuditLog(ctx context.Context, entry shared.AuditLogInput) error {
	return s.repo.InsertAuditLog(ctx, entry)
}

func (s *service) GetSummary(ctx context.Context) (*MetricsSummary, *shared.AppError) {
	activeUsers, totalReq, avgRespMs, errRate, matchesPlayed, materialsRead, err := s.repo.GetMetricsSummaryData(ctx)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	// System Health calculation
	uptimeSecs := int64(time.Since(s.startTime).Seconds())
	numGoroutines := runtime.NumGoroutine()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := float64(m.Alloc) / (1024.0 * 1024.0)

	openConns := 0
	inUseConns := 0
	if s.db != nil {
		stats := s.db.Stats()
		openConns = stats.OpenConnections
		inUseConns = stats.InUse
	}

	summary := &MetricsSummary{
		ActiveUsersToday:    activeUsers,
		TotalRequestsToday:  totalReq,
		AvgResponseTimeMs:   math.Round(avgRespMs*100) / 100,
		ErrorRatePercentage: math.Round(errRate*100) / 100,
		MatchesPlayedTotal:  matchesPlayed,
		MaterialsReadTotal:  materialsRead,
		SystemHealth: SystemHealth{
			UptimeSeconds:      uptimeSecs,
			Goroutines:         numGoroutines,
			MemoryUsageMB:      math.Round(memoryMB*100) / 100,
			DBOpenConnections:  openConns,
			DBInUseConnections: inUseConns,
		},
	}

	return summary, nil
}

func (s *service) GetApiLatency(ctx context.Context) (*ApiLatencyMetricsResponse, *shared.AppError) {
	items, err := s.repo.GetApiLatencyMetrics(ctx)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Round decimal values for clean responses
	for i := range items {
		items[i].AvgDurationMs = math.Round(items[i].AvgDurationMs*100) / 100
		items[i].P95DurationMs = math.Round(items[i].P95DurationMs*100) / 100
		items[i].P99DurationMs = math.Round(items[i].P99DurationMs*100) / 100
	}

	return &ApiLatencyMetricsResponse{
		Endpoints: items,
	}, nil
}

func (s *service) GetAuditLogs(ctx context.Context, filter AuditLogFilter) (*AuditLogsPaginatedResponse, *shared.AppError) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	logs, total, err := s.repo.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.PageSize)))
	}

	return &AuditLogsPaginatedResponse{
		Items:      logs,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) ExportAuditLogsCSV(ctx context.Context) ([]byte, *shared.AppError) {
	logs, err := s.repo.GetAllAuditLogsForExport(ctx)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV Header
	header := []string{
		"ID",
		"User ID",
		"User Email",
		"User Role",
		"Action",
		"Module",
		"Endpoint",
		"Method",
		"Status Code",
		"Duration (ms)",
		"IP Address",
		"Created At",
	}
	if err := writer.Write(header); err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Write data rows
	for _, log := range logs {
		userIDStr := ""
		if log.UserID != nil {
			userIDStr = *log.UserID
		}
		userEmailStr := ""
		if log.UserEmail != nil {
			userEmailStr = *log.UserEmail
		}
		userRoleStr := ""
		if log.UserRole != nil {
			userRoleStr = *log.UserRole
		}
		ipStr := ""
		if log.IPAddress != nil {
			ipStr = *log.IPAddress
		}

		row := []string{
			log.ID,
			userIDStr,
			userEmailStr,
			userRoleStr,
			log.Action,
			log.Module,
			log.Endpoint,
			log.Method,
			strconv.Itoa(log.StatusCode),
			strconv.Itoa(log.DurationMs),
			ipStr,
			log.CreatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			return nil, shared.ErrInternal(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return buf.Bytes(), nil
}
