# QA — Mòdul Mètriques, Monitoratge i Auditoria (`metrics`) (v1.2.0)

**Veredicte**: APTE  
**Data**: 2026-09-02  

## Compliment funcional
- [OK] **HU-MET-01 i HU-MET-02 Audit Log d'Accions d'Usuari** — Registre asíncron a la taula `audit_logs` a PostgreSQL de cada acció (`user_id`, `user_email`, `user_role`, `action`, `module`, `endpoint`, `duration_ms`, `status_code`). Taula `DataTable` paginada amb filtres per usuari, mòdul i cerca.
- [OK] **HU-MET-03 Exportació CSV d'Auditoria** — Descarrega en 1 clic del fitxer CSV d'auditoria per a inspeccions externes (`GET /metrics/audit-logs/export`).
- [OK] **HU-MET-04 a HU-MET-06 Mètriques de Latència de l'API** — Middleware Gin de càlcul de latència en ms, percentils **p95** i **p99**, peticions per segon (RPS) i taxa d'errors HTTP. Rànquing d'endpoints més lents i amb major taxa d'error.
- [OK] **HU-MET-07 a HU-MET-09 Mètriques d'Engagement i Salut del Sistema** — Comptadors de partides en directe jugades, visualitzacions de materials, uptime del servidor Go, goroutines, memòria i estat del pool de PostgreSQL.

## Qualitat de codi
- [OK] **Compilació i tests** — Suite de Go (`go test ./...`) i tests frontend/type-check (`pnpm run type-check && pnpm run test`) 100% verds (27 fitxers de test, 140 tests aprovats).
- [OK] **Seguretat i RBAC** — Restricció estricta d'accés exclusivament al rol `admin` (`RequireRole("admin")` a l'API i `roles: ['admin']` a Vue Router).

## Homogeneïtat
- [OK] **Estructura de carpetes** — `internal/metrics` (singular backend) i `src/modules/metrics` (plural frontend).

## Incidències
Cap incidència detectada. El mòdul compleix totalment amb les especificacions de `contracts/metrics.openapi.yaml` i `specs/metrics.md`.
