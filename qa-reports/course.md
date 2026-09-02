# QA — Mòdul course (Gestió de Cursos, Unitats i Guió de Classe)

**Veredicte**: APTE  
**Data**: 2026-09-02  

## Compliment funcional
- [OK] **HU-COURSE-01 a HU-COURSE-03 Gestió de Cursos i Matriculació d'Alumnes** — Llistats paginats amb filtres per cerca i estat (`draft`, `active`, `archived`), alta/edició de cursos amb codi únic, i matriculació/desmatriculació en bloc d'alumnes.
- [OK] **HU-COURSE-04 i HU-COURSE-05 Unitats Didàctiques ("Classes")** — Creació, edició i reordenació d'unitats didàctiques dins un curs, amb vinculació N a N reutilitzable amb qüestionaris del mòdul `quiz`.
- [OK] **HU-COURSE-06 i HU-COURSE-07 Guió de Classe (Visor Seqüencial de Blocs)** — Editor de seqüència de blocs (material PDF amb rang de pàgines, qüestionari en directe i pauses amb temporitzador) i reproductor interactiu per al professor (`ScriptPlayerView.vue`).
- [OK] **HU-COURSE-08 Seguiment de Progrés** — Control de l'estat d'unitat per alumne (`pending`, `in_progress`, `completed`).

## Qualitat de codi
- [OK] **Compilació i tests** — Suite de Go (`go test ./...`) i tests frontend/type-check (`pnpm run type-check && pnpm run test`) 100% verds (115 tests unitaris superats).
- [OK] **Seguretat i RBAC** — Filtres de permisos segons el rol (`admin` accés total, `teacher` només els seus cursos, `student` només lectura dels seus cursos matriculats).
- [OK] **Soft-delete i Integritat** — Esborrat lògic per a cursos i unitats (`deleted_at = NOW()`), conservant l'historial acadèmic.

## Homogeneïtat
- [OK] **Estructura de carpetes backend** — Segueix `internal/course` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`, segons la `constitution.md`.
- [OK] **Estructura de carpetes frontend** — Segueix `src/modules/courses` (plural) amb `views/`, `store.ts`, `api.ts`, `types.ts`, segons la `constitution.md`.

## Incidències
Cap incidència detectada. El mòdul compleix totalment amb les especificacions de `contracts/course.openapi.yaml` i `specs/course.md`.
