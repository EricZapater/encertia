# QA — Mòdul evaluation

**Veredicte**: APTE
**Data**: 2026-09-02

## Compliment funcional
- [OK] Resum d'avaluacions per qüestionari — Endpoint `GET /evaluations` retorna el resum de partides i alumnes avaluats per qüestionari.
- [OK] Estadístiques detallades de preguntes — Endpoint `GET /evaluations/quizzes/{quizId}` ofereix percentatge d'encert (hit rate), temps mitjà de resposta, distribució de respostes per opció i respostes no contestades.
- [OK] Detall de qualificació per alumne — Endpoint `GET /evaluations/quizzes/{quizId}/students/{studentId}` mostra l'historial de partides i respostes individuals de l'alumne.
- [OK] Ajust manual de nota — Endpoint `PUT /evaluations/{id}/grade` permet al professor modificar o corregir la nota calculada automàticament (`finalGrade`).
- [OK] Integració automàtica amb partides — El servei es registra com a oient (`RegisterFinishedListener`) quan una partida (`match`) finalitza, consolidant automàticament els resultats acadèmics.

## Qualitat de codi
- [OK] Compilació i tests — Backend Go compila i passa `go vet`. Tests frontend de les vistes `EvaluationsListView`, `QuizEvaluationView`, `StudentEvaluationView` en verd.
- [OK] Model de dades i càlculs — Càlculs de mitjanes i percentatges en decimals (`float64`) precisos.

## Homogeneïtat
- [OK] Estructura de carpetes backend — Segueix `internal/evaluation` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`.
- [OK] Estructura de carpetes frontend — Segueix `src/modules/evaluations` (plural) amb `views/`, `store.ts`, `api.ts`, `types.ts`.

## Incidències
Cap incidència detectada. El mòdul compleix les especificacions de `contracts/evaluation.openapi.yaml` i `contracts/evaluation.spec.md`.
