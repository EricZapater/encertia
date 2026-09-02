# QA — Mòdul match (session/partida en directe)

**Veredicte**: APTE
**Data**: 2026-09-02

## Compliment funcional
- [OK] Creació de partida i codi PIN — Endpoint `POST /matches` genera un codi PIN únic per partida, codi QR i URL d'unió per als alumnes.
- [OK] Unió d'alumnes a la partida — Endpoint `POST /matches/join` i comprovació `GET /matches/join/{pin}` permeten la unió d'alumnes amb nickname abans i durant la partida.
- [OK] Comunicació en temps real (WebSockets) — Endpoint `/ws/match/{matchId}` gestiona el flux del joc en directe (`JOIN_MATCH`, `START_MATCH`, `NEXT_QUESTION`, `SUBMIT_ANSWER`, `FINISH_MATCH`).
- [OK] Sistema de doble puntuació — Calcula punts de joc per al rànquing en temps real (algorisme amb penalització per temps de resposta) i conserva l'encert acadèmic absolut (sense temps) per al mòdul d'avaluació.
- [OK] Reconnexió i gestió de jugadors — Admet reconnexió de jugadors sense pèrdua d'estat i expulsió de jugadors per part del professor (host).

## Qualitat de codi
- [OK] Compilació i tests — Suite de tests unitaris de Go (`internal/match/hub_test.go`, `service_test.go`, `handler_test.go`) i frontend (`HostGameView.spec.ts`, `PlayerGameView.spec.ts`, `PlayerJoinView.spec.ts`, `wsClient.spec.ts`) 100% verda.
- [OK] Concurrència i seguretat — Maneig de goroutines i mutexes (`sync.RWMutex`) al hub de WebSockets sense ràcies de dades (data races).

## Homogeneïtat
- [OK] Estructura de carpetes backend — Segueix `internal/match` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`, `hub.go`.
- [OK] Estructura de carpetes frontend — Segueix `src/modules/match` amb `views/`, `store.ts`, `api.ts`, `wsClient.ts`, `types.ts`.

## Incidències
Cap incidència detectada. El mòdul compleix les especificacions de `contracts/match.openapi.yaml` i `contracts/match.spec.md`.
