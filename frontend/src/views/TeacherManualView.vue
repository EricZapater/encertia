<template>
  <div class="manual-page-container">
    <header class="manual-header">
      <div class="header-content">
        <div class="header-badge">
          <i class="pi pi-book"></i>
          <span>Documentació Oficial</span>
        </div>
        <h1>{{ $t('manual.title') }} — Encertia</h1>
        <p class="header-subtitle">
          {{ $t('manual.subtitle') }}
        </p>
      </div>
    </header>

    <div class="manual-main-layout">
      <!-- Sidebar Index Navigation -->
      <aside class="manual-sidebar">
        <div class="sidebar-sticky">
          <h3 class="sidebar-title">Índex de continguts</h3>
          <nav class="sidebar-nav">
            <a
              v-for="section in sections"
              :key="section.id"
              :href="'#' + section.id"
              class="sidebar-link"
              :class="{ active: activeSection === section.id }"
              @click.prevent="scrollToSection(section.id)"
            >
              <i :class="section.icon"></i>
              <span>{{ section.title }}</span>
            </a>
          </nav>
        </div>
      </aside>

      <!-- Main Documentation Content -->
      <main class="manual-content">
        <!-- Section 1: Introducció -->
        <section id="sec-intro" class="manual-section">
          <div class="section-header">
            <i class="pi pi-compass section-icon"></i>
            <h2>1. Introducció i Filosofia d'Encertia</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>
                <strong>Encertia</strong> és una plataforma educativa d'avaluació i gestió de curs dissenyada específicament per a la docència universitària i acadèmica. Combina la motivació d'un joc de preguntes en temps real en directe (estil Kahoot) amb la rigorositat d'un <strong>sistema de gestió d'aprenentatge (LMS)</strong> i el registre persistent d'avaluació per alumne.
              </p>
              <div class="features-grid">
                <div class="feature-box">
                  <i class="pi pi-bolt feature-icon primary"></i>
                  <h4>Joc en temps real</h4>
                  <p>Partides interactives amb codi PIN, codi QR, temporitzadors i podi 3D animat.</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-chart-line feature-icon success"></i>
                  <h4>Doble Puntuació</h4>
                  <p>Separació estricta entre punts de joc (rapidesa) i nota acadèmica (encert absolut 0-10).</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-desktop feature-icon warning"></i>
                  <h4>Guió de Classe</h4>
                  <p>Visor seqüencial que combina diapositives PDF, pauses i partides en un sol clic.</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-folder feature-icon info"></i>
                  <h4>Gestió de Materials</h4>
                  <p>Visor de PDF integrat, vídeos encastats i registre de lectures per alumne.</p>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 2: Autenticació -->
        <section id="sec-auth" class="manual-section">
          <div class="section-header">
            <i class="pi pi-shield section-icon"></i>
            <h2>2. Autenticació, Sessió i Seguretat</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <Accordion :value="['0']" multiple>
                <AccordionPanel value="0">
                  <AccordionHeader>Inici de Sessió i Renovació Transparent</AccordionHeader>
                  <AccordionContent>
                    <p>Accedeix amb les teves credencials (email i contrasenya) a <code>/login</code>. Encertia utilitza un sistema segur de parell de tokens JWT:</p>
                    <ul>
                      <li><strong>Access Token (15 minuts)</strong>: Utilitzat per autoritzar cada petició HTTP. S'auto-renova de manera transparent en segon pla.</li>
                      <li><strong>Refresh Token (7 dies)</strong>: Manté la sessió oberta de manera segura.</li>
                    </ul>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="1">
                  <AccordionHeader>Tancament de Sessió Efectiu (Logout Segur)</AccordionHeader>
                  <AccordionContent>
                    <p>En clicar el botó de tancar sessió a la barra superior, Encertia executa un <strong>logout real al servidor</strong>. L'Access Token actiu s'invalida immediatament a la taula de revocació de PostgreSQL (<code>revoked_access_tokens</code>), impedint que ningú pugui reutilitzar la sessió.</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="2">
                  <AccordionHeader>Control d'Accés per Rol (RBAC)</AccordionHeader>
                  <AccordionContent>
                    <p>La plataforma distingeix 3 rols d'usuari:</p>
                    <div class="tags-container">
                      <Tag severity="danger">Admin</Tag> Accés total a la configuració global i a tots els cursos.
                    </div>
                    <div class="tags-container">
                      <Tag severity="info">Professor</Tag> Creació i gestió dels seus cursos, alumnes, quizes i avaluacions.
                    </div>
                    <div class="tags-container">
                      <Tag severity="success">Alumne</Tag> Participació en partides en directe, consulta de materials i notes del seu curs.
                    </div>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </template>
          </Card>
        </section>

        <!-- Section 3: Gestió d'Usuaris -->
        <section id="sec-users" class="manual-section">
          <div class="section-header">
            <i class="pi pi-users section-icon"></i>
            <h2>3. Gestió d'Alumnes i Grup-Classe</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>Com a professor, pots gestionar el llistat d'alumnes des de la secció <router-link to="/users"><strong>Usuaris</strong></router-link>.</p>
              
              <div class="steps-list">
                <div class="step-item">
                  <div class="step-number">1</div>
                  <div class="step-body">
                    <h4>Alta Individual d'Alumne</h4>
                    <p>Clica a <strong>"Nou Usuari"</strong>, introdueix el nom, cognoms, email i assigna el rol d'alumne. El sistema generarà el compte immediatament.</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">2</div>
                  <div class="step-body">
                    <h4>Alta Massiva per CSV (Full de Càlcul)</h4>
                    <p>Fes servir l'assistent d'importació en 3 passos per pujar tot el grup-classe des d'un fitxer CSV/TSV. El cercador detecta automàticament els camps i valida cada fila abans de donar d'alta.</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">3</div>
                  <div class="step-body">
                    <h4>Reseteig Administratiu de Contrasenya</h4>
                    <p>Si un alumne oblida la contrasenya, pots resetejar-la directament des de la taula d'usuaris. La nova clau requerirà un mínim de 8 caràcters i tancarà automàticament les seves sessions anteriors.</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">4</div>
                  <div class="step-body">
                    <h4>Baixa Lògica (Soft-Delete)</h4>
                    <p>Al donar de baixa un alumne, Encertia executa un <em>soft-delete</em> (marca l'usuari com a inactiu). <strong>Cap historial acadèmic o resposta es perd mai.</strong></p>
                  </div>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 4: Qüestionaris -->
        <section id="sec-quizzes" class="manual-section">
          <div class="section-header">
            <i class="pi pi-th-large section-icon"></i>
            <h2>4. Banc de Qüestionaris (Quizzes)</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>Des de la secció <router-link to="/quizzes"><strong>Jocs & Quizzes</strong></router-link>, pots crear i mantenir el teu banc de qüestionaris reutilitzables.</p>

              <Message severity="info" class="mb-4">
                <strong>Estil Kahoot:</strong> L'editor de preguntes utilitza la paleta visual de 6 colors i formes distintives (▲ Vermell, ◆ Blau, ● Groc, ■ Verd, ★ Lila, ⬡ Taronja).
              </Message>

              <div class="manual-subsection">
                <h3>Creació i Edició de Preguntes</h3>
                <ul>
                  <li><strong>Tipus de Pregunta</strong>: Opció Única (exactament 1 resposta correcta) o Opció Múltiple (1 o més respostes correctes).</li>
                  <li><strong>Opcions de Resposta</strong>: Entre 2 i 6 opcions per pregunta.</li>
                  <li><strong>Temporitzador</strong>: Configurable individualment per pregunta (5s, 10s, 20s, 30s, 60s, 90s o 120s).</li>
                  <li><strong>Imatges</strong>: Pujada d'imatge de portada de quiz i imatge per pregunta (amb integració Cloudflare R2).</li>
                </ul>
              </div>

              <div class="manual-subsection">
                <h3>Duplicació de Qüestionaris</h3>
                <p>Pots duplicar qualsevol qüestionari existent amb l'opció de copiar només l'enunciat de les preguntes o incloure també totes les respostes de base.</p>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 5: Cursos i Unitats -->
        <section id="sec-courses" class="manual-section">
          <div class="section-header">
            <i class="pi pi-book section-icon"></i>
            <h2>5. Gestió de Cursos i Unitats Didàctiques</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>A la secció <router-link to="/courses"><strong>Cursos</strong></router-link>, s'organitza l'assignatura en unitats didàctiques seqüencials (on <em>"unitat"</em> i <em>"classe"</em> representen el mateix concepte).</p>

              <div class="features-grid">
                <div class="feature-box">
                  <h4>1. Creació del Curs</h4>
                  <p>Defineix el títol, codi d'assignatura, descripció i estat (Esborrany, Actiu, Arxivat).</p>
                </div>
                <div class="feature-box">
                  <h4>2. Matriculació d'Alumnes</h4>
                  <p>Inscriu els alumnes del grup-classe al curs per donar-los accés als materials i quizes.</p>
                </div>
                <div class="feature-box">
                  <h4>3. Unitats i Relació N:N</h4>
                  <p>Crea unitats didàctiques i vincula-hi un o més qüestionaris. Un mateix quiz es pot reutilitzar en diferents unitats.</p>
                </div>
                <div class="feature-box">
                  <h4>4. Reordenació d'Unitats</h4>
                  <p>Reordena les unitats didàctiques segons l'avanç del quadrimestre.</p>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 6: Materials Didàctics -->
        <section id="sec-materials" class="manual-section">
          <div class="section-header">
            <i class="pi pi-folder-open section-icon"></i>
            <h2>6. Materials Didàctics, Visor PDF i Mètriques</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>El mòdul <router-link to="/materials"><strong>Materials</strong></router-link> centralitza el repositori de recursos del professor.</p>

              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>Documents i Vídeos Incrustats</AccordionHeader>
                  <AccordionContent>
                    <ul>
                      <li><strong>Documents (PDF, DOCX, PPTX)</strong>: Carrega fitxers de fins a 50 MB.</li>
                      <li><strong>Vídeos Externs</strong>: Enllaça vídeos de YouTube o Vimeo. La plataforma detecta automàticament el proveïdor i genera el reproductor encastat.</li>
                    </ul>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="1">
                  <AccordionHeader>Visor de PDF Integrat</AccordionHeader>
                  <AccordionContent>
                    <p>Els alumnes i professors poden llegir els documents PDF directament a l'aplicació pàgina a pàgina, sense necessitat de descarregar-los prèviament.</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="2">
                  <AccordionHeader>Substitució Transparent de Fitxers</AccordionHeader>
                  <AccordionContent>
                    <p>Si actualitzes un PDF o document, pots pujar la nova versió conservant el mateix ID de material. Això garanteix que les unitats didàctiques i el guió de classe no perdin mai l'enllaç.</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="3">
                  <AccordionHeader>Informe d'Accessos i Lectures (Mètriques)</AccordionHeader>
                  <AccordionContent>
                    <p>Clica a la icona d'ull a qualsevol material per veure el panell d'informe: total de visualitzacions, alumnes únics que hi han accedit i la data de darrera lectura de cadascun.</p>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </template>
          </Card>
        </section>

        <!-- Section 7: Guió de Classe -->
        <section id="sec-script" class="manual-section">
          <div class="section-header">
            <i class="pi pi-desktop section-icon"></i>
            <h2>7. Guió de Classe (Visor Seqüencial)</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>El <strong>Guió de Classe</strong> permet dissenyar per endavant la seqüència exacta de la classe per projectar-la a l'aula en directe.</p>

              <div class="block-types-container">
                <div class="block-type-card material-block">
                  <i class="pi pi-file-pdf"></i>
                  <h4>Bloc de Material PDF</h4>
                  <p>Defineix un rang de pàgines d'un document (ex. "Diapositives 1 a 15") per mostrar al visor.</p>
                </div>
                <div class="block-type-card quiz-block">
                  <i class="pi pi-play"></i>
                  <h4>Bloc de Qüestionari</h4>
                  <p>Llança automàticament la partida en directe (`match`) d'un quiz associat en arribar al bloc.</p>
                </div>
                <div class="block-type-card break-block">
                  <i class="pi pi-clock"></i>
                  <h4>Bloc de Pausa / Preguntes</h4>
                  <p>Pausa temporitzada per aclariments oberts i debat amb el grup-classe.</p>
                </div>
              </div>

              <Message severity="success" class="mt-4">
                <strong>En directe:</strong> El professor només ha d'anar clicant <em>"Següent"</em> per anar avançant fluidament entre explicació i moments interactius.
              </Message>
            </template>
          </Card>
        </section>

        <!-- Section 8: Partida en Directe -->
        <section id="sec-match" class="manual-section">
          <div class="section-header">
            <i class="pi pi-play-circle section-icon"></i>
            <h2>8. Partida en Directe (Match)</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>En llançar un joc en directe des d'un quiz o des del guió de classe, s'obre el panell de projecció del moderador (`HostGameView`).</p>

              <ol class="steps-numbered">
                <li><strong>Sala d'Espera (Lobby)</strong>: Es mostra el codi PIN de 6 dígits i el <strong>codi QR interactiu natiu</strong> per a la connexió des del mòbil dels alumnes.</li>
                <li><strong>Control del Ritme</strong>: El professor activa el temps de lectura i inicia el compte enrere quan el grup està a punt.</li>
                <li><strong>Resultats en Temps Real</strong>: Al tancar el temps de cada pregunta, es projecta un gràfic de barres animat amb la distribució de vots i la resposta correcta.</li>
                <li><strong>Podi 3D Final</strong>: En acabar la partida, es revela el podi de 3D dels tres primers classificats (🥇, 🥈, 🥉).</li>
              </ol>
            </template>
          </Card>
        </section>

        <!-- Section 9: Avaluació Acadèmica -->
        <section id="sec-evaluation" class="manual-section">
          <div class="section-header">
            <i class="pi pi-chart-bar section-icon"></i>
            <h2>9. Panell d'Avaluació Acadèmica</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>A la secció <router-link to="/evaluations"><strong>Avaluacions</strong></router-link>, el professor consulta l'avaluació acadèmica consolidada.</p>

              <div class="manual-subsection">
                <h3>Sistema de Doble Puntuació</h3>
                <div class="score-comparison">
                  <div class="score-card game">
                    <h4>Punts de Joc (Gamificació)</h4>
                    <p>Considera encert + rapidesa. Alimenta el rànquing i el podi en directe.</p>
                  </div>
                  <div class="score-card academic">
                    <h4>Nota d'Avaluació (Acadèmica)</h4>
                    <p>Només considera l'encert absolut (0.00 a 10.00). La pressió del temps no afecta la nota.</p>
                  </div>
                </div>
              </div>

              <div class="manual-subsection">
                <h3>Qualificació Automàtica i Ajust Manual</h3>
                <p>En finalitzar cada partida, Encertia calcula automàticament la nota de l'alumne. El professor pot revisar el desglossament pregunta per pregunta i realitzar un <strong>ajust manual de la nota final</strong> (`finalGrade`) si ho considera oportú.</p>
              </div>
            </template>
          </Card>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Card from 'primevue/card'
import Accordion from 'primevue/accordion'
import AccordionPanel from 'primevue/accordionpanel'
import AccordionHeader from 'primevue/accordionheader'
import AccordionContent from 'primevue/accordioncontent'
import Tag from 'primevue/tag'
import Message from 'primevue/message'

useI18n()
const activeSection = ref('sec-intro')

const sections = [
  { id: 'sec-intro', title: '1. Introducció i Filosfia', icon: 'pi pi-compass' },
  { id: 'sec-auth', title: '2. Autenticació i Seguretat', icon: 'pi pi-shield' },
  { id: 'sec-users', title: '3. Gestió d\'Alumnes', icon: 'pi pi-users' },
  { id: 'sec-quizzes', title: '4. Banc de Qüestionaris', icon: 'pi pi-th-large' },
  { id: 'sec-courses', title: '5. Gestió de Cursos', icon: 'pi pi-book' },
  { id: 'sec-materials', title: '6. Materials Didàctics', icon: 'pi pi-folder-open' },
  { id: 'sec-script', title: '7. Guió de Classe', icon: 'pi pi-desktop' },
  { id: 'sec-match', title: '8. Partida en Directe', icon: 'pi pi-play-circle' },
  { id: 'sec-evaluation', title: '9. Panell d\'Avaluació', icon: 'pi pi-chart-bar' }
]

function scrollToSection(id: string) {
  activeSection.value = id
  const element = document.getElementById(id)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}
</script>

<style scoped>
.manual-page-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem 1.5rem;
}

.manual-header {
  background: linear-gradient(135deg, #4f46e5 0%, #3730a3 100%);
  border-radius: 1rem;
  padding: 2.5rem;
  color: #ffffff;
  margin-bottom: 2rem;
  box-shadow: 0 10px 15px -3px rgba(79, 70, 229, 0.2);
}

.header-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background-color: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(4px);
  padding: 0.35rem 0.85rem;
  border-radius: 2rem;
  font-size: 0.85rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.manual-header h1 {
  font-size: 2.2rem;
  font-weight: 800;
  margin: 0 0 0.75rem 0;
  letter-spacing: -0.02em;
}

.header-subtitle {
  font-size: 1.05rem;
  opacity: 0.9;
  max-width: 800px;
  line-height: 1.6;
  margin: 0;
}

.manual-main-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 2rem;
  align-items: start;
}

.manual-sidebar {
  position: relative;
}

.sidebar-sticky {
  position: sticky;
  top: 5.5rem;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.sidebar-title {
  font-size: 1rem;
  font-weight: 700;
  color: #1e293b;
  margin-top: 0;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 0.75rem;
  border-radius: 0.5rem;
  color: #475569;
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.15s ease;
}

.sidebar-link:hover {
  background-color: #f8fafc;
  color: #4f46e5;
}

.sidebar-link.active {
  background-color: #eef2ff;
  color: #4f46e5;
  font-weight: 600;
}

.manual-content {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.manual-section {
  scroll-margin-top: 5.5rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.section-icon {
  font-size: 1.5rem;
  color: #4f46e5;
  background-color: #eef2ff;
  padding: 0.6rem;
  border-radius: 0.5rem;
}

.section-header h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}

.doc-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.25rem;
  margin-top: 1.5rem;
}

.feature-box {
  background-color: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 0.6rem;
  padding: 1.25rem;
}

.feature-icon {
  font-size: 1.4rem;
  margin-bottom: 0.75rem;
}

.feature-icon.primary { color: #4f46e5; }
.feature-icon.success { color: #10b981; }
.feature-icon.warning { color: #f59e0b; }
.feature-icon.info { color: #06b6d4; }

.feature-box h4 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
}

.feature-box p {
  margin: 0;
  font-size: 0.875rem;
  color: #64748b;
  line-height: 1.5;
}

.steps-list {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  margin-top: 1.25rem;
}

.step-item {
  display: flex;
  gap: 1rem;
  align-items: start;
}

.step-number {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  background-color: #4f46e5;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.step-body h4 {
  margin: 0 0 0.35rem 0;
  font-size: 1.05rem;
  color: #1e293b;
}

.step-body p {
  margin: 0;
  color: #475569;
  font-size: 0.925rem;
  line-height: 1.5;
}

.tags-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.manual-subsection {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
}

.manual-subsection h3 {
  font-size: 1.15rem;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 0.75rem;
}

.block-types-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.block-type-card {
  padding: 1.25rem;
  border-radius: 0.6rem;
  border: 1px solid;
}

.material-block { background-color: #eff6ff; border-color: #bfdbfe; color: #1e40af; }
.quiz-block { background-color: #f0fdf4; border-color: #bbf7d0; color: #166534; }
.break-block { background-color: #fffbeb; border-color: #fef08a; color: #854d0e; }

.block-type-card i { font-size: 1.5rem; margin-bottom: 0.5rem; }
.block-type-card h4 { margin: 0 0 0.35rem 0; font-size: 1rem; }
.block-type-card p { margin: 0; font-size: 0.85rem; opacity: 0.9; }

.score-comparison {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  margin-top: 1rem;
}

.score-card {
  padding: 1.25rem;
  border-radius: 0.6rem;
  border: 1px solid;
}

.score-card.game { background-color: #faf5ff; border-color: #e9d5ff; color: #6b21a8; }
.score-card.academic { background-color: #f0fdf4; border-color: #bbf7d0; color: #166534; }

.score-card h4 { margin: 0 0 0.5rem 0; font-size: 1rem; }
.score-card p { margin: 0; font-size: 0.875rem; }

.steps-numbered {
  padding-left: 1.25rem;
  line-height: 1.7;
  color: #334155;
}

@media (max-width: 992px) {
  .manual-main-layout {
    grid-template-columns: 1fr;
  }
  .manual-sidebar {
    display: none;
  }
  .score-comparison {
    grid-template-columns: 1fr;
  }
}
</style>
