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

1. **auth** — login/logout, sessió, protecció per rol (professor/alumne).
   ✅ Spec i contracte validats (`specs/auth.md`,
   `contracts/auth.openapi.yaml`). Implementació en curs a Antigravity
   (nota: la implementació real ha optat per JWT en lloc de sessió opaca;
   pendent de revisió QA quan es doni per acabat el mòdul).
2. **user** — CRUD d'alumnes (alta individual + massiva CSV), gestió de
   professor. Depèn de `auth` (taula `users` compartida).
3. **quiz** — creació de qüestionaris, preguntes d'opció múltiple,
   veritable/fals i ordenació. Banc de preguntes reutilitzable.
4. **course** — gestió de curs, unitats didàctiques ("unitat" i "classe"
   són el mateix concepte). Vinculació N a N entre unitats i qüestionaris.
   Inclou el **guió de classe** (visor seqüencial de blocs: material,
   qüestionari, pausa/torn de preguntes) — veure `product-functional-spec.md`
   secció 3.6.
5. **material** — pujada/gestió de documents i vídeo (embegut extern), amb
   visor de PDF integrat necessari pel guió de classe del mòdul `course`.
6. **session** (partida en directe) — codi/PIN, unió d'alumnes, preguntes
   en directe amb temporitzador, WebSockets per rànquing en temps real.
   *Nota: "session" aquí és una partida de joc, no confondre amb la taula
   `sessions` d'autenticació del mòdul auth — si genera confusió, valorar
   renombrar aquest mòdul a `match` o `game` abans d'implementar-lo.*
   Inclou el **sistema de doble puntuació** (punts de joc amb temps vs.
   nota d'avaluació sense temps) — veure `product-functional-spec.md`
   secció 3.8.
7. **evaluation** — panell d'avaluació posterior pel professor, exportació
   de resultats, nota consolidada, i l'opció d'excloure les dues pitjors
   notes en calcular mitjanes (a definir l'abast exacte: per curs sencer o
   per unitat — veure `product-functional-spec.md` secció 3.9).

## Fora de tot abast per ara
- Multi-tenant / multi-professor a gran escala.
- Notificacions per email.
- Recuperació de contrasenya (veure `auth`, fora d'abast v1).

