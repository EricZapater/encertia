# UX — Filtre per Grup a les Avaluacions

**Veredicte**: A MILLORAR  
**Data**: 2026-10-06  

## Flux revisats
- **Llistat general d'avaluacions (`EvaluationsListView.vue`)**: Filtre global per grup mitjançant el desplegable `Select` de PrimeVue, amb actualització dinàmica de la taula i columpna `groupName`.
- **Avaluació de qüestionari (`QuizEvaluationView.vue`)**: Filtre per grup que recalcula tant les estadístiques globals com el llistat d'alumnes participants, incloent la columna de grup i l'acció de qualificació.
- **Detall i qualificació d'alumne (`StudentEvaluationView.vue`)**: Visualització del nom del grup de l'alumne com a `Tag` al capçal de la pàgina i l'historial complet de partides jugades.

## Usabilitat
- [OK] **Desplegable de selecció de grup clar**: El component `Select` mostra les opcions clarament estructurades amb el format `Nom (Curs Acadèmic)` (ex. `Grup A (26-27)`), fent evident a quin curs pertany cada grup.
- [OK] **Fàcil opció de reset ("Tots els grups")**: L'ús del prop `showClear` permet als usuaris cancel·lar la selecció del filtre ràpidament i tornar a la vista completa sense haver de recarregar.
- [OK] **Visualització de `groupName`**: A les taules de `EvaluationsListView` i `QuizEvaluationView`, la columna "Grup" inclou un valor per defecte (`-`) quan la partida no està vinculada a cap grup.
- [Millorable] **Falta de cerca textual per alumne o grup**: Tot i que el filtre per desplegable funciona correctament, en grups amb molts alumnes no hi ha cap camp d'entrada de cerca (`InputText`) per filtrar ràpidament per nom de la persona o per nom de grup dins la mateixa taula.

## Consistència visual
- [OK] **Ús coherent de PrimeVue**: Integració d'elements estàndard com `Select`, `DataTable`, `Column`, `Card`, `Tag`, `Button` i `InputNumber` mantenint la línia estilística del sistema.
- [OK] **Distinció visual de grup**: A la vista de l'alumne (`StudentEvaluationView.vue`), el grup es mostra destacat com un `Tag` secundari al costat del nom de l'alumne.
- [Millorable] **Cadena de text literal (hardcoded) sense i18n**: Excepte la clau `$t('evaluations.title')`, la majoria d'etiquetes i missatges (`"Tots els grups"`, `"Veure Avaluació"`, `"Estadístiques Globals per Pregunta"`, `"Resultats dels Alumnes"`, `"Qualificació Manual"`, etc.) estan escrits directament en català en lloc d'emprar el sistema d'i18n.

## Responsive/mòbil
- [OK] **Capçaleres adaptables**: Les barres d'eines superiors utilitzen `flex flex-column sm:flex-row`, fent que en dispositius mòbils el filtre de grup ocupi el 100% de l'amplada (`w-full sm:w-64`) sense encavallar-se amb el títol.
- [OK] **Taules responsives**: Les taules de dades incorporen `responsiveLayout="scroll"`, permetent la navegació tàctil horitzontal en pantalles estretes.

## Feedback a l'usuari
- [OK] **Indicadors de càrrega**: Indicador visual de progrés `:loading="store.isLoading"` a les taules de dades i spinner de càrrega centralitzat quan es recuperen dades.
- [OK] **Notificacions d'acció**: Qualificació manual d'alumnes a `StudentEvaluationView.vue` amb botó en estat de càrrega (`:loading="store.isSavingGrade"`) i toast d'èxit de PrimeVue confirmant la nota desada.

## Accessibilitat bàsica
- [OK] **Contrast visual i mides tàctils**: Textos amb bon contrast, botons d'acció amb mida adequada per a la selecció tàctil i icones representatives (`pi-eye`, `pi-pencil`, `pi-check-circle`).

## Incidències
1. **[Notable] Absència de cerca textual per nom de persona/grup a les vistes d'avaluació**: A `QuizEvaluationView.vue` i `EvaluationsListView.vue` no hi ha un camp de cerca per filtrar el llistat per nom d'alumne o nom de grup en temps real. S'hauria d'afegir un `InputText` amb filtre global a la `DataTable` per millorar l'agilitat de cerca del professorat.
2. **[Detall] Integració incompleta de traduccions i18n**: Cal extreure els textos literal de la interfície a `EvaluationsListView.vue`, `QuizEvaluationView.vue` i `StudentEvaluationView.vue` cap als fitxers de traducció (`ca.json`, `es.json`, `en.json`) per garantir la coherència multiidioma de l'aplicació.
