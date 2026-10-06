# UX — Filtre de Data (DatePicker) a Avaluacions

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Flux revisats
- **Filtratge de qüestionaris per data a `EvaluationsListView.vue`**: Selecció d'una data al component `DatePicker` de la capçalera de filtres, comprovació del filtratge reactiu immediat sobre la llista d'avaluacions segons el dia de la darrera partida (`lastMatchAt`), neteja individual (`showClear`) i global ("Netejar Filtres").
- **Filtratge d'alumnes per data de partida a `QuizEvaluationView.vue`**: Selecció d'una data concreta per consultar els alumnes que van jugar la partida en aquella data (`s.lastMatchAt || s.matchDate`), comprovació de la reactivitat instantània a la taula de resultats d'alumnes, estat buit en absència de partides aquell dia, i restabliment dels filtres.

## Usabilitat
- [OK] **Filtratge reactiu i immediat**: En seleccionar qualsevol dia al calendari, la llista de qüestionaris a `EvaluationsListView` i la taula d'alumnes a `QuizEvaluationView` s'actualitzen a l'instant mitjançant propietats computades (`computed`), sense requerir cap clic addicional ni petició redundant al servidor.
- [OK] **Format de data amigable i estàndard (`dateFormat="dd/mm/yy"`)**: El format visual utilitza l'estàndard dia/mes/any (`DD/MM/YYYY`, ex. 06/10/2026), clar, familiar i sense ambigüitats per al professorat.
- [OK] **Normalització robusta de zones horàries**: La funció auxiliar `toLocalDateString` extreu any, mes i dia en hora local tant per al `Date` seleccionat com per a les dates en format ISO emmagatzemades, evitant discrepàncies per desfasament horari (UTC vs local).
- [OK] **Mecanismes duals de neteja**: L'usuari pot netejar la data directament des del control gràcies a la icona `showClear`, o bé restablir tots els criteris de cerca amb el botó "Netejar Filtres" de la capçalera del panell.
- [OK] **Text d'orientació (Placeholder)**: Disposa del placeholder internacionalitzat "Filtrar per data" (`$t('evaluations.filters.datePlaceholder')`) present a tots els idiomes suportats (català, castellà i anglès).

## Consistència visual
- [OK] **Integració amb PrimeVue (`DatePicker`)**: Utilitza el component oficial `DatePicker` de PrimeVue amb `showIcon` i `showClear`, mantenint la mateixa aparença visual i comportament que la resta d'inputs de la plataforma.
- [OK] **Alineació a la graella de filtres (`.filter-grid`)**: S'integra harmònicament com a tercer element de la targeta `.evaluation-filter-panel`, compartint tipografia d'etiquetes (`.filter-label`), alçada de camp i proporcions amb els selectors de Grup i Joc.

## Responsive/mòbil
- [OK] **Adaptabilitat a pantalles reduïdes**: A la graella CSS flexible (`repeat(auto-fit, minmax(220px, 1fr))`), el `DatePicker` s'expandeix al 100% de la seva cel·la (`.filter-datepicker { width: 100%; }`) i s'apila de manera natural en dispositius mòbils sense trencar el flux visual ni generar desplaçament horitzontal.
- [OK] **Usabilitat tàctil**: El diàleg emergent del calendari ofereix àrees tàctils suficients per navegar entre mesos i seleccionar dies còmodament des de tauletes o mòbils.

## Feedback a l'usuari
- [OK] **Estat buit coherent**: Si no hi ha cap partida jugada en la data seleccionada, les taules mostren clarament l'estat buit per defecte ("No s'han trobat resultats" / icona `pi pi-chart-bar`), comunicant de forma inequívoca que el filtre està actiu però no hi ha coincidències.
- [OK] **Indicació visual de data seleccionada**: El camp mostra la data triada de forma prominent i manté la icona de creu per facilitar la supressió amb un sol clic.

## Accessibilitat bàsica
- [OK] **Suport complet per a Mode Fosc**: Els selectors i el calendari hereten correctament la paleta de colors del tema fosc amb alt contrast en textos, vores i dies seleccionables.
- [OK] **Etiqueta semàntica**: El control està acompanyat d'una etiqueta explícita (`<label class="filter-label">`) que identifica clarament la finalitat del camp.

## Incidències
*(Sense incidències. El filtre de data proporciona una experiència àgil, intuïtiva i plenament reactiva tant al llistat general com al detall de qüestionari).*
