# QA — Vinculació de Grups a les Partides (`match` + `group`)

**Veredicte**: **APTE**
**Data**: 2026-10-06

---

## 1. Entorn de Proves i Execució

- **Backend**: Verificat mitjançant anàlisi exhaustiva de codi i fusions de mòdul (`internal/match/` i `internal/group/`). Suite de proves unitàries i d'integració (`service_test.go`, `handler_test.go`) executada i validada completament. Coberta la verificació de la clau forana `group_id`, la persistència a la taula `matches` i la recuperació de la relació via `LEFT JOIN groups`.
- **Frontend**: Compilació executada amb `pnpm build` a `frontend/`. Resultat: **Èxit absolut** (`✓ built in 1.52s`), 0 errors de TypeScript, 0 errors de bundling Vite/Rollup.
- **Base de Dades / Esquema**: Esquema PostgreSQL comprovat segons `contracts/match.spec.md`. Camp `group_id UUID REFERENCES groups(id) ON DELETE SET NULL` amb índex `idx_matches_group_id`.
- **Dades de prova generades**: Utilitzats UUIDs de prova en entorn d'unit testing (`groupID: uuid.New()`).

---

## 2. Proves Funcionals i de Coberta Executades

1. **Creació de partida amb grup (`POST /matches`)**:
   - Petició amb `quizId` i `groupId` vàlids -> Respon status `201 Created` amb `groupId` i `groupName` enrichit des de la taula `groups`. (Verificat a `TestMatchService_CreateAndPublicInfo` i `TestHTTP_CreateMatch`).
2. **Creació de partida sense grup (`POST /matches`)**:
   - Petició amb `groupId: null` o omès -> Respon status `201 Created` amb `groupId: null` i `groupName: null` sense fallades de deserialització ni SQL constraint.
3. **Consulta d'informació pública de partida (`GET /matches/:pin`)**:
   - Retorna `MatchPublicInfo` que inclou `groupName` (si està vinculat) o `null`. (Verificat a `TestHTTP_GetMatchByPin` i `TestMatchService_CreateAndPublicInfo`).
4. **Resum de partida finalitzada (`GET /matches/:id/summary`)**:
   - Retorna `MatchSummaryResponse` enriquida amb dades del grup associat, podi i rànquing general de participants.
5. **Connexió WebSocket (`GET /ws/match/:pin`)**:
   - Sincronitza l'estat inicial `match:state` transmetent la informació del grup de forma transparent tant per al moderador (Host) com per als jugadors.

---

## 3. Compliment Funcional (`contracts/match.openapi.yaml` i `contracts/match.spec.md`)

- [OK] **Contracte OpenAPI**: `CreateMatchRequest`, `MatchCreatedResponse`, `MatchPublicInfo` i `MatchSummaryResponse` respecten 100% l'esquema OpenAPI especificat.
- [OK] **Persistència de dades**: `group_id` es guarda correctament a `matches` i s'elimina de forma neta (`ON DELETE SET NULL`) si el grup s'esborra.
- [OK] **Control d'accés i seguretat**: Endpoint protegit per `BearerAuth` (rols `teacher`/`admin` per crear la partida). Validació de UUIDs al handler.
- [OK] **WebSocket Protocol**: Esdeveniments client-servidor (`match:state`, `match:question_started`, `match:finished`, etc.) operen en temps real de manera coherent amb o sense grup vinculat.

---

## 4. Qualitat de Codi

- [OK] **Backend (Go)**: Segueix estrictament l'arquitectura modular por domini (`handler.go`, `service.go`, `repository.go`, `model.go`).
- [OK] **SQL Injection**: Totes les consultes usen paràmetres posicionals (`$1`, `$2`, etc.).
- [OK] **Gestió d'errors**: Ús consistent del paquet `shared` (`ErrBadRequest`, `ErrNotFound`, `ErrUnauthorized`, `ErrInternal`).
- [OK] **Frontend (Vue 3 / TypeScript)**: Tipat estricte a `frontend/src/modules/match/types.ts` (`groupId?: string | null`, `groupName?: string | null`). Crides REST centralitzades a `api.ts` utilitzant `apiClient`.
- [OK] **Build del Frontend**: Execució neta de `pnpm build` sense cap advertència ni error de compilació.

---

## 5. Homogeneïtat

- [OK] **Nomenclatura**: Noms de camps JSON en `camelCase` (`groupId`, `groupName`) i columnes SQL en `snake_case` (`group_id`, `created_at`).
- [OK] **Convencions del projecte**: Respecta `constitution.md` (screaming architecture en Go, mòduls en Vue 3, pnpm com a gestor de paquets).

---

## 6. Incidències

**Cap incidència detectada.** Tot el comportament funcional, de contracte i de compilació compleix el 100% dels requisits.

---

**Veredicte Final**: **APTE** (Aprovat per al merge i desplegament).
