# Constitution — Encertia (Plataforma Educativa: Quiz + LMS)

Aquest document és la font de veritat per a qualsevol agent que treballi en aquest
projecte. No es pot saltar ni reinterpretar sense actualitzar aquest fitxer primer.

**Nom del projecte**: Encertia. Repo, mòdul Go (`go.mod`) i `package.json` del
frontend han d'usar aquest nom com a base (ex. `github.com/<user>/encertia`).

## 1. Stack tecnològic (no negociable)

### Backend
- **Llenguatge**: Go (versió estable més recent, 1.22+)
- **Framework HTTP**: Gin
- **Base de dades**: PostgreSQL
- **Accés a dades**: SQL pur, sense ORM. Les queries viuen a la capa `repository`
  de cada mòdul, mai embegudes a handlers ni a la capa de servei.
- **Migracions**: eina de migracions versionades (ex. `golang-migrate`), fitxers SQL
  explícits, mai auto-generació d'esquema.
- **Arquitectura: modular per domini (screaming architecture), NO per capes planes.**
  Cada domini de negoci és un paquet propi dins `internal/`, en **singular**
  (convenció estàndard de Go), i conté les seves pròpies capes a dins:
  ```
  encertia/
    backend/
      cmd/api
      internal/
        auth/
          handler.go       -> capa HTTP (Gin), sense lògica de negoci
          service.go        -> lògica de negoci
          repository.go     -> queries SQL d'aquest domini
          model.go            -> structs de domini
        user/
          handler.go
          service.go
          repository.go
          model.go
        course/
          ...
        shared/ (o pkg/)
          -> codi transversal real: middleware, response helpers, errors,
             validació genèrica. Només hi va el que és veritablement compartit
             per ≥2 mòduls, mai un calaix de sastre.
        db/
          -> connexió, pool, migracions (transversal, no és un domini de negoci)
    frontend/
      src/
        ...(veure estructura frontend més avall)
  ```
- **Regla de dependència entre mòduls**: un mòdul pot dependre de `shared`,
  però evitar que un mòdul de domini (ex. `course`) importi directament el
  `repository` d'un altre (ex. `user`). Si un mòdul necessita dades d'un altre,
  ho fa a través del seu `service`, mai saltant-se capes.
- **Nomenclatura de nous mòduls**: singular, nom del domini de negoci, no de la
  taula (`enrollment`, no `enrollments_table`).

### Frontend
- **Framework**: Vue 3 (Composition API per defecte, `<script setup>`)
- **Build tool**: Vite
- **Gestor de paquets**: pnpm (mai npm ni yarn — si un agent genera `package-lock.json`
  o usa `npm install`, és un error a corregir)
- **HTTP client**: Axios, centralitzat en un mòdul `api/client.ts` amb interceptors
- **Gestió d'estat**: Pinia
- **Component library**: PrimeVue
- **Llenguatge**: TypeScript per defecte a tot el frontend
- **Arquitectura: modular per domini, en plural** (convenció habitual en projectes
  Vue de mida mitjana), amb una carpeta `components/` d'àmbit global només per
  a elements veritablement reutilitzables:
  ```
  frontend/
    src/
      modules/
        users/
          views/          -> pàgines/rutes d'aquest mòdul
          components/     -> components NOMÉS usats dins d'aquest mòdul
          store.ts         -> Pinia store del mòdul
          api.ts            -> crides Axios específiques del mòdul
          types.ts           -> tipus TS del domini
        courses/
          ...
        quizzes/
          ...
      components/
        -> components genuïnament compartits entre ≥2 mòduls (botons, taules,
           modals, layout, inputs). Si un component només l'usa un mòdul, viu
           dins d'aquest mòdul, no aquí.
      composables/
        -> lògica reutilitzable no visual (ex. `useDebounce`, `usePagination`)
      router/
      api/
        -> client.ts (Axios amb interceptors), config base
  ```
- **Regla per decidir on viu un component**: si dubtes si un component és
  "compartit" o "del mòdul", la resposta per defecte és que viu al mòdul.
  Es promou a `/components` només quan un segon mòdul el necessita de veritat
  (no per anticipació especulativa).

## 2. Regles d'arquitectura

- Separació estricta handler → service → repository **dins de cada mòdul**. Un
  handler mai fa una query SQL directament.
- Les queries SQL es escriuen explícitament (no query builders dinàmics tipus
  squirrel), amb paràmetres posicionals (`$1, $2...`) per evitar SQL injection.
- Un mòdul nou (backend o frontend) es crea només quan hi ha un domini de negoci
  clar (auth, user, course, quiz...). No crear mòduls per una sola pantalla o
  un sol endpoint si no representen un domini propi.
- Cap lògica de negoci al frontend que dupliqui validacions crítiques del backend
  (el backend és sempre l'última paraula en validació).
- Autenticació: JWT o sessió (a decidir a l'spec de la primera feature d'auth,
  no assumir-ho aquí).

## 3. Convencions de codi

- Go: `gofmt` + `golangci-lint` obligatoris abans de donar per tancada una tasca.
- Vue/TS: ESLint + Prettier configurats al repo; un agent no pot introduir codi
  que trenqui el lint.
- Noms de taules i columnes a Postgres: `snake_case`. Noms de variables Go:
  `camelCase`/`PascalCase` segons convenció estàndard Go.
- Commits atòmics per tasca, missatges descriptius (no "wip" ni "fix stuff").

## 4. Principis del projecte

- **Simplicitat abans que flexibilitat prematura**: no afegir capes d'abstracció
  (interfaces, factories) fins que hi hagi una necessitat real de substituir la
  implementació.
- **El professor (usuari final) no és tècnic**: qualsevol UI ha de ser clara sense
  necessitat d'explicació prèvia.
- **Dades de l'alumne són sensibles**: cap resposta ni resultat es perd mai;
  qualsevol operació destructiva (esborrar sessió, alumne, resposta) requereix
  confirmació explícita al backend (soft-delete per defecte, no DELETE físic
  en taules amb dades acadèmiques).

## 5. Quan un agent ha de generar spec/pla abans de codi

- Tasca simple (fix, ajust visual, canvi de text): pot anar directe a codi.
- Tasca que toca model de dades, afegeix un endpoint nou, o introdueix una
  entitat de domini nova: requereix spec + pla abans d'implementar.
- Tasca ambigua o que afecta més d'un mòdul (ex. "el quiz s'ha de desbloquejar
  quan s'acaba el material"): requereix spec + pla + revisió humana explícita
  abans d'implementar.

### Fitxers d'especificació per agent
Cada rol té el seu propi fitxer `.md` d'instruccions, a més d'aquesta
constitution general. La constitution defineix les regles del *projecte*;
el fitxer d'agent defineix els *límits i l'àmbit* d'aquell rol concret:

```
encertia/
  constitution.md          -> aquest document (regles globals del projecte)
  .agent/
    orchestrator.md         -> rol, què pot i no pot tocar, format d'output
    backend-agent.md          -> àmbit (backend/), convencions específiques,
                                  on llegeix el contracte OpenAPI validat
    frontend-agent.md           -> àmbit (frontend/), convencions específiques,
                                    on llegeix el mateix contracte
    infra-agent.md                -> àmbit (Docker/CI/CD), com verifica build
                                      i desplegament
    qa-agent.md                    -> àmbit (validació transversal), com
                                       avalua compliment funcional, qualitat
                                       i homogeneïtat abans del merge
  backend/
    Dockerfile               -> escrit per l'agent Infra
    ...
  frontend/
    Dockerfile               -> escrit per l'agent Infra
    ...
  docker-compose.yml       -> escrit per l'agent Infra
  .github/
    workflows/                -> escrit per l'agent Infra
  contracts/
    auth.openapi.yaml         -> contracte validat del mòdul auth
    course.openapi.yaml
    ...
  specs/
    auth.md                     -> èpiques i històries d'usuari del mòdul auth
    course.md
    ...
  qa-reports/
    auth.md                     -> informe QA del mòdul auth (APTE/NO APTE)
    ...
```

Cada fitxer d'agent hauria de respondre, com a mínim:
- **Àmbit d'escriptura**: quines carpetes pot modificar (i quines li estan
  explícitament prohibides — ex. el backend agent no toca res dins `frontend/`).
- **Font de veritat**: d'on llegeix el contracte (`contracts/*.openapi.yaml`),
  i que no pot inventar-se camps o endpoints que no hi siguin.
- **Output esperat**: on i com deixa el resultat (branca, estructura de
  commits, si genera un resum per revisió humana).
- **Quan s'atura i pregunta**: casos d'ambigüitat que ha d'escalar a
  l'Orquestrador o a l'humà, en lloc de decidir per si mateix.

L'Orquestrador té el seu propi `.agent/orchestrator.md` amb una restricció
explícita: **àmbit d'escriptura limitat a `contracts/` i `specs/`**,
mai a `backend/` ni `frontend/` directament.

## 6. Control de versions i changelog

- El projecte segueix **Semantic Versioning** (`MAJOR.MINOR.PATCH`). La versió
  viu en un únic fitxer a l'arrel: `VERSION` (text pla, ex. `0.3.0`).
- Hi ha un **únic responsable de pujar de versió i mantenir el changelog:
  l'agent Orquestrador**. Cap picacodis ni l'agent Infra toquen `VERSION` ni
  `CHANGELOG.md` directament — si acaben una tasca que ho justifica, ho
  reporten a l'Orquestrador perquè ell ho registri.
- `CHANGELOG.md` segueix el format [Keep a Changelog](https://keepachangelog.com/):
  seccions `Added`/`Changed`/`Fixed`/`Removed` per versió, amb data.
- **Quan es puja de versió**: en el Checkpoint 2 (merge a la branca principal
  d'un mòdul complet, backend + frontend integrats i validats), no abans.
  No es puja versió per cada commit intermedi d'un picacodis.
- **Criteri de bump**:
  - `PATCH`: correccions sense canvi de contracte ni de comportament visible.
  - `MINOR`: mòdul o funcionalitat nova que no trenca cap contracte existent.
  - `MAJOR`: canvi que trenca un contracte OpenAPI ja validat i en ús (ex.
    canviar la forma d'una resposta que el frontend ja consumeix). Un bump
    `MAJOR` **requereix validació humana explícita abans de fer-se**, igual
    que un canvi de contracte (secció 5, punt 3.5).
- Els tags de Git (`vMAJOR.MINOR.PATCH`) es creen en el mateix moment que es
  puja `VERSION`, per mantenir-los sincronitzats.

## 7. Flux de treball multiagent (orquestrador + picacodis + infra + QA)

Aquest projecte es desenvolupa amb **cinc rols d'agent diferenciats**:

### Rol 1 — Agent Orquestrador
- Responsable de dissenyar el **contracte d'API en OpenAPI** per a cada mòdul,
  a partir de l'spec/històries d'usuari de la feature. El contracte es desa a
  `contracts/<modul>.openapi.yaml`, a l'arrel del projecte (fora de `backend/`
  i `frontend/`, perquè és l'element compartit entre tots dos).
- No escriu codi d'implementació (ni Go ni Vue). El seu únic output és el
  contracte (fitxer `openapi.yaml` o equivalent per mòdul) + un resum llegible
  dels canvis respecte a la versió anterior, si n'hi ha.
- **No pot avançar a la fase d'implementació sense validació humana explícita**
  del contracte. Aquest és un punt de bloqueig obligatori (man-in-the-middle),
  no opcional ni saltable per l'agent.
- Un cop validat (i, si cal, corregit) el contracte, l'Orquestrador és qui
  dispara els dos agents picacodis, passant-los el contracte final com a
  única font de veritat de la interfície entre backend i frontend.
- També és responsable de detectar si un canvi demanat a meitat de feature
  trenca el contracte ja validat — si és així, torna a aturar-se i demana
  revalidació humana abans de continuar.

### Rol 2 — Agent Picacodis Backend
- Treballa exclusivament dins `internal/` (workspace propi).
- Implementa el mòdul seguint el contracte OpenAPI validat, sense reinterpretar-lo
  ni afegir camps/endpoints no contemplats sense passar-ho per l'Orquestrador.
- Branca pròpia: `feat/<modul>-backend`.

### Rol 3 — Agent Picacodis Frontend
- Treballa exclusivament dins `src/` (workspace propi).
- Genera tipus TS, store Pinia, crides Axios i vistes a partir del mateix
  contracte OpenAPI validat (idealment generant els tipus automàticament del
  YAML, per evitar transcripció manual divergent).
- Mentre el backend no estigui llest, pot treballar contra un mock fidel al
  contracte, sense bloquejar-se.
- Branca pròpia: `feat/<modul>-frontend`.

### Rol 4 — Agent Infra
- Responsable de tot el que és desplegament i pipeline: `Dockerfile` de
  `backend/` i de `frontend/`, `docker-compose.yml` a l'arrel, workflows de
  GitHub Actions (`.github/workflows/`), i validació que el build/CI/CD
  funciona de veritat (compila, passa lint/tests, la imatge aixeca i respon).
- No escriu lògica de negoci ni a `backend/internal/` ni a `frontend/src/`.
  Pot llegir aquests directoris (necessita saber com es construeix cada
  projecte: `go build`, `pnpm build`, ports, variables d'entorn) però només
  hi escriu fitxers de configuració de build/desplegament (`Dockerfile`,
  `.dockerignore`), mai codi d'aplicació.
- Treballa amb un contracte propi, més senzill que l'OpenAPI: una llista
  clara de requisits d'entorn (variables, ports, health-check endpoint) que
  demana als picacodis si no els té documentats, en lloc d'assumir-los.
- Branca pròpia: `feat/infra-<descripció>` (ex. `feat/infra-ci-backend`).
- Després de qualsevol canvi de pipeline, ha de **verificar activament**
  (no donar per fet) que el build/deploy funciona: revisar el resultat de
  l'execució de la GitHub Action, no només que el YAML sigui vàlid
  sintàcticament.

### Rol 5 — Agent QA
- Actua **al final del desenvolupament d'un mòdul, abans de la integració**
  (abans del Checkpoint de merge): un cop els agents backend i frontend
  donen per acabada la seva feina, però abans que es fusioni a la branca
  principal.
- Té **visió transversal** de tot el projecte (llegeix `backend/` i
  `frontend/` sencers, no només el mòdul acabat de fer), perquè la seva
  feina inclou comparar-lo amb la resta de mòduls ja existents.
- Tres eixos de validació, obligatoris els tres per a cada mòdul:
  1. **Compliment funcional**: el mòdul fa el que diu `product-functional-spec.md`
     i l'spec/contracte del mòdul concret? Prova cada història d'usuari de
     l'spec contra el comportament real implementat.
  2. **Qualitat de codi**: linting net (`golangci-lint`/ESLint ja haurien
     d'estar nets, però ho torna a comprovar), gestió d'errors consistent,
     absència de codi mort o de debug oblidat, absència de vulnerabilitats
     evidents (injecció SQL, dades sensibles exposades, falta de validació
     d'entrada).
  3. **Homogeneïtat**: el mòdul nou segueix els mateixos patrons,
     convencions de nom i estructura que els mòduls anteriors ja validats?
     Si `auth` fa una cosa d'una manera i `user` la fa diferent sense motiu,
     ho ha de senyalar.
- **No corregeix codi ell mateix.** Si troba una incidència, la documenta i
  l'escala a l'Orquestrador, que decideix si torna al backend/frontend
  agent corresponent per esmenar-la (i llavors l'Agent QA torna a revisar).
- **Àmbit d'escriptura**: només `qa-reports/<modul>.md` (un informe per
  mòdul). Prohibit escriure a `backend/`, `frontend/`, `contracts/` o
  qualsevol `.agent/*.md`.
- Cada informe conclou amb un veredicte explícit: **APTE** (llest per
  merge) o **NO APTE** (llista d'incidències a resoldre, classificades per
  gravetat: bloquejant / important / menor).

### Punt de control humà
- **Checkpoint 1 (obligatori)**: validació del contracte OpenAPI abans que
  l'Orquestrador dispari els picacodis.
- **Checkpoint 2 (obligatori, automàtic per procés)**: un cop backend i
  frontend acaben el mòdul, l'Agent QA valida compliment funcional,
  qualitat i homogeneïtat abans que ningú consideri el mòdul llest. Un
  veredicte "NO APTE" bloqueja l'avanç fins que es resolgui.
- **Checkpoint 3**: revisió humana abans de fer merge de les branques
  `-backend` i `-frontend` a la branca principal (amb l'informe QA "APTE"
  ja disponible com a suport a la decisió). Els agents no fan merge
  automàtic entre ells.
- **Checkpoint 4**: revisió humana abans que l'agent Infra faci el primer
  desplegament real a un entorn compartit (staging o producció) — encara
  que el pipeline hagi verificat que el build passa. Un pipeline verd no
  substitueix la validació humana del primer desplegament d'un canvi
  d'infraestructura amb impacte real (secrets, DNS, base de dades).
- Si un picacodis es troba amb una ambigüitat que el contracte no resol, no
  improvisa: reporta a l'Orquestrador, que ho eleva a l'humà si cal.

## 8. Fora d'abast (per ara)

- Mobile app nativa.
- Multi-tenant (aquest projecte és per un sol professor/curs; no dissenyar per
  SaaS multi-client encara, encara que l'arquitectura no ho impedeixi en el futur).
- Streaming de vídeo autoallotjat (s'assumeix Vimeo/YouTube no llistat, embegut
  via iframe, tret que l'spec digui explícitament el contrari).
