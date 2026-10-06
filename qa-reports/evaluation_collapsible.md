# QA — Mòdul Evaluation (Blocs Col·lapsables i Filtres)

**Veredicte**: APTE
**Data**: 2026-10-06

## Entorn de proves
- **Backend**: Compilació (`go build ./...`) i tests unitaris (`go test ./internal/evaluation/...`) completats amb èxit.
- **Frontend**: Tests unitaris (`pnpm test` -> 28 test suites, 148 tests passats) i build de producció (`pnpm build`) executats sense errors.
- **Mòdul de vistes**: `frontend/src/modules/evaluations/` (`EvaluationsListView.vue`, `QuizEvaluationView.vue`, `StudentEvaluationView.vue`).

## Proves executades i verificació funcional

### 1. Filtres superiors (Grup, Joc, Data)
- **`EvaluationsListView.vue`**:
  - Filtre de **Grup**: `Select` enllaçat amb `groupOptions` que permet triar el grup o veure tots els grups. Crida `@change="onGroupChange"` que refresca la llista des de backend filtrant per `groupId`.
  - Filtre de **Joc**: `Select` amb `quizOptions` per filtrar la taula de quizzes de manera reactiva.
  - Filtre de **Data**: `DatePicker` amb format `dd/mm/yy` i opció de neteja (`showClear`) per filtrar per la data de l'última partida.
- **`QuizEvaluationView.vue`**:
  - Filtres superiors integrats en una línia (`bg-surface-50 dark:bg-surface-800 p-3 border-round`).
  - El selector de **Joc** permet navegar directament a `/evaluations/quizzes/:selectedQuizId`.
  - El selector de **Grup** actualitza la vista d'avaluació del quiz especificat en temps real.

### 2. Blocs Col·lapsables (`Panel :toggleable="true"`)
- **`QuizEvaluationView.vue`**:
  - **Secció A: Estadístiques Globals per Pregunta**:
    - Encapsulada en un component `Panel` amb prop `:toggleable="true"` i capçalera "Estadístiques Globals per Pregunta".
    - Mostra taula amb número de pregunta, enunciat, taxa d'encert (amb etiqueta de color `Tag` segons percentatge), temps mitjà de resposta en segons, distribució de respostes i recompte de sense resposta.
  - **Secció B: Resultats dels Alumnes**:
    - Encapsulada en un component `Panel` amb prop `:toggleable="true"` i capçalera "Resultats dels Alumnes".
    - Taula paginada amb informació de l'alumne, grup, partides jugades, nota calculada i nota definitiva.
    - Accions condicionals: Etiqueta verda i botó "Editar Nota" si `isGraded: true`; "Pendent" i botó "Qualificar" en cas contrari.

### 3. Vista de Qualificació Manual i Historial
- **`StudentEvaluationView.vue`**:
  - Formulari de qualificació manual amb `InputNumber` amb límit `0.00 - 10.00` pre-omplert amb `calculatedGrade` o `finalGrade`.
  - Historial de partides participades i detall de respostes i temps.

## Compliment funcional
- [OK] Compliment del contracte `contracts/evaluation.spec.md` (seccions 5.1, 5.2 i 5.3).
- [OK] Filtres per Grup, Joc i Data operatius a totes les vistes d'avaluació.
- [OK] Panells col·lapsables integrats correctament mitjançant `PrimeVue Panel toggleable`.
- [OK] Integració fluida entre l'store Pinia (`useEvaluationStore`) i els stores de grups (`useGroupStore`) i quizzes (`useQuizStore`).

## Qualitat de codi
- [OK] TypeScript estrictament tipat a tots els components i stores de `evaluations`.
- [OK] Test unitaris complets a `EvaluationsViews.spec.ts` i `store.spec.ts`.
- [OK] Absència d'errors de linter, build de Vite/Rollup completat de manera neta.

## Homogeneïtat
- [OK] Segueix les convencions de disseny visual (PrimeVue, Flexbox utilitats CSS) i d'arquitectura Vue 3 / Pinia utilitzades a la resta del projecte.

## Incidències
Cap incidència detectada.
