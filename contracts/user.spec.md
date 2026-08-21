# Especificació del Mòdul: User (Gestió d'Usuaris)

## 1. Visió General
El mòdul d'usuaris (`user`) gestiona el cicle de vida dels usuaris de la plataforma Encertia, incloent la creació individual des de l'aplicació, la importació massiva d'alumnes (CSV/Batch), la consulta de llistats amb filtres i paginació, l'edició de perfils (pròpia i administrativa) i la baixa lògica (soft-delete).

### Rols d'Usuari a Encertia:
- **`admin`** (Administrador): Superusuari amb accés total. Pot crear, consultar, editar i donar de baixa qualsevol usuari (`admin`, `teacher`, `student`), modificar rols i canviar l'estat actiu/inactiu.
- **`teacher`** (Professor): Gestiona els seus cursos, qüestionaris i alumnes. Des de l'aplicació pot crear alumnes individuals, importar alumnes de forma massiva i consultar la llista d'alumnes.
- **`student`** (Alumne): Usuari final. Pot autoregistrar-se des de la pantalla de login/registre públic o ser donat d'alta per un professor/administrador.

---

## 2. Polítiques de Creació i Registre d'Usuaris

1. **Autoregistre públic (Login / Register)**:
   - Des del formulari públic d'autenticació (`/auth/register`), el rol assignat és exclusivament **`student`**.
   - No és possible autoregistrar-se públicament com a `teacher` ni `admin`.

2. **Creació des de l'aplicació (Backoffice / Panell intern)**:
   - **`admin`**: Pot crear usuaris amb rol `admin`, `teacher` o `student`.
   - **`teacher`**: Pot crear usuaris exclusivament amb rol `student`.
   - L'alta des de l'aplicació permet definir una contrasenya inicial o generar-ne una de provisional.

3. **Importació massiva d'alumnes**:
   - Professors i Administradors poden pujar un fitxer CSV o enviar una llista d'alumnes (`firstName`, `lastName`, `email`, `password` opcional) per donar-los d'alta en bloc.

---

## 3. Històries d'Usuari

### HU-USER-01: Llistat d'Usuaris amb Filtres i Paginació
- **Com a** Administrador o Professor.
- **Vull** visualitzar la llista d'usuaris amb filtres per rol, cerca per nom/email i estat (actiu/inactiu).
- **Per tal de** localitzar usuaris ràpidament i gestionar la base d'usuaris.
- **Criteris d'acceptació**:
  - `admin` pot llistar tots els rols (`admin`, `teacher`, `student`).
  - `teacher` pot llistar els usuaris amb rol `student`.
  - Retorna paginació (`page`, `pageSize`, `totalCount`, `totalPages`).

### HU-USER-02: Creació d'Usuari des de l'Aplicació
- **Com a** Administrador o Professor autenticat.
- **Vull** donar d'alta un nou usuari omplint les seves dades bàsiques i rol permès.
- **Per tal d'** incorporar professors o alumnes al sistema.
- **Criteris d'acceptació**:
  - Si l'actor és `teacher`, només pot crear `student` (rebutja amb `403 Forbidden` si intenta crear `teacher` o `admin`).
  - Si l'actor és `admin`, pot crear `admin`, `teacher` o `student`.
  - El correu electrònic ha de ser únic entre usuaris actius (`409 Conflict` si ja existeix).

### HU-USER-03: Importació Massiva d'Alumnes (CSV)
- **Com a** Professor o Administrador.
- **Vull** carregar un fitxer CSV amb les dades de múltiples alumnes.
- **Per tal de** donar d'alta un grup sencer d'alumnes ràpidament sense fer-ho un a un.
- **Criteris d'acceptació**:
  - Valida el format del CSV i cadascuna de les files.
  - Retorna un resum de la importació: total processats, inserits amb èxit, i llista d'errors amb el número de fila i motiu.

### HU-USER-04: Consulta de Detall d'Usuari
- **Com a** Administrador, Professor o el mateix usuari.
- **Vull** consultar la fitxa detallada d'un usuari per ID.
- **Per tal de** veure les seves dades de perfil, rol i estat.
- **Criteris d'acceptació**:
  - `admin` pot consultar qualsevol usuari.
  - `teacher` pot consultar els alumnes (`student`).
  - Retorna `404 Not Found` si no existeix o està donat de baixa.

### HU-USER-05: Modificació d'Usuari
- **Com a** Usuari autenticat o Administrador.
- **Vull** modificar dades d'usuari (`PUT /users/{id}`).
- **Per tal de** mantenir la informació actualitzada o gestionar estats.
- **Criteris d'acceptació**:
  - **Autoedició (Self-edit)**: Qualsevol usuari actiu pot modificar el seu propi nom, cognoms i email (`id == currentUserId`).
  - **Restricció de rol i estat**: Si un usuari no-admin intenta modificar el camp `role` o `isActive`, la petició és rebutjada amb `403 Forbidden`.
  - **Edició d'administrador**: Només un usuari amb rol `admin` pot canviar el `role` o modificar `isActive` (activar/desactivar) de qualsevol usuari.
  - **Usuaris inactius**: Un usuari amb compte inactiu o donat de baixa no pot modificar el seu perfil; només un `admin` pot reactivar-lo.
  - Retorna `409 Conflict` si el nou email ja està en ús per un altre usuari actiu.

### HU-USER-06: Reseteig Administratiu de Contrasenya
- **Com a** Administrador o Professor (per als seus alumnes).
- **Vull** establir una nova contrasenya per a un usuari.
- **Per tal d'** ajudar usuaris que han perdut l'accés o assignar una clau inicial.
- **Criteris d'acceptació**:
  - Valida requisits mínims de contrasenya (mínim 8 caràcters).
  - Invalida tots els `refresh_tokens` actius d'aquell usuari per forçar re-login.

### HU-USER-07: Baixa Lògica d'Usuari (Soft-Delete)
- **Com a** Administrador.
- **Vull** desactivar/eliminar un usuari de la plataforma.
- **Per tal de** revocar el seu accés sense perdre el seu historial acadèmic ni qüestionaris fets.
- **Criteris d'acceptació**:
  - Aplica soft-delete (`deleted_at = NOW()`).
  - Invalida els tokens d'accés i refresh tokens de l'usuari.
  - Mai s'executa un `DELETE` físic SQL a la base de dades.

---

## 4. Decisions Tècniques i Seguretat
1. **Control d'Accés Basat en Rols (RBAC)**:
   - Endpoint protegit amb middleware d'autorització: `RequireRole("admin")`, `RequireRole("admin", "teacher")`.
2. **Soft-Delete i Integritat**:
   - `users.deleted_at` assegura que les respostes històriques de qüestionaris i dades d'avaluació es mantinguin intactes.
3. **Paginació i Cerca**:
   - Paginació estàndard per query parameters: `page` (1-indexed, defecte 1), `pageSize` (defecte 20, màxim 100), `search` (text pla contra `first_name`, `last_name` o `email`), `role` (filtre per rol), `status` (`active`, `inactive`, `all`).

---

## 5. Estratègia de Proves Unitàries i End-to-End

### Proves Unitàries (Backend & Frontend)

#### Backend (`internal/user`)
1. **Service Tests (`service_test.go`)**:
   - `CreateUser`: Creació d'usuari per `admin` (admin, teacher, student) -> permès.
   - `CreateUser`: Creació de `teacher` o `admin` per part d'un `teacher` -> retorna error `FORBIDDEN`.
   - `CreateUser`: Creació amb email existent -> retorna `EMAIL_ALREADY_EXISTS` (409).
   - `UpdateUser`: Autoedició (nom, cognoms, email) d'un usuari actiu -> èxit (200).
   - `UpdateUser`: Autoedició canviant `role` o `isActive` per usuari no-admin -> retorna `FORBIDDEN` (403).
   - `UpdateUser`: Usuari inactiu intentant autoeditar-se -> retorna `FORBIDDEN` / `UNAUTHORIZED` (403).
   - `UpdateUser`: `admin` modificant `role` i `isActive` de qualsevol usuari -> èxit (200).
   - `BatchCreateUsers`: Validació de lot amb usuaris vàlids, duplicats i dades incorrectes -> còmput correcte de `createdCount` i `failedCount`.
   - `DeleteUser`: Comprovació de soft-delete i revocació de refresh tokens associats.

2. **Handler Tests (`handler_test.go`)**:
   - Respostes HTTP correctes per a cada cas: 200, 201, 400, 401, 403, 404, 409.
   - Validació de paràmetres de paginació (`page`, `pageSize`) i filtres (`role`, `status`, `search`).

#### Frontend (`src/modules/users`)
1. **Store Tests (`store.spec.ts`)**:
   - Carrega de llistats amb paginació i filtres.
   - Accions `createUser`, `updateUser`, `deleteUser`, `batchImport`.
2. **Component Tests (`views/__tests__`)**:
   - Formulari de creació: mostra selector de rol només si l'usuari autenticat és `admin`; si és `teacher` fixa el rol a `student`.
   - Formulari d'edició: camps `role` i `isActive` deshabilitats o ocults per a usuaris no administradors.
   - Component d'importació CSV: parseig i previsualització de fitxer abans de l'enviament.

### Proves End-to-End / Integració

1. **E2E-USER-01: Flux de gestió d'usuaris per Administrador**:
   - Login com a `admin`.
   - Alta de professor i alta d'alumne.
   - Modificació del rol i canvi d'estat a inactiu.
   - Comprovació que l'usuari inactiu no pot iniciar sessió.
   - Reactivació de l'usuari per part de l'admin.

2. **E2E-USER-02: Flux de gestió d'alumnes per Professor**:
   - Login com a `teacher`.
   - Creació d'alumnes individuals i importació massiva via CSV.
   - Intent de creació d'un altre `teacher` -> verificació de bloqueig (403).
   - Consulta i cerca d'alumnes a la taula.

3. **E2E-USER-03: Flux de perfil i autoedició**:
   - Login com a `student`.
   - Modificació del propi nom i correu electrònic.
   - Intent d'enviament directe a l'API modificant `role` a `admin` -> verificació de rebuig (403).

4. **E2E-USER-04: Autoregistre públic exclusiu com a Student**:
   - Registre des de la pàgina pública de registre.
   - Verificació que l'usuari creat té exclusivament rol `student`.
