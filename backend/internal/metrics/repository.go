package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/encertia/backend/internal/shared"
)

type Repository interface {
	InsertAuditLog(ctx context.Context, entry shared.AuditLogInput) error
	GetMetricsSummaryData(ctx context.Context) (activeUsersToday int, totalRequestsToday int, avgResponseTimeMs float64, errorRate float64, matchesPlayed int, materialsRead int, err error)
	GetApiLatencyMetrics(ctx context.Context) ([]EndpointLatencyItem, error)
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error)
	GetAllAuditLogsForExport(ctx context.Context) ([]AuditLog, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) InsertAuditLog(ctx context.Context, entry shared.AuditLogInput) error {
	if r.db == nil {
		return nil
	}

	query := `
		INSERT INTO audit_logs (
			user_id, user_email, user_role, action, module, endpoint, method, status_code, duration_ms, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		entry.UserID,
		entry.UserEmail,
		entry.UserRole,
		entry.Action,
		entry.Module,
		entry.Endpoint,
		entry.Method,
		entry.StatusCode,
		entry.DurationMs,
		entry.IPAddress,
		entry.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("error inserint audit log: %w", err)
	}

	return nil
}

func (r *sqlRepository) GetMetricsSummaryData(ctx context.Context) (int, int, float64, float64, int, int, error) {
	if r.db == nil {
		return 0, 0, 0, 0, 0, 0, nil
	}

	var activeUsersToday int
	var totalRequestsToday int
	var avgResponseTimeMs float64
	var errorRate float64
	var matchesPlayed int
	var materialsRead int

	// 1. Active users today
	queryActive := `SELECT COUNT(DISTINCT user_id) FROM audit_logs WHERE created_at >= CURRENT_DATE AND user_id IS NOT NULL`
	_ = r.db.QueryRowContext(ctx, queryActive).Scan(&activeUsersToday)

	// 2. Total requests today
	queryTotalReq := `SELECT COUNT(*) FROM audit_logs WHERE created_at >= CURRENT_DATE`
	_ = r.db.QueryRowContext(ctx, queryTotalReq).Scan(&totalRequestsToday)

	// 3. Avg response time & Error rate today
	queryStats := `
		SELECT
			COALESCE(AVG(duration_ms), 0),
			COALESCE(COUNT(*) FILTER (WHERE status_code >= 400) * 100.0 / NULLIF(COUNT(*), 0), 0)
		FROM audit_logs
		WHERE created_at >= CURRENT_DATE
	`
	_ = r.db.QueryRowContext(ctx, queryStats).Scan(&avgResponseTimeMs, &errorRate)

	// 4. Matches played total
	queryMatches := `SELECT COUNT(*) FROM matches WHERE deleted_at IS NULL`
	_ = r.db.QueryRowContext(ctx, queryMatches).Scan(&matchesPlayed)

	// 5. Materials read total
	queryMaterials := `SELECT COUNT(*) FROM material_views`
	_ = r.db.QueryRowContext(ctx, queryMaterials).Scan(&materialsRead)

	return activeUsersToday, totalRequestsToday, avgResponseTimeMs, errorRate, matchesPlayed, materialsRead, nil
}

func (r *sqlRepository) GetApiLatencyMetrics(ctx context.Context) ([]EndpointLatencyItem, error) {
	if r.db == nil {
		return []EndpointLatencyItem{}, nil
	}

	query := `
		SELECT
			method,
			endpoint,
			COUNT(*) as request_count,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 0) as p95_duration_ms,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms), 0) as p99_duration_ms,
			COUNT(*) FILTER (WHERE status_code >= 400) as error_count
		FROM audit_logs
		GROUP BY method, endpoint
		ORDER BY request_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error consultant latència d'API: %w", err)
	}
	defer rows.Close()

	var items []EndpointLatencyItem
	for rows.Next() {
		var item EndpointLatencyItem
		if err := rows.Scan(
			&item.Method,
			&item.Endpoint,
			&item.RequestCount,
			&item.AvgDurationMs,
			&item.P95DurationMs,
			&item.P99DurationMs,
			&item.ErrorCount,
		); err != nil {
			return nil, fmt.Errorf("error llegint dades de latència: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []EndpointLatencyItem{}
	}

	return items, nil
}

func (r *sqlRepository) GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error) {
	if r.db == nil {
		return []AuditLog{}, 0, nil
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(user_email ILIKE $%d OR endpoint ILIKE $%d OR action ILIKE $%d OR ip_address ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	if filter.Module != "" {
		conditions = append(conditions, fmt.Sprintf("module = $%d", argIdx))
		args = append(args, filter.Module)
		argIdx++
	}

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error comptant registres d'auditoria: %w", err)
	}

	// 2. Items query
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	itemsQuery := fmt.Sprintf(`
		SELECT id, user_id, user_email, user_role, action, module, endpoint, method, status_code, duration_ms, ip_address, user_agent, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	itemArgs := append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, itemsQuery, itemArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error consultant registres d'auditoria: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var logItem AuditLog
		var userID, userEmail, userRole, ipAddress, userAgent sql.NullString

		if err := rows.Scan(
			&logItem.ID,
			&userID,
			&userEmail,
			&userRole,
			&logItem.Action,
			&logItem.Module,
			&logItem.Endpoint,
			&logItem.Method,
			&logItem.StatusCode,
			&logItem.DurationMs,
			&ipAddress,
			&userAgent,
			&logItem.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("error llegint línia d'audit log: %w", err)
		}

		if userID.Valid {
			logItem.UserID = &userID.String
		}
		if userEmail.Valid {
			logItem.UserEmail = &userEmail.String
		}
		if userRole.Valid {
			logItem.UserRole = &userRole.String
		}
		if ipAddress.Valid {
			logItem.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			logItem.UserAgent = &userAgent.String
		}

		logs = append(logs, logItem)
	}

	if logs == nil {
		logs = []AuditLog{}
	}

	return logs, total, nil
}

func (r *sqlRepository) GetAllAuditLogsForExport(ctx context.Context) ([]AuditLog, error) {
	if r.db == nil {
		return []AuditLog{}, nil
	}

	query := `
		SELECT id, user_id, user_email, user_role, action, module, endpoint, method, status_code, duration_ms, ip_address, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 5000
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error consultant registres d'auditoria per exportar: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var logItem AuditLog
		var userID, userEmail, userRole, ipAddress, userAgent sql.NullString

		if err := rows.Scan(
			&logItem.ID,
			&userID,
			&userEmail,
			&userRole,
			&logItem.Action,
			&logItem.Module,
			&logItem.Endpoint,
			&logItem.Method,
			&logItem.StatusCode,
			&logItem.DurationMs,
			&ipAddress,
			&userAgent,
			&logItem.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error llegint audit log per exportar: %w", err)
		}

		if userID.Valid {
			logItem.UserID = &userID.String
		}
		if userEmail.Valid {
			logItem.UserEmail = &userEmail.String
		}
		if userRole.Valid {
			logItem.UserRole = &userRole.String
		}
		if ipAddress.Valid {
			logItem.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			logItem.UserAgent = &userAgent.String
		}

		logs = append(logs, logItem)
	}

	if logs == nil {
		logs = []AuditLog{}
	}

	return logs, nil
}
