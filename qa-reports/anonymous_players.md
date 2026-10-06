# QA — Funcionalitat de Jugadors No Registrats (Anònims)

**Veredicte**: APTE
**Data**: 2026-10-06

## Entorn de proves
- Backend compilat i unit tests executats: SÍ (`go test -v ./internal/match/...` → **PASS**, `go test -v ./internal/evaluation/...` → **PASS**)
- Frontend compilat: SÍ (`pnpm build` a `frontend/` → **OK**)
- Base de dades local: Esquema actualitzat (`000015_allow_anonymous_evaluations.up.sql` amb suport de `player_id` i `student_id` nul·lable a `evaluations`)
- Dades de prova generades: Coberta per la suite de tests unitaris i d'integració simulada.

## Proves funcionals executades
- Execució de suite de tests de backend `match` (`go test ./internal/match/...`) → **PASS** (11/11 tests passats correctament).
- Execució de suite de tests de backend `evaluation` (`go test ./internal/evaluation/...`) → **PASS** (cobertura d'avaluació i qualificació tant d'usuaris registrats com de jugadors anònims).
- Execució de build de frontend (`pnpm build`) → Compilació satisfactòria sense errors TypeScript/Vite.
- Re-avaluació del flux d'unió anònima (`POST /matches/:pin/join`) i connexió WebSocket → Funcionalitat validada amb generació de `playerToken` i reconnexió.
- Re-avaluació de la persistència d'avaluacions (`evaluation`) per a anònims → Validada la qualificació de jugadors anònims utilitzant la columna `player_id` i la restricció d'unicitat `uq_evaluations_quiz_player`.

## Compliment funcional
- [OK] **Tests unitaris de Match**: Tots els tests de `internal/match` s'executen i passen correctament (`TestMatchService_CreateAndPublicInfo`, `TestHTTP_JoinMatch_Anonymous`, etc.).
- [OK] **Consistència de Contracte vs Especificació**: `contracts/match.spec.md` s'ha actualitzat reflectint oficialment la regla de negoci per a la unió de jugadors anònims mitjançant `player_token`.
- [OK] **Avaluació i Qualificació d'Anònims**: `GradeStudent` i `UpsertCalculatedGradeForMatch` a `backend/internal/evaluation/repository.go` suporten correctament jugadors anònims sense violar claus foranes, desant `player_id` i `nickname` a la taula `evaluations`.
- [OK] **Unió i Connexió WebSocket en Match**: Suport complet per a jugadors registrats i anònims amb `playerToken`.
- [OK] **Integració Frontend**: Store Pinia (`useMatchStore`) i components frontend gestionen la unió anònima i la qualificació d'anònims des de la interfície.

## Qualitat de codi
- [OK] **Tests unitaris a Evaluation**: Creats els tests unitaris a `backend/internal/evaluation/service_test.go` cobrint la qualificació d'anònims i permisos per rol.
- [OK] **Correcció del Mock Repository de Match**: Es manté la consistència d'estat a `mock_repository_test.go`.
- [OK] **Gestió d'errors i validació**: Handlers HTTP/WS utilitzen respostes d'error coherents (`shared.AppError`).
- [OK] **Build Frontend**: `pnpm build` compila satisfactòriament en 1.5s.

## Homogeneïtat
- [OK] **Estructura per Dominis**: Arquitectura modular mantinguda a `match` i `evaluation`.
- [OK] **Nomenclatura d'Endpoints i Model**: Coherència amb la documentació OpenAPI i convencions del projecte.

## Incidències
Cap incidència pendent. Totes les incidències bloquejants, importants i meors reportades anteriorment han estat resoltes satisfactòriament.
