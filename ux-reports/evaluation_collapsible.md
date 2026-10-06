# UX — Blocs Col·lapsables i Barra de Filtres a Avaluacions

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Flux revisats
- **Navegació i filtratge a la llista d'avaluacions (`EvaluationsListView.vue`)**: Revisa la barra de filtres superior (**Grup**, **Joc**, **Data**) i la seva interacció amb la taula principal d'avaluacions per qüestionari.
- **Detall i blocs col·lapsables a l'avaluació de qüestionari (`QuizEvaluationView.vue`)**: Revisa l'ús de panels col·lapsables PrimeVue (`Panel :toggleable="true"`) per separar "Estadístiques Globals per Pregunta" i "Resultats dels Alumnes", així com la barra de filtres superior (**Grup**, **Joc**, **Data**).

## Usabilitat
- [OK] **Estructura clara en blocs col·lapsables (`QuizEvaluationView.vue`)**: Els dos panels col·lapsables PrimeVue (`Panel :toggleable="true"`) permeten organitzar la informació en dues seccions diferenciades: estadístiques globals del conjunt de preguntes i resultats individuals dels alumnes.
- [OK] **Control d'opcions als desplegables de la barra de filtres**: Als desplegables de **Grup** i **Joc**, la incorporació de `showClear` permet restaurar fàcilment el filtre per defecte ("Tots els grups", "Tots els jocs").
- [OK] **Filtre de Data reactiu a `QuizEvaluationView.vue` i `EvaluationsListView.vue`**: El component `DatePicker` està vinculat reactivament via `v-model="selectedDate"` i aplica el filtratge per data (`same day`) a les propietats computades `filteredStudents` i `filteredEvaluations`.

## Consistència visual
- [OK] **Integració de components PrimeVue**: Ús homogeni de `Panel`, `Select`, `DatePicker`, `DataTable`, `Column` i `Tag` seguint el sistema de disseny de la plataforma.
- [OK] **Disseny de la barra de filtres**: Fons neutre (`bg-surface-50 dark:bg-surface-800`), cantonades arrodonides (`border-round`) i separació nítida (`p-3`, `gap-3`) coherent entre `EvaluationsListView.vue` i `QuizEvaluationView.vue`.
- [OK] **Internacionalització completa (i18n)**: Tots els textos literals, etiquetes de filtres, placeholders ("Tots els grups", "Filtrar per data"), capçaleres de panels i capçaleres de taula utilitzen claus de traducció `$t(...)`.

## Responsive/mòbil
- [OK] **Barra de filtres adaptable**: Implementació de `flex flex-column sm:flex-row flex-wrap` amb amplada mínima (`min-w-12rem`) que s'apila correctament en dispositius mòbils sense encavallaments.
- [OK] **Scroll horitzontal a taules**: Les taules interiors dels panels utilitzen `responsiveLayout="scroll"`, evitant desbordaments de pantalla en dispositius mòbils.

## Feedback a l'usuari
- [OK] **Estat de càrrega global**: Es mostra un spinner central (`pi-spin pi-spinner`) a `QuizEvaluationView.vue` i l'indicador `:loading="store.isLoading"` a `EvaluationsListView.vue`.
- [OK] **Slot d'estat buit (`#empty`) a les taules**: S'ha afegit `<template #empty>{{ $t('common.noResults') }}</template>` a totes les taules `DataTable` de les vistes d'avaluació, garantint un missatge d'estat buit quan cap registre coincideix amb els filtres.

## Accessibilitat bàsica
- [OK] **Contrast i mida dels controls**: Els selectors `Select` i `DatePicker` compleixen els requeriments de contrast visual i mides tàctils adequades per a dispositius mòbils.
- [OK] **Botó de plegar/desplegar accessible**: Els controls natius de PrimeVue `Panel` incorporen icones de col·lapse (`pi-chevron-down` / `pi-chevron-up`) amb comportament del teclat i atributs ARIA per defecte.

## Incidències
*(Sense incidències pendents. Totes les incidències detectades en la revisió prèvia han estat resoltes satisfactòriament).*
