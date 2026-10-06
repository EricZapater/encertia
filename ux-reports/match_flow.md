# UX — Mòdul Partida en Directe (Moderador / `HostGameView.vue`)

**Veredicte**: APTE
**Data**: 2026-10-06

## Flux revisats
- **Sala d'Espera / Lobby (`lobby`)**: Visualització immediata del PIN de la partida en gran, codi QR generat automàticament, enllaç directe (`/play?pin=...`), nom del grup si la partida és per a un grup específic, i recompte de jugadors connectats en temps real. Inclou acció d'expulsió de jugadors no desitjats amb modal de confirmació dialog PrimeVue.
- **Previsualització de Pregunta (`question_preview`)**: Pausa de lectura abans de comenar el temps de la pregunta amb botó clar "Iniciar Temps".
- **Pregunta Activa (`question_active`)**: Pantalla de seguiment en directe amb temporitzador circular en groc, recompte en temps real de respostes rebudes respecte al total de jugadors, graella d'opcions de resposta acolorides amb icones geomètriques i botó d'acció per forçar el tancament ("Tancar i Mostrar Resultats").
- **Resultats de Pregunta (`question_results`)**: Gràfic de barres verticals acolorides amb el percentatge i nombre absolut de vots per opció, ressaltat visual de l'opció correcta (zoom 1.05x i xapa amb icona `pi-check` verd) i barra d'accions del moderador.
- **Rànquing Parcial (`leaderboard`)**: Classificació Top 8 de la partida amb insígnies per posició (Or, Plata, Bronze) i informació de la pregunta actual respecte al total.
- **Podi Final (`finished`)**: Pantalla final de celebració amb podi 3D animat per als 3 primers classificats i botó de sortida de la partida.

## Usabilitat
- [OK] **Disposició i claredat dels botons primaris a la fase de resultats (`question_results`)**:
  - Botó **"Veure Rànquing"** (`severity="primary"`): Botó principal destacat que convida al moderador a ensenyar la taula de classificació parcial per motivar els alumnes.
  - Botó **"Següent Pregunta"** (`severity="secondary"`): Botó secundari ben diferenciat que permet al moderador saltar la fase de rànquing i avançar directament a la següent pregunta si vol mantenir un ritme de classe més àgil.
  - La jerarquia visual entre `primary` i `secondary` és clara, evident i evita confusions en el moment de la projecció.
- [OK] **Botó "Anar a la següent pregunta" a la fase de rànquing (`leaderboard`)**:
  - Botó primari centralitzat (`severity="primary"`) que condueix el flux cap a la següent pregunta.
  - Canvi dinàmic automàtic: Quan la partida arriba a l'última pregunta (`matchStore.isLastQuestion`), el botó canvia de text a **"Finalitzar i Veure Podi"** i de color a verd (`severity="success"` amb icona `pi-trophy`), oferint un senyal d'èxit final molt clar.
- [OK] **Control del moderador a la sala d'espera**: Botó "Començar Partida" en verd (`severity="success"`), deshabilitat si no hi ha cap jugador a la sala (`activePlayersCount === 0`) amb un missatge d'ajuda explicatiu ("Esperant que s'uneixi com a mínim 1 jugador...").
- [OK] **Gestió d'expulsió clara**: Expulsió de jugadors des de la sala d'espera protegida per un modal de confirmació `Dialog` que evita clics no desitjats.

## Consistència visual
- [OK] **Ús coherent de PrimeVue**: Utilització estricta de components PrimeVue (`Button`, `Tag`, `Dialog`) amb configuració de severitat (`primary`, `secondary`, `success`, `warn`, `info`).
- [OK] **Estètica de joc en directe (Kahoot-style)**: Integració harmònica del fons dark slate (`#0b1329`), lletres i xifres en gran contrast, i les 4 formes/colors geomètriques de les opcions de resposta.
- [OK] **Badges contextuals d'informació**: Utilització del component `Tag` per indicar el nom del grup associat tant a l'header superior com al lobby i al podi final.

## Responsive/mòbil
- [OK] **Disseny adaptat a pantalla gran i projectors**: Dissenyat específicament com a pantalla de projecció d'aula (projector o pantalla de moderador).
- [OK] **Adaptabilitat a tauletes**: Disposa de regles adaptatives `@media (max-width: 900px)` que converteixen les graelles de 2 columnes en 1 columna, permetent que el professor pugui gestionar la partida des d'una tauleta tàctil o portàtil.

## Feedback a l'usuari
- [OK] **Comptador de jugadors i respostes en temps real**: Indicador ràpid de respostes rebudes ("X / Y Respostes") durant la fase activa que dona tranquil·litat al moderador sobre si tota la classe ha contestat.
- [OK] **Gràfic de barres animat**: Transició fluida de les barres de resultats (`transition: height 0.6s`) al moment de mostrar quina ha estat la distribució de vots.
- [OK] **Marcatge explícit de la resposta correcta**: La columna correcta s'amplia lleugerament i mostra una bombolla verda amb `pi-check`, fent innecessària qualsevol explicació verbal addicional del professor.

## Accessibilitat bàsica
- [OK] **Dimensions clicables "Hero"**: Els botons principals d'acció (`.btn-action-hero`) tenen una mida de font d'1.35rem i un padding d'1rem 2.5rem, fent-los fàcilment activables des de qualsevol pantalla tàctil.
- [OK] **Identificadors visuals dobles (Colors + Icones)**: Tant les respostes com els botons d'acció combinen iconografia `PrimeIcons` amb textos expressius i colors diferenciats.
- [OK] **PIN en format gegant**: Tipografia de 3.5rem amb un letter-spacing de 0.2em que facilita la lectura a distància des del fons de l'aula.

## Incidències
1. **[Detall] Textos d'interfície directes en català sense claus d'i18n**: Igual que a altres vistes del mòdul match, els etiquetatges de botons ("Veure Rànquing", "Següent Pregunta", "Anar a la següent pregunta") estan escrits directament en català al template de `HostGameView.vue`. Es recomana externalitzar aquests textos al fitxer de traduccions `i18n` per permetre el suport multiidioma.
