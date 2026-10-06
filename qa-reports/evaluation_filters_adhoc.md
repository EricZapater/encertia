# QA — Mòdul Evaluation (Panell de Filtres amb CSS Ad-hoc)

**Veredicte**: APTE
**Data**: 2026-10-06

## Entorn de proves
- **Backend**: Tests unitaris (`go test ./internal/evaluation/...` i `go test ./...`) passats correctament.
- **Frontend**: 
  - Tests unitaris (`pnpm test` -> 28 test suites, 152 tests passats) executats amb èxit.
  - Build de producció (`pnpm build`) completat sense errors.
- **Fitxers revisats**:
  - `frontend/src/modules/evaluations/views/EvaluationsListView.vue`
  - `frontend/src/modules/evaluations/views/QuizEvaluationView.vue`
  - `frontend/src/modules/evaluations/__tests__/EvaluationsViews.spec.ts`
  - `frontend/src/i18n/locales/ca.json`, `es.json`, `en.json`

## Proves executades i verificació funcional

### 1. Disseny i Estils CSS Ad-hoc (Sense dependències Tailwind)
- **Substitució completa de classes Tailwind**:
  - Els components de filtres s'han estilitzat exclusivament mitjançant classes scoped i CSS natiu: `.evaluation-filter-panel`, `.filter-panel-card`, `.filter-panel-header`, `.filter-title`, `.filter-clear-btn`, `.filter-grid`, `.filter-item`, `.filter-label`, `.filter-select`, `.filter-datepicker`.
  - Distribució responsiva moderna basada en CSS Grid (`grid-template-columns: repeat(auto-fit, minmax(220px, 1fr))`) amb alineació inferior d'elements per mantenir l'harmonia visual amb els selectors i `DatePicker`.
  - Estil visual coherent amb la resta de l'aplicació: targeta blanca amb vora subtil (`#e2e8f0`), ombra suau (`box-shadow`), capçalera amb icona `pi-filter` blava i botó de neteja d'acció ràpida.
- **Suport complet per a Mode Fosc**:
  - Regles CSS dedicades per a classe `:global(.dark-mode)` i `@media (prefers-color-scheme: dark)`.
  - Colors ajustats per a fons foscos (`#1e293b`), vores (`#334155`), textos secundaris (`#94a3b8`) i icones.

### 2. Reactivitat i Comportament dels Filtres
- **`EvaluationsListView.vue`**:
  - Filtre de **Grup**: Reactiu via `selectedGroupId` i `groupOptions`. En canviar de selecció (`@change="onGroupChange"`), crida `store.fetchEvaluationsList(groupId)` per actualitzar les dades del servidor.
  - Filtre de **Joc**: Reactiu via `selectedQuizId` i `quizOptions` (calculat unint els jocs del quizStore i els presents a la llista d'avaluacions).
  - Filtre de **Data**: Reactiu via `selectedDate` (`DatePicker`), comparant any, mes i dia amb la propietat `lastMatchAt`.
  - Còmput reactiu `filteredEvaluations` que aplica de forma combinada els filtres de grup, joc i data.
  - Botó de **Neteja**: `clearFilters()` reinicia totes les variables d'estat a `null` i recarrega la llista completa des del backend.
- **`QuizEvaluationView.vue`**:
  - Filtre de **Grup**: Reactiu; en canviar refina els resultats del quiz actiu (`fetchQuizEvaluation`).
  - Filtre de **Joc**: Reactiu; canviar de joc navega directament a la ruta de detall del nou quiz (`/evaluations/quizzes/:quizId`).
  - Filtre de **Data**: Filtra la llista d'alumnes (`filteredStudents`) comparant la data de l'última partida (`lastMatchAt` / `matchDate`).
  - Botó de **Neteja**: Restableix filtres i refà la petició al servei.

### 3. Internacionalització (i18n)
- Tots els literals dels panells de filtres utilitzen el sistema `$t(...)`:
  - `evaluations.filters.title` ("Filtres d'Avaluació" / "Filtros de Evaluación" / "Evaluation Filters")
  - `evaluations.filters.clear` ("Netejar Filtres" / "Limpiar Filtros" / "Clear Filters")
  - `evaluations.filters.group` ("Grup" / "Grupo" / "Group")
  - `evaluations.filters.allGroups` ("Tots els grups" / "Todos los grupos" / "All groups")
  - `evaluations.filters.game` ("Joc" / "Juego" / "Game")
  - `evaluations.filters.allGames` ("Tots els jocs" / "Todos los juegos" / "All games")
  - `evaluations.filters.selectGame` ("Selecciona un joc" / "Selecciona un juego" / "Select a game")
  - `evaluations.filters.date` ("Data" / "Fecha" / "Date")
  - `evaluations.filters.datePlaceholder` ("Filtrar per data" / "Filtrar por fecha" / "Filter by date")
- Verificat el suport a `ca.json`, `es.json` i `en.json`.

## Compliment funcional
- [OK] Estils ad-hoc encapsulats sense ús de classes d'utilitat Tailwind.
- [OK] Reactivitat correcta a la selecció de grup, joc i data, així com la neteja de filtres.
- [OK] Cobertura de tests unitaris per al panell de filtres i accions de neteja.
- [OK] Suport d'internacionalització complet en català, castellà i anglès.

## Qualitat de codi
- [OK] Codi TypeScript net i sense tipatges laxes.
- [OK] Proves unitàries de Vitest passades amb èxit (152/152 tests).
- [OK] Compilació de Vite/Rollup completada satisfactòriament.

## Homogeneïtat
- [OK] Els panells de filtres mantenen una estructura i comportament idèntics entre el llistat d'avaluacions (`EvaluationsListView`) i la vista detallada de quiz (`QuizEvaluationView`).

## Incidències
Cap incidència detectada.
