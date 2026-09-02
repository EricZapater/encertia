# QA — Mòdul Mètriques, Monitoratge i Auditoria (`metrics`) (v1.2.0)

**Veredicte**: APTE  
**Data**: 2026-09-02  

## Entorn de proves
- Backend engegat: SÍ (`go test -v ./internal/metrics/...` i `go vet ./...` executats amb èxit, exit code 0).
- Frontend engegat: SÍ (`pnpm run type-check`, `pnpm run test` i `pnpm run build` executats amb èxit, exit code 0).
- Base de dades local: connectada via mock i migració 000011 PostgreSQL (`audit_logs`).
- Dades de prova generades: Mock data a suites de unit tests (`TestMetricsService` i `TestMetricsEndpoints`).

## Proves funcionals executades
- `GET /metrics/summary` amb rol `admin` → **200 OK** (Resum de mètriques de sistema, usuaris actius, latència i salut).
- `GET /metrics/summary` amb rol `teacher` o `student` → **403 Forbidden** (Accés denegat per restricció RBAC estricta).
- `GET /metrics/summary` sense token JWT → **401 Unauthorized**.
- `GET /metrics/api-latency` → **200 OK** (Retorna latència mitjana, p95, p99 i taxa d'errors per endpoint).
- `GET /metrics/audit-logs?page=1&pageSize=20` → **200 OK** (Retorna llista paginada de registres d'auditoria).
- `GET /metrics/audit-logs/export` → **200 OK** (Baixada de fitxer `text/csv; charset=utf-8` amb `Content-Disposition`).
- Navegació frontend a la vista `/metrics` amb rol `teacher` → Redirecció automàtica a `/profile` via `router.beforeEach`.

## Compliment funcional
- [OK] **HU-MET-01 i HU-MET-02 Audit Log d'Accions d'Usuari** — Registre asíncron a la taula `audit_logs` a PostgreSQL de cada acció (`user_id`, `user_email`, `user_role`, `action`, `module`, `endpoint`, `duration_ms`, `status_code`). Taula paginada amb filtres per usuari, mòdul i cerca.
- [OK] **HU-MET-03 Exportació CSV d'Auditoria** — Descarrega en 1 clic del fitxer CSV d'auditoria per a inspeccions externes (`GET /metrics/audit-logs/export`).
- [OK] **HU-MET-04 a HU-MET-06 Mètriques de Latència de l'API** — Middleware Gin de càlcul de latència en ms, percentils **p95** i **p99**, peticions per segon (RPS) i taxa d'errors HTTP. Rànquing d'endpoints més lents i amb major taxa d'error.
- [OK] **HU-MET-07 a HU-MET-09 Mètriques d'Engagement i Salut del Sistema** — Comptadors de partides en directe jugades, visualitzacions de materials, uptime del servidor Go, goroutines, memòria RAM i estat del pool de PostgreSQL.

## Qualitat de codi
- [OK] **Compilació i tests Backend** — `go test -v ./...` i `go vet ./...` 100% verds a `backend/internal/metrics`.
- [OK] **Compilació i tests Frontend** — `pnpm run type-check`, `pnpm run test` (27 fitxers de test, 140 tests aprovats) i `pnpm run build` 100% verds a `frontend/`.
- [OK] **Seguretat i RBAC** — Restricció estricta d'accés exclusivament al rol `admin` (`RequireRole("admin")` a l'API Gin i `roles: ['admin']` a Vue Router).

## Homogeneïtat
- [OK] **Estructura de carpetes** — `internal/metrics` (singular backend: `handler.go`, `service.go`, `repository.go`, `model.go`) i `src/modules/metrics` (plural frontend: `views/`, `store.ts`, `api.ts`, `types.ts`, `__tests__/`).
- [OK] **Nomenclatura d'estructures** — Respostes unificades amb `MetricsSummaryResponse`, `ApiLatencyMetricsResponse`, `AuditLogsPaginatedResponse` tant al contracte OpenAPI com als tipus TypeScript i Go structs.

## Incidències
Cap incidència detectada. El mòdul `metrics` compleix totalment amb les especificacions de `contracts/metrics.openapi.yaml` i `specs/metrics.md`.
