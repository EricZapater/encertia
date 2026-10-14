# Especificació Tècnica i Funcional: Autorització de Dispositius per a Docents (`device_auth`)

## 1. Visió General i Objectiu

El mòdul **`device_auth`** garanteix que el rol de **Professor (`teacher`)** només pugui accedir a la plataforma Encertia i iniciar sessió des d'ordinadors d'aula o equips físics prèviament autoritzats per un administrador.

### Principis de Disseny i Abast
1. **Només Docents Restringits**:
   - **Alumnes (`student`)**: Accés **100% lliure i obert** des de qualsevol dispositiu web (ordinador personal, mòbil, tauleta) sense cap agent ni restricció.
   - **Administradors (`admin`)**: Accés web universal des de qualsevol navegador mitjançant les seves credencials per permetre la gestió remota i d'emergències.
   - **Professors (`teacher`)**: **Accés restringit obligatòriament**. Només poden iniciar sessió si la petició prové d'un equip autoritzat i assignat al seu compte.
2. **Identificador Hardware Immutable**:
   - Es descarta l'adreça MAC per evitar fallades quan l'equip commuta entre WiFi i cable o quan el sistema operatiu aplica MACs aleatòries de privadesa.
   - S'utilitza exclusivament el **UUID de la Placa Base / BIOS (SMBIOS UUID)**, que és immutable i físicament únic per màquina.
3. **Client Lleuger en Go (`encertia-agent`)**:
   - Binari autònom i multiplataforma (Windows/macOS/Linux) sense dependències externes.
   - Disposa de dos modes: **Mode d'execució normal** (per a l'ús diari del docent) i **Mode registre interactiu per consola** (per a la configuració per part de l'administrador).

---

## 2. Model de Dades (PostgreSQL)

### 2.1 Taula `authorized_devices`

```sql
CREATE TABLE IF NOT EXISTS authorized_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hardware_uuid_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 del UUID de la placa base
    name VARCHAR(100) NOT NULL,                     -- Nom descriptiu (ex: "Portàtil Aula 3B")
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    registered_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_authorized_devices_hw ON authorized_devices (hardware_uuid_hash);
CREATE INDEX IF NOT EXISTS idx_authorized_devices_teacher ON authorized_devices (teacher_id);
```

### 2.2 Taula `device_bootstrap_tokens` (Tokens d'arrencada temporals)

Tokens efímers d'un sol ús utilitzats com a pont de confiança entre l'agent en Go i el navegador web:

```sql
CREATE TABLE IF NOT EXISTS device_bootstrap_tokens (
    token VARCHAR(64) PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES authorized_devices(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_device_bootstrap_lookup ON device_bootstrap_tokens (token, expires_at) WHERE used_at IS NULL;
```

---

## 3. Cicle de Vida i Fluxos de Treball

### 3.1 Flux 1: Registre Semiautomàtic per Consola (`encertia-agent --register`)

Executat pel tècnic o administrador a l'ordinador que es vol autoritzar:

```
[ Administrador / Tècnic ]
       │
       ▼ Executa: `encertia-agent --register`
[ Agent en Go ]
       ├─ Demana: Email de l'administrador
       ├─ Demana: Contrasenya de l'administrador (oculta a la consola)
       ├─ Demana: Email del professor autoritzat a fer servir aquest equip
       ├─ Demana: Nom descriptiu de l'ordinador (ex: "Portàtil Docent Aula 2")
       │
       ├─ Llegeix el UUID físic de la placa base i en calcula el SHA-256.
       │
       ├─ Fa POST /api/v1/auth/login amb credencials d'admin -> Obté JWT d'Admin.
       │
       ├─ Fa POST /api/v1/admin/devices:
       │    Headers: Authorization: Bearer <AdminJWT>
       │    Body: { hardwareUuidHash, name, teacherEmail }
       │
[ Backend Encertia ]
       ├─ Valida que l'usuari sigui 'admin'.
       ├─ Comprova que el professor existeixi i tingui rol 'teacher'.
       ├─ Registra o reactiva el dispositiu a `authorized_devices`.
       └─ Respon 201 Created.
       │
[ Agent en Go ]
       └─ Mostra missatge d'èxit: "✅ Equip autoritzat amb èxit per al professor: Joan Docent"
```

---

### 3.2 Flux 2: Ús Diari del Docent (Doble Clic / Arrencada)

El professor no necessita recordar comandes; només fa doble clic a l'accés directe `encertia-agent` creat a l'escriptori:

```
[ Professor ]
       │
       ▼ Doble clic a `encertia-agent`
[ Agent en Go ]
       ├─ 1. Llegeix el UUID físic de la placa base i calcula el hash SHA-256.
       │
       ├─ 2. Fa crida pública:
       │     POST /api/v1/devices/verify
       │     Body: { hardwareUuidHash }
       │
[ Backend Encertia ]
       ├─ Comprova a `authorized_devices` si el hash existeix i està actiu (`is_active = true`).
       ├─ Genera un token aleatori segur (TTL de 60 segons) a `device_bootstrap_tokens`.
       └─ Respon 200 OK: { bootstrapToken: "xyz...", teacherEmail: "joan@encertia.cat" }
       │
[ Agent en Go ]
       ├─ 3. Construeix la URL de destí:
       │     https://encertia.cat/login?deviceToken=xyz...&email=joan@encertia.cat
       │
       └─ 4. Obre el navegador per defecte del sistema operatiu amb la URL.
       │
[ Navegador Web ]
       ├─ 5. El frontend detecta el `deviceToken` a la query string i el desa en memòria/estat.
       ├─ 6. El professor introdueix la seva contrasenya i fa clic a "Iniciar Sessió".
       ├─ 7. El frontend envia:
       │     POST /api/v1/auth/login
       │     Body: { email, password, deviceToken }
       │
[ Backend Encertia ]
       ├─ Valida usuari i contrasenya.
       ├─ Si l'usuari és `student` o `admin`: permet el login directament (deviceToken opcional).
       ├─ Si l'usuari és `teacher`:
       │    - Comprova que `deviceToken` estigui present.
       │    - Comprova que sigui vàlid, no expirat, no utilitzat i assignat a aquest `teacher_id`.
       │    - Marca el token com a utilitzat (`used_at = NOW()`) i actualitza `last_used_at` de l'equip.
       │    - Si el token és absent o invàlid: retorna `403 FORBIDDEN` (`DEVICE_AUTHORIZATION_REQUIRED`).
       └─ Respon 200 OK amb els tokens JWT de sessió normal.
```

---

## 4. Obtenció de l'Identificador Hardware en Go

L'agent en Go obté el UUID de la placa base de manera nativa per a cada sistema operatiu sense dependències externes feixugues:

```go
// Extracció segons el sistema operatiu
func GetHardwareUUID() (string, error) {
    switch runtime.GOOS {
    case "windows":
        // Execució WMI / PowerShell nativa
        // Get-CimInstance Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID
        return execCommand("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_ComputerSystemProduct).UUID")
    case "darwin":
        // ioreg -d2 -c IOPlatformExpertDevice
        return getMacHardwareUUID()
    case "linux":
        // Lectura directa de DMI (kernel)
        data, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
        if err == nil {
            return strings.TrimSpace(string(data)), nil
        }
        // Fallback a /etc/machine-id
        data, err = os.ReadFile("/etc/machine-id")
        return strings.TrimSpace(string(data)), err
    default:
        return "", errors.New("sistema operatiu no suportat")
    }
}
```

---

## 5. Endpoints de l'API REST (`/api/v1`)

### 5.1 `POST /devices/verify` (Públic / Agent)
- **Objectiu**: Validar l'equip i emetre un `bootstrapToken` de curta durada.
- **Request**:
  ```json
  {
    "hardwareUuidHash": "a1b2c3d4..."
  }
  ```
- **Responses**:
  - `200 OK`:
    ```json
    {
      "bootstrapToken": "dbt_8f9a2b...",
      "teacherEmail": "joan@encertia.cat",
      "deviceName": "Portàtil Aula 3B",
      "expiresInSeconds": 60
    }
    ```
  - `403 Forbidden`: `DEVICE_NOT_AUTHORIZED` (ordinador no registrat o desactivat).

### 5.2 `POST /admin/devices` (Requereix Rol `admin`)
- **Objectiu**: Donar d'alta un equip assignat a un professor.
- **Request**:
  ```json
  {
    "hardwareUuidHash": "a1b2c3d4...",
    "name": "Portàtil Aula 3B",
    "teacherEmail": "joan@encertia.cat"
  }
  ```
- **Responses**:
  - `201 Created`: Equip registrat amb èxit.
  - `400 Bad Request`: Format invàlid o professor no trobat.
  - `403 Forbidden`: L'usuari emissor no és administrador.

### 5.3 `GET /admin/devices` (Requereix Rol `admin`)
- **Objectiu**: Llistar tots els dispositius registrats, estat i darrera connexió.

### 5.4 `DELETE /admin/devices/{id}` (Requereix Rol `admin`)
- **Objectiu**: Revocar/desactivar immediatament un ordinador.

### 5.5 Modificació a `POST /auth/login`
- S'afegeix el camp opcional `deviceToken?: string` al `LoginRequest`.
- **Fase 1 (Desplegament suau sense fricció)**:
  - Si es proporciona `deviceToken`, el backend el valida, el marca com a utilitzat i actualitza `last_used_at` de l'equip a la base de dades.
  - Si el professor inicia sessió directament sense `deviceToken`, se li permet l'accés amb normalitat per evitar disrupcions d'aula mentre es despleguen els agents als ordinadors.
- **Fase 2 (Bloqueig estricte activable)**:
  - Mitjançant configuració / variable d'entorn (`ENFORCE_TEACHER_DEVICE_AUTH=true`), s'activarà el bloqueig estricte (`403 DEVICE_AUTHORIZATION_REQUIRED`) si no es presenta un `deviceToken` vàlid per a comptes `teacher`.

---

## 6. Matriu de Seguretat i Criteris d'Acceptació (Fase 1)

| Escenari | Rol | Des d'on accedeix? | Agent Go? | Resultat (Fase 1) |
| :--- | :--- | :--- | :---: | :--- |
| Alumne responent un quiz | `student` | Mòbil / Ordinador de casa | ❌ No | ✅ **Accés permès** immediat |
| Alumne iniciant sessió | `student` | Qualsevol navegador | ❌ No | ✅ **Accés permès** |
| Admin des de qualsevol lloc | `admin` | Portàtil personal o despatx | ❌ No | ✅ **Accés permès** amb user/password |
| Professor via Agent Go | `teacher` | Ordinador d'aula registrat | ✅ Sí | ✅ **Accés permès** (Dispositiu verificat) |
| Professor sense Agent Go | `teacher` | Qualsevol ordinador | ❌ No | ✅ **Accés permès** (Mode suau Fase 1) |
| Professor amb token d'altre docent | `teacher` | Ordinador assignat a un altre | ⚠️ Token invàlid | ❌ **403 Bloquejat** (Token no correspon a l'usuari) |

