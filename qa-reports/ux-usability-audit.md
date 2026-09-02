# Auditoria d'Usabilitat, Experiència d'Usuari (UX/UI) i Coherència Visual — Encertia

> **Projecte**: Encertia (Plataforma Educativa: Quiz + LMS)  
> **Data**: Segona meitat de 2026  
> **Àmbit**: Tota la interfície d'usuari (`frontend/src/`)  
> **Auditor**: Agent UX / UI & Usability Expert  
> **Principi Guia (Constitution §4)**: *"El professor/usuari final no és tècnic; qualsevol UI ha de ser clara sense necessitat d'explicació prèvia."*

---

## 1. Resum Executiu

Aquesta auditoria avalua l'experiència d'usuari, ergonomia docent, accessibilitat, responsivitat mòbil i coherència visual de tots els mòduls de la plataforma **Encertia**.

En línies generals, Encertia presenta un disseny modern, estructurat i clarament orientat al sector educatiu. La integració del framework **PrimeVue** (tema Aura) amb la paleta dinàmica d'estil **Kahoot** (6 colors i formes geomètriques úniques: ▲ Vermell, ◆ Blau, ● Groc, ■ Verd, ★ Lila, ⬡ Taronja) aconsegueix un balanç entre la rigorositat acadèmica d'un LMS i la ludificació d'una activitat en directe.

### Punts Forts Destacats:
1. **Doble Puntuació Acadèmica**: Distribució clara entre punts de joc (rapidesa + encert) i nota acadèmica (encert absolut 0.00-10.00), evitant penalitzar alumnes per lentitud en l'avaluació oficial.
2. **Guió de Classe Seqüencial**: Integració fluida entre explicació teòrica (PDF), pauses/preguntes i llançament automàtic de partides en directe.
3. **Ergonomia del Moderador (`HostGameView`)**: Visualització clara del PIN gegant de 6 dígits, generació de codi QR natiu en temps real, gestió d'expulsió de jugadors i podi 3D animat.
4. **Accessibilitat i Multilingüisme**: Suport integrat per a Català (`ca`), Castellà (`es`) i Anglès (`en`), accessible des del navbar superior i el perfil.
5. **Prevenció d'Errors i Dades Sensibles**: Missatges informatius en diàlegs destructius senyalant la baixa lògica (*soft-delete*), protegint l'historial acadèmic de l'alumnat.

---

## 2. Auditoria Mòdul per Mòdul

### 2.1. Mòdul `auth` (Login, Register, Profile)
* **Fitxers analitzats**: [LoginView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/auth/views/LoginView.vue), [RegisterView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/auth/views/RegisterView.vue), [ProfileView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/auth/views/ProfileView.vue)

#### Anàlisi i Troballes UX:
- **Login & Registre**: Formulari centrat amb targeta PrimeVue (`Card`), fons amb degradat suau (`#f0f4f8` a `#e2e8f0`) que transmet tranquil·litat i claredat visual.
- **Selector d'Idioma**: Inclòs tant a la part superior de la targeta d'autenticació com al perfil, permetent canviar ràpidament entre Català, Castellà i Anglès.
- **Profil d'Usuari**: Avatar amb inicials generades automàticament i degradat blau, insígnies de rol (`Admin`: vermell/danger, `Professor`: blau/info, `Alumne`: verd/success) i blocs de dret d'accés segons rol (`adminPanel`, `teacherPanel`, `studentPanel`).
- **Aspectes a millorar**:
  - A `RegisterView.vue`, el feedback de validació quan la contrasenya té menys de 8 caràcters utilitza el missatge genèric `t('common.error')` en lloc d'un detall d'error específic (`"La contrasenya ha de tenir almenys 8 caràcters"`).

---

### 2.2. Mòdul `users` (Users List, Forms, Batch CSV Import, Reset Password)
* **Fitxers analitzats**: [UsersListView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/users/views/UsersListView.vue), [UserFormModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/users/views/UserFormModal.vue), [BatchImportModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/users/views/BatchImportModal.vue), [ResetPasswordModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/users/views/ResetPasswordModal.vue), [UserDetailView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/users/views/UserDetailView.vue)

#### Anàlisi i Troballes UX:
- **Llistat i Cerca**: Utilitza `DataTable` paginada amb suport lazy, cerca amb *debounce* de 350ms per evitar sobrecarregar el servidor, i filtres per estat (`Actiu`/`Inactiu`) i rol (`Admin`, `Teacher`, `Student`).
- **Control d'Accés Docent**: Si l'usuari és un professor, el selector de rol es fixa automàticament a `student` per evitar confusions o modificacions indegudes de permisos d'administrador.
- **Importació Massiva en 3 Passos (CSV)**:
  1. *Selecció/Dropzone*: Zona d'arrossegament de fitxers amb descarrega de plantilla d'exemple en 1 clic (`plantilla_alumnes_encertia.csv`).
  2. *Previsualització*: Taula de resum que mostra clarament quines files són vàlides (verd) i quines tenen errors de format (vermell).
  3. *Resultats*: Resum estadístic de files sol·licitades, creades i amb error de servidor.
- **Reset de Contrasenya**: Diàleg clar que adverteix explícitament que la renovació revocarà les sessions actives per seguretat.
- **Confirmació de Baixa**: Avís clar de *soft-delete* que explica que l'alumne no podrà accedir però que el seu historial acadèmic es manté intacte.

---

### 2.3. Mòdul `quizzes` (Quizzes List, Quiz Editor, Duplicate, Preview, Settings)
* **Fitxers analitzats**: [QuizzesListView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/QuizzesListView.vue), [QuizEditorView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/QuizEditorView.vue), [QuizSettingsModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/QuizSettingsModal.vue), [QuizPreviewModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/QuizPreviewModal.vue), [DuplicateQuizModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/DuplicateQuizModal.vue)

#### Anàlisi i Troballes UX:
- **Targetes de Jocs (`QuizzesListView`)**: Graella de targetes amb imatge de portada o *placeholder* degradat, badges d'estat (`Esborrany`, `Publicat`, `Arxivat`), recompte de preguntes i tags.
- **Acció Ràpida "Llançar"**: Permet iniciar la partida en directe (`match`) directament des del llistat amb un sol clic si el joc està publicat.
- **Editor de Preguntes Kahoot-Style (`QuizEditorView`)**:
  - *Panell lateral de la llista de preguntes*: Thumbnails amb reordenació (botons amunt/avall), duplicació i eliminació ràpida.
  - *Paleta de respostes Kahoot*: Utilitza la constant `KAHOOT_THEME_SHAPES` per assignar un color i símbol geomètric diferencial a cadascuna de les 2 a 6 opcions de resposta.
  - *Gestió de Temps i Imatges*: Selector de temps límit (5s-120s) i integració de pujada d'imatge de la pregunta.
- **Previsualització i Duplicació**: Modals dedicats per provar l'experiència abans de publicar i clonar jocs existents.

---

### 2.4. Mòdul `courses` (Courses List, Course Detail, Unit Editor, Script Player)
* **Fitxers analitzats**: [CoursesListView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/CoursesListView.vue), [CourseDetailView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/CourseDetailView.vue), [UnitEditorView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/UnitEditorView.vue), [ScriptPlayerView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/ScriptPlayerView.vue)

#### Anàlisi i Troballes UX:
- **Detall del Curs (`CourseDetailView`)**: Organitzat en dues pestanyes principals (*Unitats Didàctiques* i *Alumnes Matriculats*).
- **Matriculació d'Alumnes**: El modal `MultiSelect` de matriculació sol·licita els usuaris no matriculats prèviament, facilitant la cerca i selecció per nom i correu.
- **Editor de Guió de Classe (`UnitEditorView`)**:
  - Vinculació N:N de qüestionaris del banc general de quizzes.
  - Disseny del guió seqüencial combinant 3 tipus de blocs: `Material PDF` (amb seleccionador de pàgines inici-fi), `Qüestionari` i `Descans / Pausa`.
- **Reproductor de Guió (`ScriptPlayerView`)**: Permet al professor seguir la seqüència de la classe sense haver de canviar entre finestres o pestanyes del navegador.

---

### 2.5. Mòdul `materials` (Materials List, Upload Form, PDF Viewer, Video Embed, Views Report)
* **Fitxers analitzats**: [MaterialsListView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/materials/views/MaterialsListView.vue), [MaterialFormModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/materials/components/MaterialFormModal.vue), [PdfViewerModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/materials/components/PdfViewerModal.vue), [MaterialViewsReportModal.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/materials/components/MaterialViewsReportModal.vue)

#### Anàlisi i Troballes UX:
- **Gestió Unificada de Recursos**: Suport per a Documents (PDF, Word, PPTX) i Vídeos externs (detectant automàticament proveïdors com YouTube o Vimeo).
- **Visor de PDF Integrat (`PdfViewerModal`)**: Permet als alumnes llegir els apunts i documents pàgina a pàgina directament a la web, evitant descàrregues innecessàries.
- **Informe d'Accessos i Lectura**: Panell per al professorat que registra el total de visualitzacions, alumnes únics i la darrera data d'accés.
- **Substitució Transparent de Fitxers**: En actualitzar un PDF, es conserva l'ID del material, evitant que els guions de classe o unitats didàctiques quedin desvinculats.

---

### 2.6. Mòdul `match` (Player Join, Player Game, Host Lobby amb PIN/QR, Host Game, Podi 3D)
* **Fitxers analitzats**: [PlayerJoinView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/PlayerJoinView.vue), [PlayerGameView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/PlayerGameView.vue), [HostLobbyView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/HostLobbyView.vue), [HostGameView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/HostGameView.vue)

#### Anàlisi i Troballes UX:
- **Pantalla d'Entrada de l'Alumne (`PlayerJoinView`)**:
  - Camp PIN de 6 dígits numerat amb font gran i centrat.
  - Auto-emplenat del nom/nickname si l'alumne ja té la sessió iniciada.
  - Redirecció intel·ligent des d'URL amb paràmetre `?pin=123456` (ex. des del QR).
- **Lobby del Moderador (`HostGameView`)**:
  - PIN de la sala en format gegant (`font-weight: 900`).
  - Codi QR natiu generat dinàmicament en alta resolució.
  - Llista d'alumnes connectats amb opció d'expulsió (*kick*) per part del professor.
- **Flux de Partida i Visualització de Resultats**:
  - *Pregunta Activa*: temporitzador dinàmic, recompte de respostes registrades en temps real.
  - *Resultats de Pregunta*: gràfic de barres verticals amb percentatges, recompte de vots i indicació visual de la resposta correcta.
  - *Podi 3D Final*: animació dels pedestals per al 1r, 2n i 3r lloc amb corona i puntuacions.
- **Pantalla de Joc de l'Alumne (`PlayerGameView`)**:
  - Botons estil Kahoot adaptats a pantalles mòbils, amb feedback immediat de selecció i notificació de resposta enviada correctament.

---

### 2.7. Mòdul `evaluations` (Evaluations List, Quiz Evaluation, Student Evaluation, Grade Adjustment)
* **Fitxers analitzats**: [EvaluationsListView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/EvaluationsListView.vue), [QuizEvaluationView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/QuizEvaluationView.vue), [StudentEvaluationView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/StudentEvaluationView.vue)

#### Anàlisi i Troballes UX:
- **Estadístiques Globals per Pregunta**: Mostra la taxa d'encert (*hit rate* amb insígnia de color verd/groc/vermell segons el rendiment), el temps mitjà de resposta i la distribució percentual de totes les opcions.
- **Ajust Manual de Notes**: Permet al professorat consultar la nota calculada automàticament (`calculatedGrade`) i introduir una nota definitiva (`finalGrade`) de 0.00 a 10.00 amb l'assistent `InputNumber`.

---

### 2.8. Mòdul `manual` (Teacher User Manual View)
* **Fitxers analitzats**: [TeacherManualView.vue](file:///Users/eric.zapater/Developer/encertia/frontend/src/views/TeacherManualView.vue)

#### Anàlisi i Troballes UX:
- **Manual de l'Usuari Integrat**: Navegació mitjançant un índex lateral enganxós (*sticky sidebar*) de 9 seccions.
- **Claredat d'Explicació**: Utilitza acordions interactius, passos numerats i missatges destacats per guiar els professors no tècnics en la configuració i ús de la plataforma.

---

## 3. Anàlisi Transversal de Coherència Visual & UX

| Aspecte UX | Estat Actual | Valoració / Observació |
| :--- | :--- | :--- |
| **Biblioteca de Components** | PrimeVue (Aura Theme) | **Excel·lent**. Implementació consistent de `Button`, `DataTable`, `Dialog`, `Select`, `InputText`, `Password`, `Tag`, `Card`, `Toast`. |
| **Notificacions Toast** | `primevue/toastservice` | **Molt Bona**. Totes les accions clau (crear, editar, eliminar, errors de xarxa) emeten toasts temporitzats de 3 a 4 segons. |
| **Coherència de Colors** | Violeta/Indigo (`#6366f1` / `#4338ca`) + Paleta Kahoot | **Molt Bona**. Es manté la distinció clara entre accions primàries, secundàries, d'advertència i d'eliminació. |
| **Accessibilitat (a11y)** | Etiquetes `label`, contrast alt, lectors de pantalla | **Bona**. La majoria dels inputs tenen `id` i `label` explícits. Es recomana afegir `aria-label` als botons només amb icona. |
| **Responsivitat Mòbil** | Media queries a navbar, modals i `PlayerGameView` | **Molt Bona**. El panell del jugador (`/play`) està dissenyat prioritàriament per a pantalles d'smartphone. |
| **Prevenció d'Errors** | Diàlegs de confirmació amb detalls | **Excel·lent**. En operacions d'eliminació o baixa, s'informa clarament sobre l'ús de *soft-delete* per preservar l'historial acadèmic. |

---

## 4. Matriu de Recomanacions de Millora Ergònoma i UX

> [!NOTE]
> Les següents recomanacions estan classificades per prioritat per guiar futures millores de la interfície d'usuari.

### 🔴 Prioritat Alta (Impacte Directe en l'Usuari):
1. **Missatge de Validació Específic a Registre (`RegisterView.vue`)**:
   - *Problema*: Quan la contrasenya fa menys de 8 caràcters o falta algun camp, es mostra un detall genèric `t('common.error')`.
   - *Solució*: Substituir per un text descriptiu clar com `"La contrasenya ha de tenir com a mínim 8 caràcters"`.
2. **Accessibilitat als Botons d'Acció de Taules i Targetes**:
   - *Problema*: Diversos botons només amb icona (ex. botons d'edició o eliminació a les taules) no tenen l'atribut `aria-label` per a lectors de pantalla.
   - *Solució*: Afegir `:aria-label="tooltip || label"` a tots els botons que no tenen text visible.

### 🟡 Prioritat Mitjana (Millores d'Ergonomia Docent):
1. **Confirmació d'Abandó a l'Editor de Preguntes (`QuizEditorView.vue`)**:
   - *Proposta*: Si un professor fa canvis a un qüestionari i intenta sortir de la ruta sense desar, afegir un *Navigation Guard* de Vue Router (`onBeforeRouteLeave`) per demanar confirmació i evitar pèrdues accidentals de feina.
2. **Tecles de Drecera al Guió de Classe (`ScriptPlayerView.vue`)**:
   - *Proposta*: Permetre avançar de bloc de guió premant la tecla `Espai` o la fletxa `Dreta` (`→`), facilitant el control durant la projecció a l'aula sense haver de buscar el ratolí.

### 🟢 Prioritat Menor (Refinaments Visuals):
1. **Indicador Visual de Progrés al Panell d'Avaluació**:
   - *Proposta*: Afegir una barra de progrés circular o de percentatge a la llista d'avaluacions per mostrar visualment quin tant per cent dels alumnes de la partida han estat avaluats definitivament pel professor.

---

## 5. Veredicte Final UX

| Criteri d'Avaluació | Nota / Estat |
| :--- | :--- |
| **Claredat per a Usuari No Tècnic** | **4.8 / 5.0** |
| **Coherència Component i Disseny (PrimeVue)** | **4.9 / 5.0** |
| **Ergonomia Docent (Gestió Aula & Guió)** | **4.9 / 5.0** |
| **Accessibilitat i Multilingüisme** | **4.7 / 5.0** |
| **Responsivitat Mòbil (Alumnat)** | **5.0 / 5.0** |
| **VEREDICTE GLOBAL UX** | **APTE AMB EXCEL·LÈNCIA** |

---
*Informe redactat i auditat per l'Agent UX del projecte Encertia.*
