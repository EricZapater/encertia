# QA — Flux de Partides en Directe (Match Flow)

**Veredicte**: APTE
**Data**: 2026-10-06

## Entorn de proves
- **Backend**: Compilat i unit tests executats satisfactòriament (`go test -count=1 ./internal/match/...` → **PASS**, tots els tests passats).
- **Frontend**: Compilat i unit tests executats satisfactòriament (`pnpm build` i `pnpm test` a `frontend/` → **PASS**, 27/27 fitxers de test passats, 146/146 unit tests verficats).
- **Especificació i Contracte**: Validat contra `contracts/match.openapi.yaml` i `contracts/match.spec.md`.

## Proves funcionals executades
1. **Tancament automàtic de la pregunta quan tots els jugadors connectats contesten**:
   - Backend (`backend/internal/match/service.go`): Verificat la lògica a `handlePlayerSubmitAnswer`, on un cop enregistrada la resposta es comprova `if connectedCount > 0 && len(answers) >= connectedCount`, transicionant automàticament la partida a l'estat `question_results` (`showQuestionResults`) i emetent l'esdeveniment `match:question_ended` a tots els clients connectats.
   - Test unitari de backend `TestMatchService_AutoCloseQuestion` executat i validat.
2. **Navegació i control de la pantalla de resultats**:
   - Frontend (`frontend/src/modules/match/views/HostGameView.vue`): A la fase `question_results`, el moderador disposa dels dos botons requerits:
     - **"Veure Rànquing"** (`btn-show-leaderboard`), que emet `host:show_leaderboard` per transicionar a la pantalla de rànquing parcial (`leaderboard`).
     - **"Següent Pregunta"** (`btn-next-question`), que emet `host:next_question` per avançar directament a la següent pregunta o finalitzar la partida.
   - Tests unitaris de frontend a `HostGameView.spec.ts` executats i comprovats.
3. **Navegació des de la pantalla de rànquing**:
   - Frontend (`HostGameView.vue`): A la fase `leaderboard`, el moderador disposa del botó **"Anar a la següent pregunta"** (`btn-next-question`), o bé **"Finalitzar i Veure Podi"** (`btn-finish-podium`) si s'ha arribat a la darrera pregunta.
   - Tests unitaris de frontend a `HostGameView.spec.ts` executats i comprovats.

## Compliment funcional
- [OK] **Tancament Automàtic**: L'auto-tancament opera tan bon punt tots els jugadors actius connectats han enviat la seva resposta sense necessitat que s'esgoti el temps limitat.
- [OK] **Navegació del Moderador a Resultats**: Existència i funcionament dels botons *"Veure Rànquing"* i *"Següent Pregunta"*.
- [OK] **Navegació del Moderador a Leaderboard**: Existència i funcionament del botó *"Anar a la següent pregunta"*.
- [OK] **Especificació Tècnica i Contracte**: Sincronització plena amb les regles de negoci 5 i 6 descrites a `contracts/match.spec.md` i l'esquema OpenAPI `contracts/match.openapi.yaml`.

## Qualitat de Codi i Verificació Unitària
- **Backend Unit Tests**: Passats tots els tests (`go test -count=1 ./internal/match/...`).
- **Frontend Unit Tests**: Passats tots els tests unitaris (`pnpm test` amb 146/146 tests en verd).
- **Frontend Build**: Compilació neta i ràpida en 1.64s (`pnpm build`).

## Veredicte Final
**APTE**: El flux de partides en directe compleix tots els requeriments funcionals, regles de negoci i contractes especificats, amb totes les suites de verificació en verd.
