# UX — Panell de Filtres d'Avaluacions amb CSS Ad-hoc

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Flux revisats
- **Llistat d'avaluacions (`EvaluationsListView.vue`)**: Avaluació de la nova targeta de filtres superior (`.evaluation-filter-panel`), capçalera amb títol/icona i acció de neteja, disposició en graella CSS (`.filter-grid`) dels selectors de Grup, Joc i Data, i el seu comportament de filtratge reactiu.
- **Avaluació de qüestionari (`QuizEvaluationView.vue`)**: Avaluació de la integració del panell de filtres ad-hoc en coherència amb el panell de llistat, transició en canviar de joc, neteja de filtres, i convivència amb els blocs col·lapsables de resultats i estadístiques globals.

## Usabilitat
- [OK] **Jerarquia i accions clares a la capçalera (`.filter-panel-header`)**: El títol identificador ("Filtres d'Avaluació" amb icona `pi pi-filter`) aporta context immediat, mentre que el botó d'acció ràpida "Netejar Filtres" (`pi pi-filter-slash`) permet restablir l'estat inicial en un sol clic.
- [OK] **Comoditat d'ús dels selectors**: Tant el selector de grup com el de joc i el selector de data disposen d'opció de neteja individual (`showClear`) i placeholders descriptius internacionalitzats ("Tots els grups", "Tots els jocs" / "Selecciona un joc", "Filtrar per data").
- [OK] **Alineació visual dels controls (`align-items: end`)**: Tots els components d'entrada es troben alineats per la línia de base inferior en la graella, assegurant uniformitat fins i tot amb etiquetes de diferent longitud.

## Consistència visual
- [OK] **Elegància i estètica de targeta (`.filter-panel-card`)**: Ús de fons blanc net, vora subtil (`#e2e8f0`), cantonades suaus (`0.75rem`), encoixinat espaiós (`1.25rem`) i ombra lleugera (`box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05)`), perfectament harmonitzat amb la resta de targetes i taules de la plataforma.
- [OK] **Homogeneïtat entre vistes**: Mateixa estructura de panell, tipografia, separadors i distribució de classes tant a `EvaluationsListView.vue` com a `QuizEvaluationView.vue`.
- [OK] **Internacionalització completa (i18n)**: Ús rigorós de claus `$t('evaluations.filters.*')` per a tots els textos literals en català, castellà i anglès.

## Responsive/mòbil
- [OK] **Disposició en graella CSS flexible (`.filter-grid`)**: L'ús de `grid-template-columns: repeat(auto-fit, minmax(220px, 1fr))` amb `gap: 1rem` garanteix una adaptació fluida des de 3 columnes en monitors d'escriptori, 2 columnes en tauletes, fins a 1 columna apilada en mòbil sense desbordaments ni necessitat de regles rígides.
- [OK] **Amplada i àrees tàctils adequades**: Els controls `Select` i `DatePicker` ocupen el 100% de l'amplada de cada cel·la de la graella (`width: 100%`) i ofereixen una alçada de toc còmoda (> 40px/44px) per a navegació mòbil.

## Feedback a l'usuari
- [OK] **Indicació d'accions actives i neteja**: El botó "Netejar Filtres" i els indicadors `showClear` interns de cada camp ofereixen un retorn visual clar de l'estat del filtre.
- [OK] **Estats buits ben integrats**: En cas de no trobar coincidències amb els filtres aplicats, les taules mostren el missatge/icona d'estat buit corresponent (`empty-state`).

## Accessibilitat bàsica
- [OK] **Mode Fosc (Dark Mode)**: S'han definit estils complets per a `:global(.dark-mode)` i per a `@media (prefers-color-scheme: dark)`, amb fons fosc (`#1e293b`), vores atenuades (`#334155`), etiquetes en to suau (`#94a3b8`) i icones en blau d'alt contrast (`#60a5fa`), complint els nivells de contrast recomanats per la WCAG.
- [OK] **Etiquetatge semàntic de formularis**: Cada control disposa de la seva etiqueta corresponent (`.filter-label`) amb pes visual adequat (`font-weight: 600`) per a fàcil identificació.

## Incidències
*(Sense incidències. El nou panell de filtres amb CSS ad-hoc millora substancialment la claredat, estètica i usabilitat de les vistes d'avaluació).*
