# Roadmap de mòduls — Encertia

Ordre previst d'implementació. L'Orquestrador llegeix aquest document abans
de dissenyar l'spec/contracte de qualsevol mòdul, per tenir context de com
encaixa amb la resta.

## v1 — Nucli mínim (Opció A: Kahoot amb registre)

1. **auth** — login/logout, sessió, protecció per rol (professor/alumne).
   ✅ Implementat i validat (v0.2.0).
2. **user** — CRUD d'usuaris (alta individual + massiva CSV), rols admin/teacher/student.
   ✅ Implementat i validat (v0.3.0).
3. **quiz** — creació de qüestionaris, preguntes d'opció múltiple i respostes Kahoot.
   ✅ Implementat i validat (v0.4.0).
4. **match** (partida en directe) — codi/PIN QR, unió d'alumnes, preguntes
   en directe amb temporitzador, WebSockets per rànquing en temps real.
5. **evaluation** — panell d'avaluació posterior pel professor, exportació
   de resultats.

## v2 — Ampliació a LMS (Opció B)

6. **course** — gestió de curs, unitats didàctiques.
7. **material** — pujada/gestió de documents i vídeo (embegut extern).
8. **enrollment** — vincle alumne↔curs, progrés per unitat.