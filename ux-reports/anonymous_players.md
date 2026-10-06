# UX — Mòdul Jugadors Anònims / Partida en Directe (`PlayerJoinView.vue` & `PlayerGameView.vue`)

**Veredicte**: APTE
**Data**: 2026-10-06

## Flux revisats
- **Accés i Unió des de `/play` (Usuari No Autenticat)** — Fluid i immediat. L'usuari no registrat pot introduir un PIN de 6 dígits i un Nickname (o entrar directament mitjançant paràmetre de consulta `?pin=XXXXXX`), enviant la petició d'unió REST i obtenint el `playerToken`.
- **Connexió WebSocket amb `playerToken`** — Establiment automàtic de la connexió WebSocket incloent el paràmetre `playerToken` a la URL. Maneig transparent de la sessió de jugador no registrat emmagatzemada a `localStorage`.
- **Flux de Joc Interactiu a `PlayerGameView.vue`** — Transició d'estats en temps real (Sala d'Espera, Previsualització de Pregunta, Pregunta Activa, Resultat de Pregunta, Rànquing i Pantalla Final/Podi) amb resposta immediata al clic de les opcions de resposta.
- **Reconnexió i Sortida Voluntària** — Si l'usuari recarrega la pàgina a `/play/:pin`, es manté connectat usant el `playerToken` guardat. En prémer "Tornar als Jocs" o "Tornar a l'inici", es neteja la sessió i el WebSocket es tanca de manera ordenada.

## Usabilitat
- [OK] **Formulari d'unió extremadament senzill**: Únicament demana PIN de 6 dígits i Nickname (de 2 a 30 caràcters).
- [OK] **Suport de PIN per Query String**: Permet l'accés directe mitjançant codis QR o enllaços compartits (`/play?pin=123456`), pre-omplint el camp del PIN sense requerir tecleig manual.
- [OK] **Teclat numèric adaptat**: El camp de PIN utilitza `inputmode="numeric"` i `pattern="[0-9]*"`, forçant l'obertura del teclat numèric en dispositius mòbils.
- [OK] **Integració transparent del `playerToken`**: El jugador no necessita compte d'usuari ni contrasenya; el token anònim es genera i utilitza en segon pla per autenticar el canal WebSocket.
- [OK] **Maneig intuïtiu de selecció múltiple vs. selecció única**: A les preguntes de resposta única, la selecció s'envia automàticament per accelerar el ritme de joc; a les de selecció múltiple es requereix la confirmació mitjançant el botó dedicat "Confirmar i Enviar Respostes".

## Consistència visual
- [OK] **Ús coherent de components PrimeVue**: Utilització estàndard de `Card`, `InputText`, `Button`, `Tag`, `ProgressBar` i `Toast`.
- [OK] **Paleta Kahoot-style per a respostes**: Identificació clara de les 4 opcions amb colors vius (Vermell, Blau, Groc, Verd) combinada amb un fons fosc (`#0f172a`) que millora el contrast i redueix la fatiga visual.
- [OK] **Capçalera d'estat persistent**: Mostra en tot moment el Nickname del jugador, el PIN de la sala i la puntuació acumulada amb icona d'estrella.

## Responsive/mòbil
- [OK] **Disseny Mobile-First**: Disseny adaptat a pantalles de telèfon intel·ligent (on habitualment juga l'alumne).
- [OK] **Àrees tàctils àmplies (Target Size)**: Els botons de resposta tenen un `min-height` de 100px i un encoixinat generós (`1.5rem`), superant amb escreix el mínim de 44x44px recomanat per a pantalles tàctils.
- [OK] **Reorganització per a pantalles petites**: Graella adaptativa (`grid-template-columns: 1fr` a `<640px` i `repeat(2, 1fr)` en pantalles amples).

## Feedback a l'usuari
- [OK] **Indicador visual de resposta enviada**: Quan el jugador respon, la pantalla canvia immediatament a la vista "Resposta registrada!", evitant clics duplicats o incertesa.
- [OK] **Comptador de temps visual i numèric**: Compta els segons restants amb una barra de progrés animada (`ProgressBar`) i el temps numèric en gran.
- [OK] **Feedback immediat d'encert/error**: Feedback clar al final de cada pregunta ("Molt bé! Resposta Correcta! +1 punt" en verd vs "Oh no! Resposta Incorrecta +0 punts" en vermell).
- [OK] **Gestió d'expulsió**: Pantalla dedicada (`screen-kicked`) amb missatge explicatiu si el moderador expulsa el jugador de la partida.
- [OK] **Alertes d'error via Toast**: Missatges d'error de validació o de connexió notificats clarament a través de `useToast`.

## Accessibilitat bàsica
- [OK] **Símbols geomètrics a les respostes**: Cada botó de resposta inclou un símbol (triangle, rombe, cercle, quadrat) a més del color, complint els criteris d'accessibilitat per a usuaris amb daltonisme.
- [OK] **Contrast elevat de text**: Text blanc sobre fons fosc slate a la pantalla de joc i fons blau graduat a la pantalla d'unió.
- [OK] **Etiquetes de formulari associades**: Tots els camps d'entrada tenen `id` i `<label>` associat correctament.

## Incidències
1. **[Detall] Textos directes en català sense claus d'i18n**: Els components `PlayerJoinView.vue` i `PlayerGameView.vue` contenen cadenes de text directament en català en lloc de fer servir `$t()` o `useI18n()` amb les claus definides a `src/i18n/locales/*.json` (per exemple, `match.join` i `match.player`). Es recomana refactoritzar les cadenes per suportar el canvi d'idioma de la plataforma (Català / Castella / Anglès).
2. **[Detall] Neteja de `playerToken` caducat a `localStorage`**: Si l'usuari tanca la pestanya sense prémer "Tornar a l'inici", el `playerToken` roman guardat a `localStorage`. Encara que s'amaga correctament quan l'usuari inicia o s'uneix a una nova partida, seria convenient afegir una comprovació de caducitat o neteja de tokens antics.
