package metrics

import "time"

// AuditLog representa un registre d'auditoria del sistema.
type AuditLog struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"userId,omitempty"`
	UserEmail  *string   `json:"userEmail,omitempty"`
	UserRole   *string   `json:"userRole,omitempty"`
	Action     string    `json:"action"`
	Module     string    `json:"module"`
	Endpoint   string    `json:"endpoint"`
	Method     string    `json:"method"`
	StatusCode int       `json:"statusCode"`
	DurationMs int       `json:"durationMs"`
	IPAddress  *string   `json:"ipAddress,omitempty"`
	UserAgent  *string   `json:"userAgent,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// SystemHealth informació de salut i recursos del servidor.
type SystemHealth struct {
	UptimeSeconds      int64   `json:"uptimeSeconds"`
	Goroutines         int     `json:"goroutines"`
	MemoryUsageMB      float64 `json:"memoryUsageMb"`
	DBOpenConnections  int     `json:"dbOpenConnections"`
	DBInUseConnections int     `json:"dbInUseConnections"`
}

// MetricsSummary dades de resum global de mètriques i salut.
type MetricsSummary struct {
	ActiveUsersToday    int          `json:"activeUsersToday"`
	TotalRequestsToday  int          `json:"totalRequestsToday"`
	AvgResponseTimeMs   float64      `json:"avgResponseTimeMs"`
	ErrorRatePercentage float64      `json:"errorRatePercentage"`
	MatchesPlayedTotal  int          `json:"matchesPlayedTotal"`
	MaterialsReadTotal  int          `json:"materialsReadTotal"`
	SystemHealth        SystemHealth `json:"systemHealth"`
}

// EndpointLatencyItem rendiment i latència d'un endpoint específic.
type EndpointLatencyItem struct {
	Method        string  `json:"method"`
	Endpoint      string  `json:"endpoint"`
	RequestCount  int     `json:"requestCount"`
	AvgDurationMs float64 `json:"avgDurationMs"`
	P95DurationMs float64 `json:"p95DurationMs"`
	P99DurationMs float64 `json:"p99DurationMs"`
	ErrorCount    int     `json:"errorCount"`
}

// ApiLatencyMetricsResponse resposta per a GET /metrics/api-latency.
type ApiLatencyMetricsResponse struct {
	Endpoints []EndpointLatencyItem `json:"endpoints"`
}

// AuditLogsPaginatedResponse resposta paginada per a GET /metrics/audit-logs.
type AuditLogsPaginatedResponse struct {
	Items      []AuditLog `json:"items"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

// AuditLogFilter filtres de cerca per a consultes de registres d'auditoria.
type AuditLogFilter struct {
	Page     int
	PageSize int
	Search   string
	Module   string
	UserID   string
}
