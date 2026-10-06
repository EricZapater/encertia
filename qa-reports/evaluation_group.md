# QA — Filtre de Grup a les Avaluacions (`evaluation` + `group`)

**Veredicte**: **APTE**
**Data**: 2026-10-06

---

## 1. Entorn de Proves i Execució

- **Backend**: Verificat mitjançant anàlisi de codi i integració dels mòduls `internal/evaluation/` i `internal/group/`. La suite de proves unitàries (`go test -count=1 -v ./internal/evaluation/...`) s'ha executat amb resultat favorable (**PASS**, 0.337s), cobrint la filtració per `groupId` a la llista d'avaluacions, les estadístiques de quiz per grup, i la qualificació manual tant per a alumnes registrats com per a jugadors anònims.
- **Frontend**: Compilació executada satisfactòriament amb `pnpm build` a `frontend/`. Resultat: **Éxit de build** (`✓ built in 2.75s`), 0 errors de TypeScript i 0 errors de bundling Vite/Rollup.
- **Base de Dades / Esquema**: Esquema PostgreSQL integrat segons `contracts/evaluation.spec.md`. Suport per a la relació entre partides (`matches.group_id`), membres de grup (`group_students`) i registres d'avaluació (`evaluations`).
- **Dades de prova generades**: Generació de UUIDs sintètics en entorn d'unit testing (`group-123`, `user-1`, `player-anon-1`, etc.).

---

## 2. Proves Funcionals i de Coberta Executades

1. **Llistat d'avaluacions amb filtre de grup (`GET /evaluations?groupId=...`)**:
   - Petició enviada amb `groupId` opcional -> Filtra la llista de quizzes mostrant només aquelles partides associades al grup especificat o amb alumnes pertanyents a aquest grup (verificat a `TestEvaluationService_ListAndGrade`).
2. **Obtenir avaluació completa de quiz filtrada per grup (`GET /evaluations/quizzes/:quizId?groupId=...`)**:
   - Petició amb `quizId` i `groupId` -> Retorna `QuizEvaluationResponse` recalculant la taxa d'encert (`hitRate`), temps mitjà de resposta (`avgResponseTimeMs`), distribució de respostes i llista d'alumnes restringida al grup sol·licitat.
3. **Detall d'avaluació d'un alumne (`GET /evaluations/quizzes/:quizId/students/:studentId`)**:
   - Retorna `StudentEvaluationDetail` incloent `groupId` i `groupName` de l'alumne, així com l'historial de partides jugades i respostes per pregunta.
4. **Qualificació manual d'un alumne (`PUT /evaluations/quizzes/:quizId/students/:studentId/grade`)**:
   - Validació de rang de nota (\(0.00 \le \text{finalGrade} \le 10.00\)). Persistència a la base de dades i actualització de `is_graded = true`, `graded_by` i `graded_at`.
5. **Selector de grup al Frontend**:
   - Interfície d'usuari a `EvaluationsListView.vue` i `QuizEvaluationView.vue` que inclou el selector de grup (`Select` de PrimeVue) sincronitzat amb el Pinia store (`useEvaluationStore`), recarregant dinàmicament les dades en canviar de grup.

---

## 3. Compliment Funcional (`contracts/evaluation.openapi.yaml` i `contracts/evaluation.spec.md`)

- [OK] **Contracte OpenAPI**: Tot el mòdul d'avaluació compleix al 100% el contracte `contracts/evaluation.openapi.yaml`, respectant els paràmetres de cerca `groupId` a la query i els camps opcionalment nul·lables `groupId` i `groupName` als esquemes JSON de resposta.
- [OK] **Integració amb Grups**: Supostos d'enriquiment de grup coberts tant per a partides creades directament amb `group_id` com per a la coincidència de jugadors anònims adscrits al grup (`group_students`).
- [OK] **Control d'accés i RBAC**: Endpoint protegit per `BearerAuth`. El rol `student` rep un codi `403 Forbidden` a tots els endpoints d'avaluació. Els professors (`teacher`) només tenen accés als quizzes creats per ells mateixos.

---

## 4. Qualitat de Codi

- [OK] **Backend (Go)**: Estruturat en capes modulars (`handler.go`, `service.go`, `repository.go`, `model.go`).
- [OK] **SQL Injection**: Les consultes SQL a `repository.go` usen paràmetres posicionals (`$1`, `$2`, etc.) per evitar injeccions SQL.
- [OK] **Gestió d'Errors**: Retorn estandarditzat mitjançant el paquet `shared` (`ErrForbidden`, `ErrNotFound`, `ErrBadRequest`, `ErrInternal`).
- [OK] **Frontend (Vue 3 / TypeScript)**: Tipat estricte a `frontend/src/modules/evaluations/types.ts` (`groupId?: string | null`, `groupName?: string | null`). Crides REST centralitzades a `api.ts` utilitzant `apiClient`.
- [OK] **Build del Frontend**: Execució neta de `pnpm build` sense cap error de compilació TypeScript ni d'empaquetat Vite.

---

## 5. Homogeneïtat

- [OK] **Nomenclatura**: Noms de camps JSON en `camelCase` (`groupId`, `groupName`, `calculatedGrade`, `finalGrade`) i columnes SQL en `snake_case` (`group_id`, `calculated_grade`, `final_grade`).
- [OK] **Convencions del projecte**: Respecta `constitution.md` (screaming architecture en Go, mòduls Vue 3 a `src/modules/`, pnpm com a gestor de paquets).

---

## 6. Incidències

**Cap incidència detectada.** Tot el comportament funcional, el filtre de grup a les avaluacions, el contracte OpenAPI i la compilació compleixen al 100% els requisits especificats.

---

**Veredicte Final**: **APTE** (Aprovat per al merge i desplegament).
