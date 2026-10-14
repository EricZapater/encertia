# UX — Mòdul Autorització de Dispositius (device_auth)

**Veredicte**: APTE
**Data**: 2026-10-14

## Flux revisats
- **Arrencada i verificació des de l'Agent CLI (`encertia-agent`)** — Flux fluid, amb missatges clars per consola, detecció de maquinari transparent per al professor i obertura automàtica del navegador amb el token temporal i l'email del docent.
- **Assistent interactiu de registre (`encertia-agent --register`)** — Assistent guiat en 4 passos per a administradors, amb sol·licitud clara de credencials, pre-càlcul del nom de l'equip a partir del hostname del sistema i resum final d'èxit.
- **Pàgina d'inici de sessió amb dispositiu autoritzat (`/login?deviceToken=...&email=...`)** — Presència del distintiu de confiança PrimeVue (`Tag` verd amb icona `pi pi-shield` "Dispositiu autoritzat i verificat"), pre-ompliment automàtic de l'email i camp de contrasenya llest per introduir.
- **Pàgina d'inici de sessió estàndard (sense token)** — Comportament transparent i sense friccions per a alumnes i administradors, sense mostrar distintius innecessaris.
- **Canvi d'idioma dinàmic a la vista de Login (CA, ES, EN)** — Paritat idiomàtica completa del missatge d'autorització i dels formularis, amb commutació immediata des de la capçalera de la targeta.

## Usabilitat
- [OK] **Fricció d'aula mínima per al professor**: El docent només fa doble clic a l'accés directe de l'agent. No ha d'escriure comandes, ni recordar el seu correu a la pantalla d'inici (ja ve pre-omplert); només posa la contrasenya i accedeix directament.
- [OK] **Assistent de registre senzill per al tècnic/admin**: El registre interactiu (`--register`) numera els passos de l'1 al 4, oferint un valor per defecte intel·ligent per al nom de l'equip (`[Ordinador (<hostname>)]`).
- [OK] **Missatges d'error explicatius a la consola**: Si un equip no està autoritzat, la consola no es tanca immediatament: explica que l'equip no té accés, com contactar amb l'administrador o registrar-lo (`encertia-agent --register`) i sol·licita prémer Enter per tancar la finestra (`Prem Enter per sortir...`), evitant el tancament sobtat en dobles clics a Windows o macOS.
- [Millorable] **Ocultació efectiva de contrasenya a la consola**: `readPassword()` llegeix la contrasenya mitjançant un buffer estàndard sense desactivar l'eco del terminal (manca integració de `golang.org/x/term` o `termios`), fent que els caràcters siguin visibles si s'escriu per consola oberta.
- [Millorable] **Focalització automàtica del cursor**: Quan la URL porta l'email pre-omplert, seria convenient posicionar el focus automàticament (`autofocus`) al camp de contrasenya per agilitzar encara més l'inici de sessió.

## Consistència visual
- [OK] **Ús canònic de PrimeVue**: S'utilitzen exclusivament components oficials de PrimeVue (`Card`, `Tag`, `InputText`, `Password`, `Button`, `Toast`) amb estils alineats amb la línia gràfica corporativa d'Encertia.
- [OK] **Distintiu de confiança visual (`Tag` verd + escut)**: El distintiu `Tag` amb `severity="success"`, icona `pi pi-shield` i text "Dispositiu autoritzat i verificat" transmet una sensació clara de seguretat, professionalitat i certesa.
- [OK] **Camp de contrasenya amigable**: El component `Password` incorpora `toggleMask`, permetent al docent fer visible la contrasenya per verificar possibles errors de tecleig abans d'enviar.

## Responsive/mòbil
- [OK] **Adaptabilitat de pantalla**: El contenidor utilitza `max-width: 440px` amb `width: 100%` i padding de seguretat `1.5rem`, adaptant-se sense trencaments ni desbordaments des de pantalles petites (375px) fins a grans pantalles d'escriptori.
- [OK] **Àrees tàctils**: Els inputs i el botó d'inici de sessió mantenen una alçada i separació superiors a 44px, facilitant la interacció tàctil en tauletes o pantalles tàctils d'aula.

## Feedback a l'usuari
- [OK] **Estat de càrrega al formulari**: Durant l'autenticació, el botó mostra un spinner de PrimeVue (`:loading="isSubmitting"`) i els camps queden deshabilitats per prevenir dobles enviaments.
- [OK] **Indicadors de progrés a la CLI**: Durant les crides de xarxa a l'assistent de registre, es mostren missatges d'espera (`⏳ Autenticant administrador...`, `⏳ Registrant equip a la plataforma...`).
- [OK] **Notificacions Toast davant d'errors**: Qualsevol error en la credencial o comunicació amb l'API es presenta mitjançant un Toast flotant (`severity: 'error'`) durant 4 segons amb explicació del motiu.
- [OK] **Fallback si falla l'obertura del navegador**: Si el sistema operatiu no pot obrir el navegador automàticament, la CLI mostra un missatge d'avís (`⚠️`) acompanyat de la URL completa perquè el docent pugui copiar-la i enganxar-la manualment.

## Accessibilitat bàsica
- [OK] **Contrast i llegibilitat**: Colors d'alt contrast (blau fosc `#1e3a8a` sobre fons clar, text `#0f172a` i verd PrimeVue d'èxit accessible).
- [OK] **Semàntica HTML**: Etiquetes `<label for="...">` associades correctament amb els atributs `id` de cada input, facilitant la lectura per lectors de pantalla.
- [OK] **Atributs d'autocompletat**: Ús adequat d'atributs estàndard com `autocomplete="email"` i `autocomplete="current-password"`.

## Paritat i18n
- [OK] **Traducció completa**: Les claus `auth.login.deviceVerified`, `auth.login.title`, `auth.login.subtitle`, `auth.login.email`, `auth.login.password`, `auth.login.submit`, `auth.login.noAccount`, `auth.login.registerLink` i `auth.login.error` estan traduïdes amb total paritat a `ca.json`, `es.json` i `en.json`.

| Clau i18n | Català (CA) | Castellà (ES) | Anglès (EN) |
| :--- | :--- | :--- | :--- |
| `auth.login.deviceVerified` | Dispositiu autoritzat i verificat | Dispositivo autorizado y verificado | Authorized and verified device |
| `auth.login.title` | Inici de Sessió | Inicio de Sesión | Sign In |
| `auth.login.subtitle` | Accedeix a la plataforma Encertia | Accede a la plataforma Encertia | Access the Encertia platform |
| `auth.login.email` | Correu electrònic | Correo electrónico | Email address |
| `auth.login.password` | Contrasenya | Contraseña | Password |
| `auth.login.submit` | Iniciar Sessió | Iniciar Sesión | Sign In |

## Incidències
1. **[Detall] Ocultació de la contrasenya a la consola CLI (`agent/cmd/main.go`)**:
   - *Flux*: Assistent de registre interactiu `encertia-agent --register` (pas 2: contrasenya de l'administrador).
   - *Impacte*: La funció `readPassword()` utilitza un fallback `bufio.NewReader(os.Stdin).ReadBytes('\n')` que no deshabilita l'eco a la terminal.
   - *Suggeriment*: Incorporar `golang.org/x/term` amb `term.ReadPassword(int(syscall.Stdin))` per amagar la contrasenya amb asteriscos o caràcters buits en qualsevol terminal compatible.
2. **[Detall] Autofocus al camp de contrasenya (`LoginView.vue`)**:
   - *Flux*: Accés del professor redirigit des de l'agent CLI amb correu electrònic pre-omplert.
   - *Impacte*: El professor ha de fer clic manualment al camp de contrasenya tot i tenir el correu ja emplenat.
   - *Suggeriment*: Si `credentials.email` ve emplenat a través de `route.query.email`, assignar programàticament el focus a l'element d'entrada de la contrasenya en muntar la vista.
