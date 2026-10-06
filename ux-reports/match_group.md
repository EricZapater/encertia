# UX — Selecció de Grup en Llançar Partida (`match_group`)

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Flux revisats
- **Accés al llançament de partida des del llistat (`QuizzesListView.vue`)**: Botó "Llançar" visible exclusivament en qüestionaris publicats (`published`). Obre de manera immediata i fluida el modal d'inicialització.
- **Modal de selecció de grup (`LaunchMatchModal.vue`)**: Permet triar entre partida lliure ("Sense grup associat") o un dels grups de la classe ("Nom (Curs)"), amb explicació clara sobre la conservació de dades d'avaluació.
- **Visualització al Lobby del moderador (`HostGameView.vue` - Fase Lobby)**: Identificació visual destacar del grup mitjançant un `Tag` distintiu a la part superior de la capçalera i sobre el panell del codi QR/PIN.
- **Visualització durant el joc i Podi Final (`HostGameView.vue` - Fase Finished)**: El nom del grup es manté visible a la capçalera en totes les preguntes i es mostra prominentment al resum final del podi.

## Usabilitat
- [OK] **Opció clara entre partida lliure o de grup**: El selector `Select` de PrimeVue distingeix clarament la primera opció "Sense grup associat (Partida lliure)" respecte a la llista de grups acadèmics (`Nom (Curs)`).
- [OK] **Text d'ajuda contextual**: La nota explicativa sota el desplegables orienta el professor sobre l'impacte de vincular un grup (desament d'estadístiques i avaluació dels alumnes).
- [OK] **Càrrega automàtica de grups**: Si el store de grups no té les dades carregades, el modal executa `fetchGroups()` en obrir-se sense requerir acció addicional per part de l'usuari.

## Consistència visual
- [OK] **Coherència amb PrimeVue**: Ús correcte dels components `Dialog`, `Select`, `Button` i `Tag` mantenint la línia estilística i la paleta de colors de l'aplicació.
- [Millorable] **Absència de claus i18n al modal**: Els textos de la interfície a `LaunchMatchModal.vue` (títol del diàleg, etiquetes del selector, opció per defecte, missatge d'ajuda i botons) estan escrits directament en català en lloc d'emprar `$t()`.

## Responsive/mòbil
- [OK] **Diàleg modal adaptable**: El modal utilitza una amplada de `:style="{ width: '90vw', maxWidth: '480px' }"`, adaptant-se sense problemes a pantalles petites de mòbil o tauleta.
- [OK] **Disseny responsiu del Lobby del Host**: La vista del moderador reorganitza la graella de 2 columnes a 1 sola columna en pantalles estretes (`@media (max-width: 900px)`).

## Feedback a l'usuari
- [OK] **Estats de càrrega visual**: El desplegable de grups mostra un spinner intern (`:loading="groupStore.isLoading"`) mentre es recuperen els grups, i el botó "Iniciar Partida" inclou `:loading="isLoading"` durant la creació de la partida per evitar clics duplicats.
- [OK] **Gestió d'errors clar**: En cas de fallada en la creació de la partida, es mostra un missatge d'error destacat en vermell (`.error-msg.text-red-500`) a la part inferior del modal.

## Accessibilitat bàsica
- [OK] **Associació `<label>` / `<input>`**: L'etiqueta `<label for="group-select">` està vinculada al desplegable d'opcions.
- [OK] **Icones visuals d'acció**: Botons dotats d'icones comprensibles (`pi-play` per iniciar, `pi-times` per cancel·lar, `pi-users` per al grup).

## Incidències
1. **[Detall] Traducció i18n a `LaunchMatchModal.vue`**: Cal extreure les cadenes de text hardcodegades (`"Iniciar Partida en Directe"`, `"Sense grup associat (Partida lliure)"`, `"Selecciona un grup..."`, `"Si vincules la partida a un grup..."`) als fitxers de localització (`ca.json`, `es.json`, `en.json`) per mantenir el suport multi-idioma de la plataforma.
