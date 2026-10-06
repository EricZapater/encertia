# QA — Mòdul group

**Veredicte**: APTE  
**Data**: 2026-10-06  

## Entorn de proves
- Backend compilat i verificat: SÍ (`backend/internal/group/` amb handler, service, repository, model i suite de unit tests `service_test.go`).
- Frontend compilat i verificat: SÍ (`pnpm --prefix frontend run build` executat amb resultat 100% verd i zero errors de compilació TypeScript/Vite).
- Base de dades local: l'esquema de dades compleix amb les taules `groups` (amb soft-delete `deleted_at`) i `group_students`.
- Dades de prova generades: no requereix neteja addicional (entorn efímer de comprovació).

## Proves funcionals executades
- `POST /api/groups` (HU-GROUP-01: Creació de grup amb nom i curs acadèmic '26-27' per defecte) → Retorna 201 Created amb l'estructura `GroupResponse` (`{ data: Group }`).
- `GET /api/groups` (HU-GROUP-02: Llistat paginat i filtrat per nom, curs acadèmic i assignatura) → Retorna 200 OK amb l'estructura `GroupListResponse` (`items`, `total`, `page`, `pageSize`, `totalPages`).
- `PUT /api/groups/{id}` (HU-GROUP-03: Actualització de nom/curs/assignatura) → Retorna 200 OK amb la dada actualitzada.
- `DELETE /api/groups/{id}` (HU-GROUP-03: Esborrat lògic de grup) → Executa soft-delete (`deleted_at = NOW()`) i retorna 204 No Content.
- `POST /api/groups/{id}/students` & `DELETE /api/groups/{id}/students/{studentId}` (HU-GROUP-04: Assignació i desassignació d'alumnes) → Permet afegir en bloc UUIDs d'alumnes i treure'ls del grup.

## Compliment funcional
- [OK] **HU-GROUP-01 (Creació de grup)** — Implementada la creació de grups associant-hi el nom, assignatura (`course_id` opcional) i el curs acadèmic (`academic_year`), amb el valor per defecte `"26-27"`.
- [OK] **HU-GROUP-02 (Llistat i filtrat de grups)** — Suport per a paginació, cerca per nom (`search`) i filtre per curs acadèmic (`academicYear`) o curs/assignatura (`courseId`). Restricció per rol: els professors veuen els seus grups, els alumnes veuen els grups on estan matriculats i els admins veuen tots els grups.
- [OK] **HU-GROUP-03 (Edició i esborrat lògic de grup)** — Permet modificar la informació del grup. L'esborrat s'executa com a soft-delete (`deleted_at`), preservant l'històric d'activitat acadèmica.
- [OK] **HU-GROUP-04 (Assignació i desassignació d'alumnes)** — Endpoints `/groups/{id}/students` per a afegir alumnes (individualment o en bloc via array d'UUIDs) i eliminar-los de la taula de relació `group_students`.

## Qualitat de codi
- [OK] **Compilació i build frontend** — Execució de `pnpm --prefix frontend run build` totalment neta, generant el bundle a `dist/` sense cap error de tipatge ni sintaxi.
- [OK] **Tests unitaris backend** — Suite `service_test.go` amb mocks de repositori que valida la creació, validació de nom obligatori, actualització, soft-delete i gestió d'alumnes.
- [OK] **Seguretat i SQL** — Separació estricta Handler -> Service -> Repository. Les consultes SQL fan servir paràmetres posicionals (`$1, $2...`) evitant injeccions SQL. Filtres per rol implementats a la capa de servei/repositori.
- [OK] **Validació d'entrada i gestió d'errors** — Retorn d'errors normalitzats utilitzant `shared.AppError` (`shared.RespondWithError`) i validació de valors nuls o UUIDs malformats.

## Homogeneïtat
- [OK] **Estructura de carpetes backend** — Ubicat a `backend/internal/group/` (singular), seguint l'arquitectura screaming architecture de la `constitution.md` amb `handler.go`, `service.go`, `repository.go`, `model.go`.
- [OK] **Estructura de carpetes frontend** — Ubicat a `frontend/src/modules/groups/` (plural), estructurat amb `views/`, `types.ts`, `api.ts`, `store.ts` (Pinia) segons la `constitution.md`.
- [OK] **Fidelitat al contracte OpenAPI** — Compliment 100% de les rutes, esquemes de petició/resposta i codis d'estat de `contracts/group.openapi.yaml`.

## Incidències
Cap incidència detectada. El mòdul `group` compleix estrictament amb les especificacions funcionals, el contracte OpenAPI i les regles de la constitution.
