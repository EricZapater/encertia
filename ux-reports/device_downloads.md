# UX — Mòdul Gestió de Dispositius i Descàrrega d'Agents (device_downloads)

**Veredicte**: APTE
**Data**: 2026-10-14

## Flux revisats
- **Descàrrega d'Agents i Certificats (`DevicesManagementView.vue`)** — Flux àgil i ben estructurat. Quatre targetes visuals per a Windows, macOS, Linux i Certificat X.509, amb identificació d'arquitectura, instruccions de consola i botó de descàrrega directe amb feedback d'espera i notificació d'èxit.
- **Còpia ràpida de comandes de consola** — Copiat al porta-retalls amb un sol clic des de cada targeta de plataforma, mostrant la comanda exacta i confirmació Toast emergent.
- **Consulta i jerarquia de la taula de dispositius autoritzats** — Visualització clara i ordenada dels equips registrats, docent assignat amb correu, estat actiu/revocat mitjançant `Tag` semàntic de PrimeVue i registre temporal d'activitat.
- **Flux de revocació d'un dispositiu autoritzat** — Acció segura mitjançant botó de perill amb icona `pi pi-ban` que obre un modal de confirmació amb context complet de l'equip i explicació clara de les conseqüències abans d'executar l'acció.
- **Navegació i control d'accés des d'`AppNavbar.vue`** — Enllaç "Dispositius" visible exclusivament per a usuaris amb rol d'administrador (`authStore.isAdmin`), mantenint l'estat actiu de la ruta i amb icona representativa `pi pi-desktop`.
- **Commutació dinàmica d'idiomes (CA, ES, EN)** — Canvi fluid entre idiomes a la barra de navegació, amb paritat total de totes les cadenes del panell.

## Usabilitat
- [OK] **Targetes d'agent clares i orientades a l'acció**: Cada targeta comunica d'immediat la plataforma (Windows, macOS, Linux, Certificat), l'arquitectura compatible (ex. `macOS (Universal)`, `Windows (x86_64)`) i el propòsit d'ús.
- [OK] **Instruccions de consola preparades**: La comanda necessària per registrar o atorgar permisos (`.\encertia-agent-windows.exe --register`, `chmod +x`, `certutil`) està destacada en un bloc de codi monoespaiat amb selecció completa en fer clic (`user-select: all`).
- [OK] **Prevenció d'errors destructius en la revocació**: El modal de confirmació no és genèric; indica el nom de l'ordinador afectat i el docent assignat en negreta, advertint clarament que l'equip deixarà de permetre l'accés d'aula immediatament.
- [OK] **Protecció visual d'equips ja revocats**: La columna d'accions només ofereix el botó de revocació per a dispositius actius (`v-if="data.isActive"`). Els equips revocats mostren una etiqueta atenuada, evitant clics inútils.
- [Millorable] **Feedback visual al botó de copiar comanda**: En prémer el botó de copiar, es mostra un missatge Toast però la icona del botó (`pi pi-copy`) es manté inalterada. Canviar temporalment la icona a `pi pi-check` durant 1-2 segons oferiria un reforç visual immediat al punt d'interacció.
- [Millorable] **Guia a l'estat buit de la taula**: Quan encara no hi ha dispositius donats d'alta, la taula mostra "No hi ha cap dispositiu registrat." Estaria bé afegir-hi una breu indicació o recordatori: "Executa l'assistent `encertia-agent --register` des de l'equip d'aula per donar d'alta el primer ordinador."

## Consistència visual
- [OK] **Ús canònic de components PrimeVue**: S'integren `DataTable`, `Column`, `Button`, `Tag`, `Dialog` i `Toast` respectant el sistema de disseny establert al projecte.
- [OK] **Identitat gràfica diferenciada per plataforma**: Cada targeta de descàrrega disposa d'un distintiu d'icona amb colors suaus acordats (blau per a Windows, pissarra per a macOS, ambre per a Linux i verd maragda per al certificat públic).
- [OK] **Jerarquia tipogràfica harmonitzada**: Títols de pàgina (`page-title`), subtítols atenuats (`#64748b`), títols de secció amb icones (`pi pi-download`, `pi pi-shield`) i textos de taula amb diferenciació nítida entre nom principal i dades secundàries (email, data).
- [OK] **Coherència en els botons d'acció**: Botons de descàrrega primaris amb ample complet, botons d'actualització secundaris contornejats (`outlined`) i botons de revocació en perill suau (`severity="danger" text rounded`).

## Responsive/mòbil
- [OK] **Graella adaptable (`cards-grid`)**: Utilitza `repeat(auto-fit, minmax(280px, 1fr))` adaptant-se automàticament des de pantalles d'escriptori amples (4 columnes) fins a tauletes (2 columnes) i mòbils (1 columna a partir de `768px`).
- [OK] **Desplaçament protegit a la taula**: La `DataTable` utilitza `responsive-layout="scroll"` amb amplades mínimes per columna, evitant solapaments de columnes o trencament de dades en tauletes.
- [OK] **Navegació mòbil a `AppNavbar.vue`**: En resolucions inferiors a 768px, les etiquetes de text s'amaguen mentre que les icones d'acció (`pi pi-desktop`, etc.) romanen accessibles, preservant l'espai útil de navegació.

## Feedback a l'usuari
- [OK] **Estat de càrrega i spinner a la descàrrega**: Quan es prem "Descarregar", el botó entra en estat de càrrega (`:loading="downloadingPlatform === item.platform"`), canvia la icona a un spinner animat (`pi pi-spin pi-spinner`) i bloqueja múltiples pulsacions simultànies.
- [OK] **Notificació Toast d'èxit de descàrrega**: En finalitzar la preparació del blob de fitxer, es dispara una notificació d'èxit (`downloadSuccessTitle`) mostrant el nom exacte del fitxer descarregat.
- [OK] **Gestió d'errors de descàrrega i llistat**: En cas de fallada en la connexió o autorització de l'endpoint, es notifica l'administrador amb un Toast d'error detallat (`downloadErrorTitle`).
- [OK] **Estat d'execució en la revocació**: El botó de confirmació al modal mostra un spinner actiu (`:loading="isRevoking"`) mentre dura la crida `DELETE /admin/devices/{id}`. Un cop finalitzada, es mostra un Toast d'èxit i la taula es refresca automàticament.

## Accessibilitat bàsica
- [OK] **Suport de tema fosc (`.dark-mode`)**: Implementació completa d'estils amb selectors `:global(.dark-mode)` que adapten fons de targetes (`#1e293b`), vores (`#334155`), blocs de codi (`#0f172a`), títols (`#f8fafc`) i textos secundaris (`#94a3b8`) garantint un contrast òptim segons criteris WCAG AA.
- [OK] **Contrast i llegibilitat**: Colors de text de gran contrast sobre fons blancs o foscos, amb codi monoespaiat fàcilment distingible.
- [OK] **Interacció modal accessible**: El component `Dialog` s'utilitza amb l'atribut `modal`, proporcionant enfosquiment de fons, retenció de focus de teclat i possibilitat de tancament via botó o tecla d'escapament.
- [OK] **Descripcions accessibles (`title`)**: Els botons d'icona (copiar, revocar) disposen d'atributs `title` traduïts per a suport de lectors de pantalla i tooltips natius del navegador.

## Paritat i18n
- [OK] **Paritat total (100%) entre CA, ES i EN**: Totes les cadenes de text del panell i de la navegació estan implementades a `ca.json`, `es.json` i `en.json`.

| Clau i18n | Català (CA) | Castellà (ES) | Anglès (EN) |
| :--- | :--- | :--- | :--- |
| `nav.devices` | Dispositius | Dispositivos | Devices |
| `devices.title` | Gestió de Dispositius & Agents | Gestión de Dispositivos & Agentes | Device & Agent Management |
| `devices.subtitle` | Autorització d'equips físics d'aula per a docents i descàrrega d'agents d'escriptori. | Autorización de equipos físicos de aula para docentes y descarga de agentes de escritorio. | Classroom physical computer authorization for teachers and desktop agent downloads. |
| `devices.downloads.title` | Descàrrega d'Agents i Certificats | Descarga de Agentes y Certificados | Agent & Certificate Downloads |
| `devices.downloads.subtitle` | Descarrega els binaris d'arrencada per als ordinadors d'aula o el certificat públic de signatura. | Descarga los binarios de inicio para los ordenadores de aula o el certificado público de firma. | Download startup binaries for classroom computers or public signing certificate. |
| `devices.cards.windows.title` | Windows | Windows | Windows |
| `devices.cards.windows.instruction` | Assistent de registre per consola: | Asistente de registro por consola: | Registration wizard via console: |
| `devices.cards.windows.button` | Descarregar per a Windows (.exe) | Descargar para Windows (.exe) | Download for Windows (.exe) |
| `devices.cards.darwin.title` | macOS | macOS | macOS |
| `devices.cards.darwin.instruction` | Permisos d'execució necessaris: | Permisos de ejecución necesarios: | Required execute permissions: |
| `devices.cards.darwin.button` | Descarregar per a macOS | Descargar para macOS | Download for macOS |
| `devices.cards.linux.title` | Linux | Linux | Linux |
| `devices.cards.linux.instruction` | Permisos d'execució necessaris: | Permisos de ejecución necesarios: | Required execute permissions: |
| `devices.cards.linux.button` | Descarregar per a Linux | Descargar para Linux | Download for Linux |
| `devices.cards.cert.title` | Certificat Públic | Certificado Público | Public Certificate |
| `devices.cards.cert.instruction` | Instal·lació com a Root Publisher (PowerShell admin): | Instalación como Root Publisher (PowerShell admin): | Installation as Root Publisher (PowerShell admin): |
| `devices.cards.cert.button` | Descarregar Certificat (.cer) | Descargar Certificado (.cer) | Download Certificate (.cer) |
| `devices.table.title` | Dispositius Autoritzats | Dispositivos Autorizados | Authorized Devices |
| `devices.table.name` | Nom de l'ordinador | Nombre del ordenador | Computer name |
| `devices.table.teacher` | Professor assignat | Profesor asignado | Assigned teacher |
| `devices.table.status` | Estat | Estado | Status |
| `devices.table.lastActivity` | Darrera activitat | Última actividad | Last activity |
| `devices.table.registeredAt` | Data de registre | Fecha de registro | Registration date |
| `devices.table.actions` | Accions | Acciones | Actions |
| `devices.status.active` | Actiu | Activo | Active |
| `devices.status.revoked` | Revocat | Revocado | Revoked |
| `devices.never` | Mai | Nunca | Never |
| `devices.noDevices` | No hi ha cap dispositiu registrat. | No hay ningún dispositivo registrado. | No devices registered. |
| `devices.revokeTooltip` | Revocar dispositiu | Revocar dispositivo | Revoke device |
| `devices.copiedTitle` | Comanda copiada | Comando copiado | Command copied |
| `devices.downloadSuccessTitle` | Descàrrega iniciada | Descarga iniciada | Download started |
| `devices.downloadErrorTitle` | Error de descàrrega | Error de descarga | Download error |
| `devices.downloadError` | No s'ha pogut descarregar el fitxer. | No se ha podido descargar el archivo. | Could not download file. |
| `devices.revokeModal.title` | Confirmar Revocació de Dispositiu | Confirmar Revocación de Dispositivo | Confirm Device Revocation |
| `devices.revokeModal.confirmText` | Estàs segur que vols revocar l'autorització del dispositiu | ¿Estás seguro de que deseas revocar la autorización del dispositivo | Are you sure you want to revoke authorization for device |
| `devices.revokeModal.hint` | Aquest equip deixarà de permetre l'accés del docent immediatament. L'equip es podrà tornar a registrar més endavant. | Este equipo dejará de permitir el acceso del docente inmediatamente. El equipo se podrá registrar de nuevo más adelante. | This computer will immediately prevent teacher login. It can be registered again later. |
| `devices.revokeModal.cancel` | Cancel·lar | Cancelar | Cancel |
| `devices.revokeModal.confirm` | Revocar Dispositiu | Revocar Dispositivo | Revoke Device |
| `devices.revokeSuccess` | Dispositiu revocat satisfactòriament. | Dispositivo revocado satisfactoriamente. | Device revoked successfully. |
| `devices.revokeError` | Error en revocar el dispositiu. | Error al revocar el dispositivo. | Error revoking device. |

## Incidències
1. **[Detall] Estat visual temporal al botó de còpia (`DevicesManagementView.vue`)**:
   - *Flux*: Clic sobre el botó de còpia d'una comanda de consola (`copyCommand`).
   - *Impacte*: Tot i que el Toast emergent informa de la còpia, el botó manté la icona estàtica `pi pi-copy`.
   - *Suggeriment*: Canviar temporalment l'estat local o icona del botó a `pi pi-check` (icona de verificació) durant 1,5 segons per donar retroacció directa i contextual sobre el punt d'interacció.
2. **[Detall] Gestió d'errors silenciada a l'API del porta-retalls (`DevicesManagementView.vue`)**:
   - *Flux*: Execució de `copyCommand` en contextos sense suport del porta-retalls (`http://` sense context segur o denegació de permisos al navegador).
   - *Impacte*: El bloc `catch` és silenciós (`// Silent`), i l'usuari no rep cap feedback si la còpia falla.
   - *Suggeriment*: Mostrar una notificació Toast d'avís o error indicant que no s'ha pogut accedir al porta-retalls per advertir l'usuari que seleccioni el text manualment.
3. **[Detall] Crida a l'acció informativa a l'estat buit de la taula (`DevicesManagementView.vue`)**:
   - *Flux*: Panell acabat d'instal·lar o sense dispositius registrats.
   - *Impacte*: L'estat buit mostra "No hi ha cap dispositiu registrat", sense enllaçar visualment amb la secció superior de descàrrega d'agents ni indicar com procedir.
   - *Suggeriment*: Afegir un text secundari a l'estat buit indicant que cal descarregar l'agent i executar `encertia-agent --register` a l'equip d'aula per afegir-hi el primer dispositiu.
