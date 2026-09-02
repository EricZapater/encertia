# QA — Mòdul quiz

**Veredicte**: APTE
**Data**: 2026-09-02

## Compliment funcional
- [OK] Creació i edició de qüestionaris — Editor interactiu que permet gestionar qüestionaris amb metadades (`title`, `description`, `coverImageUrl`, `status`, `tags`).
- [OK] Tipus de preguntes i temporitzador — Suporta `single_choice` i `multiple_choice` amb durades de temporitzador vàlides (5, 10, 20, 30, 60, 90, 120 segons).
- [OK] Validació de respostes — Exigeix entre 2 i 6 opcions de resposta per pregunta. Verifica que `single_choice` tingui exactament 1 resposta correcta i `multiple_choice` almenys 1.
- [OK] Duplicació de qüestionaris — Endpoint `POST /quizzes/{id}/duplicate` suporta el paràmetre `includeAnswers` (defecte `false`), permetent copiar només les preguntes o també les opcions de resposta.
- [OK] Gestió d'imatges — Endpoint `POST /quizzes/upload-image` suporta pujada a Cloudflare R2 amb fallback local, validant mides (màx 5 MB) i formats d'imatge (PNG, JPG, WEBP, GIF).
- [OK] Llistat i filtres — Paginació, cerca per text, filtre per estat (`draft`, `published`, `archived`) i filtre per tags.
- [OK] Baixa lògica — Endpoint `DELETE /quizzes/{id}` aplica soft-delete (`deleted_at = NOW()`).

## Qualitat de codi
- [OK] Compilació i tests — Tests de Go (`internal/quiz`) i unitaris de Vue/Vitest (`QuizEditorView.spec.ts`, `QuizzesListView.spec.ts`, `DuplicateQuizModal.spec.ts`) en verd.
- [OK] Neteja i validació — Validació d'entrada rigorosa en la creació/actualització de preguntes. Sense logs de debug o comentaris oblidats.
- [OK] SQL pur i transaccions — Queries SQL explícites amb paràmetres posicionals `$1, $2...`. Transaccions utilitzades correctament al modificar preguntes/respostes en bloc.

## Homogeneïtat
- [OK] Estructura de carpetes backend — Segueix `internal/quiz` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`.
- [OK] Estructura de carpetes frontend — Segueix `src/modules/quizzes` (plural) amb `views/`, `components/`, `store.ts`, `api.ts`, `types.ts`.
- [OK] Disseny visual — Implementa els colors i formes distintives estil Kahoot a l'editor i visualitzador de preguntes.

## Incidències
Cap incidència detectada. El mòdul compleix totalment amb les especificacions de `contracts/quiz.openapi.yaml` i `contracts/quiz.spec.md`.
