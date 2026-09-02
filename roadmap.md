# Roadmap de mòduls — Encertia

Ordre previst d'implementació. L'Orquestrador llegeix aquest document abans
de dissenyar l'spec/contracte de qualsevol mòdul, per tenir context de com
encaixa amb la resta.

**Procés per a cada mòdul** (veure `constitution.md`, secció 7): spec +
contracte → validació humana (Checkpoint 1) → backend + frontend →
**validació QA (Checkpoint 2)** → merge (Checkpoint 3) → Infra
build/deploy (Checkpoint 4 si és el primer desplegament real). Un mòdul no
es considera tancat fins que l'informe QA a `qa-reports/<modul>.md` diu
APTE.

## v1 — Nucli mínim (Opció A: Kahoot amb registre)

1. **auth** — login/logout, sessió JWT amb taula de revocació de tokens, protecció per rol (`admin`, `teacher`, `student`).
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/auth.md`).
2. **user** — CRUD d'alumnes (alta individual + massiva CSV), gestió de professor i admin, soft-delete, reset de clau.
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/user.md`).
3. **quiz** — creació de qüestionaris, preguntes d'opció múltiple/única, editor Kahoot-style, Cloudflare R2 i duplicació.
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/quiz.md`).
4. **course** — gestió de curs, unitats didàctiques ("unitat" i "classe" són el mateix concepte), matriculacions d'alumnes, vinculació N a N entre unitats i qüestionaris, i **guió de classe** (visor seqüencial de blocs: material, qüestionari, pausa/preguntes).
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/course.md`).
5. **material** — pujada/gestió de documents (PDF/DOCX/PPTX) i vídeo (embegut extern YouTube/Vimeo), amb visor de PDF integrat per al guió de classe i registre d'accessos d'alumnes.
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/material.md`).
6. **session / match** (partida en directe) — codi/PIN, unió d'alumnes, preguntes en directe amb temporitzador, WebSockets per rànquing en temps real, podi 3D i doble puntuació.
   ✅ Spec, contracte i implementació completats (`match`). QA: **APTE** (`qa-reports/match.md`).
7. **evaluation** — panell d'avaluació posterior pel professor, estadístiques per pregunta, qualificació automàtica i ajust manual.
   ✅ Spec, contracte i implementació completats. QA: **APTE** (`qa-reports/evaluation.md`).

## Fora de tot abast per ara
- Multi-tenant / multi-professor a gran escala.
- Notificacions per email.
- Recuperació de contrasenya (veure `auth`, fora d'abast v1).

