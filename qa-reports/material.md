# QA — Mòdul material (Gestió de Material Didàctic, Documents i Vídeo)

**Veredicte**: APTE  
**Data**: 2026-09-02  

## Compliment funcional
- [OK] **HU-MAT-01 i HU-MAT-02 Pujada i enllaçat de materials** — Suport per a documents (PDF, DOCX, PPTX fins a 50 MB) i vídeos incrustats de YouTube/Vimeo amb autodetecció del proveïdor.
- [OK] **HU-MAT-03 Actualització transparent** — Permet actualitzar el fitxer d'un material preservant l'ID del material (`material_id`), la taula d'associacions `unit_materials` i la taula de guió de classe `script_blocks`.
- [OK] **HU-MAT-04 i HU-MAT-05 Visor de PDF Integrat i Reproductor** — Visor integrat pàgina a pàgina per a documents PDF (`PdfViewerModal.vue`) i player embed per a vídeos.
- [OK] **HU-MAT-07 i HU-MAT-08 Registre d'Accessos i Mètriques** — Enregistrament automàtic de lectures d'alumnes (`material_views`) i panell d'informe d'accessos per al professor.

## Qualitat de codi
- [OK] **Compilació i tests** — Suite de Go (`go test ./...`) i tests frontend/type-check (`pnpm run type-check && pnpm run test`) 100% verds (129 tests unitaris superats).
- [OK] **Seguretat i RBAC** — Filtres de permisos segons el rol (`admin` i `teacher` gestionen/pugen materials, `student` només veu materials de cursos matriculats).
- [OK] **Soft-delete** — Esborrat lògic per a materials (`deleted_at = NOW()`).

## Homogeneïtat
- [OK] **Estructura de carpetes backend** — Segueix `internal/material` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`, segons la `constitution.md`.
- [OK] **Estructura de carpetes frontend** — Segueix `src/modules/materials` (plural) amb `views/`, `components/`, `store.ts`, `api.ts`, `types.ts`, segons la `constitution.md`.

## Incidències
Cap incidència detectada. El mòdul compleix totalment amb les especificacions de `contracts/material.openapi.yaml` i `specs/material.md`.
