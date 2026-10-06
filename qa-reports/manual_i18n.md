# QA — Mòdul Manual i Traduccions (i18n)

**Veredicte**: APTE
**Data**: 2026-10-06

## Entorn de proves
- Execució de tests unitaris de frontend: SÍ (`pnpm test` executat a `frontend/`, 28 test files passats, 150 tests passats exitosament).
- Compilació de producció de frontend: SÍ (`pnpm build` executat a `frontend/`, built in 1.53s sense errors de TypeScript ni Bundling).
- Verificació de paritat de les claus de traducció (`ca.json`, `es.json`, `en.json`): 100% de coincidència (429 claus a cadascun dels tres fitxers, 0 valors buits o nuls).
- Verificació del manual de professor (`TeacherManualView.vue`): 100% integrat amb `vue-i18n` i estructurat en 10 seccions col·lapsables/navegables.

## Proves funcionals executades
- Execució del suite de unit tests del frontend amb Vitest: 150/150 tests passats exitosament.
- Validació de compilació del frontend amb Vite/Vue-TSC (`pnpm build`): compilat satisfactòriament sense cap error.
- Anàlisi automatitzada de paritat d'estructures JSON (`ca.json`, `es.json`, `en.json`): verificat que les 429 claus existeixen exactament en els 3 idiomes sense omissions.
- Verificació del manual de professor (`TeacherManualView.vue`): confirmat que la vista utilitza `$t()` i `useI18n()` per a tots els textos dinàmics i està degudament enllaçada a les 10 seccions.

## Compliment funcional
- [OK] **Paritat i completitud i18n**: Tots els textos de l'aplicació i del manual estan presents i traduïts en català (`ca.json`), castellà (`es.json`) i anglès (`en.json`).
- [OK] **Documentació de Grups**: Inclosa a la Secció 3 del manual ("Gestió de Grups") així com referenciada a la Secció 9 (selecció de grup en partides) i Secció 10 (filtres per grup).
- [OK] **Documentació de Jugadors anònims**: Inclosa a la Secció 9 ("Partides en Directe i Jugadors Anònims"), explicant la unió via PIN de 6 dígits + Nickname sense requerir registre previ.
- [OK] **Documentació de Tancament automàtic de partida**: Inclosa a la Secció 9, pas 2 ("Control de la Partida i Tancament Automàtic"), detallant que les respostes es tanquen automàticament al vèncer el temps o en respondre tots els jugadors actius.
- [OK] **Documentació de Botons de navegació del host**: Inclosa a la Secció 9, pas 3 ("Controls del Moderador i Navegació"), detallant els botons "Veure Rànquing", "Següent Pregunta" i "Anar a la següent pregunta" des del rànquing.
- [OK] **Documentació de Filtres d'avaluació per Grup/Joc/Data**: Inclosa a la Secció 10 ("Filtres Avançats per Grup, Joc i Data"), detallant com filtrar el rendiment acadèmic.
- [OK] **Documentació de Blocs col·lapsables**: Inclosa a la Secció 10 ("Panells Col·lapsables"), detallant la utilització dels components `Panel :toggleable="true"` per a "Estadístiques Globals per Pregunta" i "Resultats dels Alumnes".

## Qualitat de codi
- [OK] **Build i Types**: `pnpm build` s'executa correctament sense cap fallada de compilació o tipat de Vue/TypeScript.
- [OK] **Tests unitaris**: `pnpm test` passa el 100% de les proves (150 tests).
- [OK] **Estrutura i18n**: Format de claus clar, jeràrquic i homologat entre idiomes (`ca`, `es`, `en`).
- [OK] **Sense claus orfes ni valors buits**: Cap clau no te valors buits o `null`.

## Homogeneïtat
- [OK] **Navegació i disseny**: La vista `TeacherManualView.vue` segueix l'estil visual d'Encertia, utilitza components de PrimeVue (`Card`, `Accordion`, `AccordionPanel`, `Tag`, `Message`) i respecta el patró de navegació i layout de l'aplicació.

## Incidències
Cap incidència detectada. Tot el mòdul de manual i les traduccions compleixen estrictament amb les especificacions.
