<template>
  <div class="manual-page-container">
    <header class="manual-header">
      <div class="header-content">
        <div class="header-badge">
          <i class="pi pi-book"></i>
          <span>{{ $t('manual.officialDoc') }}</span>
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
          <h3 class="sidebar-title">{{ $t('manual.indexTitle') }}</h3>
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
            <h2>{{ $t('manual.sections.intro.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>
                {{ $t('manual.sections.intro.desc') }}
              </p>
              <div class="features-grid">
                <div class="feature-box">
                  <i class="pi pi-bolt feature-icon primary"></i>
                  <h4>{{ $t('manual.sections.intro.realtimeTitle') }}</h4>
                  <p>{{ $t('manual.sections.intro.realtimeDesc') }}</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-chart-line feature-icon success"></i>
                  <h4>{{ $t('manual.sections.intro.scoreTitle') }}</h4>
                  <p>{{ $t('manual.sections.intro.scoreDesc') }}</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-desktop feature-icon warning"></i>
                  <h4>{{ $t('manual.sections.intro.scriptTitle') }}</h4>
                  <p>{{ $t('manual.sections.intro.scriptDesc') }}</p>
                </div>
                <div class="feature-box">
                  <i class="pi pi-folder feature-icon info"></i>
                  <h4>{{ $t('manual.sections.intro.materialsTitle') }}</h4>
                  <p>{{ $t('manual.sections.intro.materialsDesc') }}</p>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 2: Autenticació -->
        <section id="sec-auth" class="manual-section">
          <div class="section-header">
            <i class="pi pi-shield section-icon"></i>
            <h2>{{ $t('manual.sections.auth.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <Accordion :value="['0']" multiple>
                <AccordionPanel value="0">
                  <AccordionHeader>{{ $t('manual.sections.auth.loginTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.auth.loginDesc') }}</p>
                    <ul>
                      <li>{{ $t('manual.sections.auth.accessToken') }}</li>
                      <li>{{ $t('manual.sections.auth.refreshToken') }}</li>
                    </ul>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="1">
                  <AccordionHeader>{{ $t('manual.sections.auth.logoutTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.auth.logoutDesc') }}</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="2">
                  <AccordionHeader>{{ $t('manual.sections.auth.rbacTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.auth.rbacDesc') }}</p>
                    <div class="tags-container mb-2">
                      <Tag severity="danger">Admin</Tag> {{ $t('manual.sections.auth.roleAdmin') }}
                    </div>
                    <div class="tags-container mb-2">
                      <Tag severity="info">{{ $t('nav.roles.teacher') }}</Tag> {{ $t('manual.sections.auth.roleTeacher') }}
                    </div>
                    <div class="tags-container">
                      <Tag severity="success">{{ $t('nav.roles.student') }}</Tag> {{ $t('manual.sections.auth.roleStudent') }}
                    </div>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </template>
          </Card>
        </section>

        <!-- Section 3: Gestió de Grups -->
        <section id="sec-groups" class="manual-section">
          <div class="section-header">
            <i class="pi pi-users section-icon"></i>
            <h2>{{ $t('manual.sections.groups.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.groups.intro') }}</p>

              <div class="steps-list">
                <div class="step-item">
                  <div class="step-number">1</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.groups.step1Title') }}</h4>
                    <p>{{ $t('manual.sections.groups.step1Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">2</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.groups.step2Title') }}</h4>
                    <p>{{ $t('manual.sections.groups.step2Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">3</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.groups.step3Title') }}</h4>
                    <p>{{ $t('manual.sections.groups.step3Desc') }}</p>
                  </div>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 4: Gestió d'Alumnes -->
        <section id="sec-users" class="manual-section">
          <div class="section-header">
            <i class="pi pi-user section-icon"></i>
            <h2>{{ $t('manual.sections.users.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.users.intro') }}</p>

              <div class="steps-list">
                <div class="step-item">
                  <div class="step-number">1</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.users.step1Title') }}</h4>
                    <p>{{ $t('manual.sections.users.step1Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">2</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.users.step2Title') }}</h4>
                    <p>{{ $t('manual.sections.users.step2Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">3</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.users.step3Title') }}</h4>
                    <p>{{ $t('manual.sections.users.step3Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">4</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.users.step4Title') }}</h4>
                    <p>{{ $t('manual.sections.users.step4Desc') }}</p>
                  </div>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 5: Qüestionaris -->
        <section id="sec-quizzes" class="manual-section">
          <div class="section-header">
            <i class="pi pi-th-large section-icon"></i>
            <h2>{{ $t('manual.sections.quizzes.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.quizzes.intro') }}</p>

              <Message severity="info" class="mb-4">
                {{ $t('manual.sections.quizzes.styleKahoot') }}
              </Message>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.quizzes.creationTitle') }}</h3>
                <ul>
                  <li>{{ $t('manual.sections.quizzes.questionType') }}</li>
                  <li>{{ $t('manual.sections.quizzes.answerOptions') }}</li>
                  <li>{{ $t('manual.sections.quizzes.timer') }}</li>
                  <li>{{ $t('manual.sections.quizzes.images') }}</li>
                </ul>
              </div>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.quizzes.duplicationTitle') }}</h3>
                <p>{{ $t('manual.sections.quizzes.duplicationDesc') }}</p>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 6: Cursos i Unitats -->
        <section id="sec-courses" class="manual-section">
          <div class="section-header">
            <i class="pi pi-book section-icon"></i>
            <h2>{{ $t('manual.sections.courses.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.courses.intro') }}</p>

              <div class="features-grid">
                <div class="feature-box">
                  <h4>{{ $t('manual.sections.courses.step1Title') }}</h4>
                  <p>{{ $t('manual.sections.courses.step1Desc') }}</p>
                </div>
                <div class="feature-box">
                  <h4>{{ $t('manual.sections.courses.step2Title') }}</h4>
                  <p>{{ $t('manual.sections.courses.step2Desc') }}</p>
                </div>
                <div class="feature-box">
                  <h4>{{ $t('manual.sections.courses.step3Title') }}</h4>
                  <p>{{ $t('manual.sections.courses.step3Desc') }}</p>
                </div>
                <div class="feature-box">
                  <h4>{{ $t('manual.sections.courses.step4Title') }}</h4>
                  <p>{{ $t('manual.sections.courses.step4Desc') }}</p>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 7: Materials Didàctics -->
        <section id="sec-materials" class="manual-section">
          <div class="section-header">
            <i class="pi pi-folder-open section-icon"></i>
            <h2>{{ $t('manual.sections.materials.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.materials.intro') }}</p>

              <Accordion value="0">
                <AccordionPanel value="0">
                  <AccordionHeader>{{ $t('manual.sections.materials.docsTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.materials.docsDesc') }}</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="1">
                  <AccordionHeader>{{ $t('manual.sections.materials.pdfTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.materials.pdfDesc') }}</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="2">
                  <AccordionHeader>{{ $t('manual.sections.materials.replaceTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.materials.replaceDesc') }}</p>
                  </AccordionContent>
                </AccordionPanel>

                <AccordionPanel value="3">
                  <AccordionHeader>{{ $t('manual.sections.materials.metricsTitle') }}</AccordionHeader>
                  <AccordionContent>
                    <p>{{ $t('manual.sections.materials.metricsDesc') }}</p>
                  </AccordionContent>
                </AccordionPanel>
              </Accordion>
            </template>
          </Card>
        </section>

        <!-- Section 8: Guió de Classe -->
        <section id="sec-script" class="manual-section">
          <div class="section-header">
            <i class="pi pi-desktop section-icon"></i>
            <h2>{{ $t('manual.sections.script.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.script.intro') }}</p>

              <div class="block-types-container">
                <div class="block-type-card material-block">
                  <i class="pi pi-file-pdf"></i>
                  <h4>{{ $t('manual.sections.script.pdfBlockTitle') }}</h4>
                  <p>{{ $t('manual.sections.script.pdfBlockDesc') }}</p>
                </div>
                <div class="block-type-card quiz-block">
                  <i class="pi pi-play"></i>
                  <h4>{{ $t('manual.sections.script.quizBlockTitle') }}</h4>
                  <p>{{ $t('manual.sections.script.quizBlockDesc') }}</p>
                </div>
                <div class="block-type-card break-block">
                  <i class="pi pi-clock"></i>
                  <h4>{{ $t('manual.sections.script.breakBlockTitle') }}</h4>
                  <p>{{ $t('manual.sections.script.breakBlockDesc') }}</p>
                </div>
              </div>

              <Message severity="success" class="mt-4">
                {{ $t('manual.sections.script.liveNote') }}
              </Message>
            </template>
          </Card>
        </section>

        <!-- Section 9: Partida en Directe i Jugadors Anònims -->
        <section id="sec-match" class="manual-section">
          <div class="section-header">
            <i class="pi pi-play-circle section-icon"></i>
            <h2>{{ $t('manual.sections.match.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.match.intro') }}</p>

              <div class="steps-list">
                <div class="step-item">
                  <div class="step-number">1</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.match.step1Title') }}</h4>
                    <p>{{ $t('manual.sections.match.step1Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">2</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.match.step2Title') }}</h4>
                    <p>{{ $t('manual.sections.match.step2Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">3</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.match.step3Title') }}</h4>
                    <p>{{ $t('manual.sections.match.step3Desc') }}</p>
                  </div>
                </div>

                <div class="step-item">
                  <div class="step-number">4</div>
                  <div class="step-body">
                    <h4>{{ $t('manual.sections.match.step4Title') }}</h4>
                    <p>{{ $t('manual.sections.match.step4Desc') }}</p>
                  </div>
                </div>
              </div>
            </template>
          </Card>
        </section>

        <!-- Section 10: Avaluació Acadèmica i Blocs Col·lapsables -->
        <section id="sec-evaluation" class="manual-section">
          <div class="section-header">
            <i class="pi pi-chart-bar section-icon"></i>
            <h2>{{ $t('manual.sections.evaluations.title') }}</h2>
          </div>
          <Card class="doc-card">
            <template #content>
              <p>{{ $t('manual.sections.evaluations.intro') }}</p>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.evaluations.filtersTitle') }}</h3>
                <p>{{ $t('manual.sections.evaluations.filtersDesc') }}</p>
              </div>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.evaluations.panelsTitle') }}</h3>
                <p>{{ $t('manual.sections.evaluations.panelsDesc') }}</p>
              </div>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.evaluations.scoringTitle') }}</h3>
                <div class="score-comparison">
                  <div class="score-card game">
                    <h4>{{ $t('evaluations.panels.globalStats') }}</h4>
                    <p>{{ $t('manual.sections.evaluations.gamePoints') }}</p>
                  </div>
                  <div class="score-card academic">
                    <h4>{{ $t('evaluations.panels.studentResults') }}</h4>
                    <p>{{ $t('manual.sections.evaluations.academicGrade') }}</p>
                  </div>
                </div>
              </div>

              <div class="manual-subsection">
                <h3>{{ $t('manual.sections.evaluations.gradingTitle') }}</h3>
                <p>{{ $t('manual.sections.evaluations.gradingDesc') }}</p>
              </div>
            </template>
          </Card>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Card from 'primevue/card'
import Accordion from 'primevue/accordion'
import AccordionPanel from 'primevue/accordionpanel'
import AccordionHeader from 'primevue/accordionheader'
import AccordionContent from 'primevue/accordioncontent'
import Tag from 'primevue/tag'
import Message from 'primevue/message'

const { t } = useI18n()
const activeSection = ref('sec-intro')

const sections = computed(() => [
  { id: 'sec-intro', title: t('manual.sections.intro.nav'), icon: 'pi pi-compass' },
  { id: 'sec-auth', title: t('manual.sections.auth.nav'), icon: 'pi pi-shield' },
  { id: 'sec-groups', title: t('manual.sections.groups.nav'), icon: 'pi pi-users' },
  { id: 'sec-users', title: t('manual.sections.users.nav'), icon: 'pi pi-user' },
  { id: 'sec-quizzes', title: t('manual.sections.quizzes.nav'), icon: 'pi pi-th-large' },
  { id: 'sec-courses', title: t('manual.sections.courses.nav'), icon: 'pi pi-book' },
  { id: 'sec-materials', title: t('manual.sections.materials.nav'), icon: 'pi pi-folder-open' },
  { id: 'sec-script', title: t('manual.sections.script.nav'), icon: 'pi pi-desktop' },
  { id: 'sec-match', title: t('manual.sections.match.nav'), icon: 'pi pi-play-circle' },
  { id: 'sec-evaluation', title: t('manual.sections.evaluations.nav'), icon: 'pi pi-chart-bar' }
])

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
