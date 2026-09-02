package metrics

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu        sync.Mutex
	auditLogs []AuditLog
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		auditLogs: make([]AuditLog, 0),
	}
}

func (m *mockRepository) InsertAuditLog(ctx context.Context, entry shared.AuditLogInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logItem := AuditLog{
		ID:         uuid.New().String(),
		UserID:     entry.UserID,
		UserEmail:  entry.UserEmail,
		UserRole:   entry.UserRole,
		Action:     entry.Action,
		Module:     entry.Module,
		Endpoint:   entry.Endpoint,
		Method:     entry.Method,
		StatusCode: entry.StatusCode,
		DurationMs: entry.DurationMs,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now(),
	}

	m.auditLogs = append(m.auditLogs, logItem)
	return nil
}

func (m *mockRepository) GetMetricsSummaryData(ctx context.Context) (int, int, float64, float64, int, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	activeUsersMap := make(map[string]bool)
	totalReq := len(m.auditLogs)
	totalDuration := 0
	errorCount := 0

	for _, log := range m.auditLogs {
		if log.UserID != nil {
			activeUsersMap[*log.UserID] = true
		}
		totalDuration += log.DurationMs
		if log.StatusCode >= 400 {
			errorCount++
		}
	}

	avgResp := 0.0
	errRate := 0.0
	if totalReq > 0 {
		avgResp = float64(totalDuration) / float64(totalReq)
		errRate = (float64(errorCount) * 100.0) / float64(totalReq)
	}

	return len(activeUsersMap), totalReq, avgResp, errRate, 10, 25, nil
}

func (m *mockRepository) GetApiLatencyMetrics(ctx context.Context) ([]EndpointLatencyItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	type key struct {
		method   string
		endpoint string
	}

	stats := make(map[key]*EndpointLatencyItem)

	for _, log := range m.auditLogs {
		k := key{method: log.Method, endpoint: log.Endpoint}
		item, exists := stats[k]
		if !exists {
			item = &EndpointLatencyItem{
				Method:   log.Method,
				Endpoint: log.Endpoint,
			}
			stats[k] = item
		}
		item.RequestCount++
		item.AvgDurationMs += float64(log.DurationMs)
		if log.StatusCode >= 400 {
			item.ErrorCount++
		}
	}

	items := make([]EndpointLatencyItem, 0, len(stats))
	for _, item := range stats {
		if item.RequestCount > 0 {
			item.AvgDurationMs = item.AvgDurationMs / float64(item.RequestCount)
			item.P95DurationMs = item.AvgDurationMs * 1.1
			item.P99DurationMs = item.AvgDurationMs * 1.2
		}
		items = append(items, *item)
	}

	return items, nil
}

func (m *mockRepository) GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var filtered []AuditLog
	for _, log := range m.auditLogs {
		if filter.Module != "" && log.Module != filter.Module {
			continue
		}
		if filter.UserID != "" && (log.UserID == nil || *log.UserID != filter.UserID) {
			continue
		}
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			match := strings.Contains(strings.ToLower(log.Action), searchLower) ||
				strings.Contains(strings.ToLower(log.Endpoint), searchLower)
			if log.UserEmail != nil {
				match = match || strings.Contains(strings.ToLower(*log.UserEmail), searchLower)
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, log)
	}

	total := len(filtered)
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []AuditLog{}, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (m *mockRepository) GetAllAuditLogsForExport(ctx context.Context) ([]AuditLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logsCopy := make([]AuditLog, len(m.auditLogs))
	copy(logsCopy, m.auditLogs)
	return logsCopy, nil
}
