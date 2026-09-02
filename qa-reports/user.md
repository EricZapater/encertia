# QA — Mòdul user

**Veredicte**: APTE
**Data**: 2026-09-02

## Compliment funcional
- [OK] HU-USER-01 Llistat d'usuaris amb filtres i paginació — Suporta `page`, `pageSize`, `search`, `role` i `status`. Un professor només pot veure alumnes (`student`), mentre que un administrador pot veure qualsevol rol.
- [OK] HU-USER-02 Creació d'usuari des de l'aplicació — `teacher` només pot crear `student` (rebutja amb 403 si intenta crear `admin` o `teacher`), mentre que `admin` pot crear qualsevol rol.
- [OK] HU-USER-03 Importació massiva d'alumnes (Batch / CSV) — Valida les dades per fila, dona d'alta en bloc i retorna l'informe d'èxits/errors.
- [OK] HU-USER-04 Consulta de detall d'usuari — Funciona per ID validant permisos i comprovant l'estat d'esborrat lògic (soft-delete).
- [OK] HU-USER-05 Modificació d'usuari — Permet l'autoedició de dades bàsiques a usuaris actius, reservant la modificació de `role` i `isActive` exclusivament als usuaris amb rol `admin`.
- [OK] HU-USER-06 Reseteig administratiu de contrasenya — Valida la longitud mínima de contrasenya (8 caràcters) i revoca tots els `refresh_tokens` actius de l'usuari modificat.
- [OK] HU-USER-07 Baixa lògica d'usuari (Soft-Delete) — Executa l'esborrat lògic omplint `deleted_at = NOW()` i revoca sessions actives. No s'efectua cap `DELETE` físic SQL.

## Qualitat de codi
- [OK] Compilació i tests — Backend Go compila net (`go vet` passa) i tests unitaris de service/handler en verd. Frontend TypeScript i tests de Pinia/views en verd.
- [OK] Seguretat i hashing — Les contrasenyes es guarden amb hash `bcrypt`. No s'exposen en respostes JSON (`json:"-"`).
- [OK] Gestió d'errors i validacions — Respostes d'error coherents utilitzant l'esquema `ErrorResponse` i `AppError` del paquet `shared`.

## Homogeneïtat
- [OK] Estructura de carpetes backend — Segueix `internal/user` (singular) amb `handler.go`, `service.go`, `repository.go`, `model.go`.
- [OK] Estructura de carpetes frontend — Segueix `src/modules/users` (plural) amb `views/`, `components/`, `store.ts`, `api.ts`, `types.ts`.
- [KO] Divergència de noms de rol — Discrepància de nomenclatura entre mòduls: `user` utilitza `admin`, `teacher`, `student` (coherent amb `user.openapi.yaml` i `user.spec.md`), mentre que `auth.openapi.yaml` especificava `professor` i `alumne`.

## Incidències
1. **[Menor] Divergència de nomenclatura de rols entre mòduls**:
   - `user.openapi.yaml` i el codi de `user` utilitzen `admin`, `teacher` i `student`.
   - `auth.openapi.yaml` havia definit `professor` i `alumne`.
   - **Recomanació**: Unificar tot el projecte a l'enum `[admin, teacher, student]` en l'especificació d'auth o adaptar user a català.
