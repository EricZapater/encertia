import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/modules/auth/store'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: () => {
      const authStore = useAuthStore()
      return authStore.isAuthenticated ? '/profile' : '/login'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/modules/auth/views/LoginView.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/modules/auth/views/RegisterView.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/modules/auth/views/ProfileView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/users',
    name: 'users-list',
    component: () => import('@/modules/users/views/UsersListView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/users/:id',
    name: 'user-detail',
    component: () => import('@/modules/users/views/UserDetailView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/quizzes',
    name: 'quizzes-list',
    component: () => import('@/modules/quizzes/views/QuizzesListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/quizzes/new',
    name: 'quiz-create',
    component: () => import('@/modules/quizzes/views/QuizEditorView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/quizzes/:id/edit',
    name: 'quiz-edit',
    component: () => import('@/modules/quizzes/views/QuizEditorView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/play',
    name: 'match-join',
    component: () => import('@/modules/match/views/PlayerJoinView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/play/:pin',
    name: 'match-player',
    component: () => import('@/modules/match/views/PlayerGameView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/matches/:id/host',
    name: 'match-host',
    component: () => import('@/modules/match/views/HostGameView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/matches/:pin/host',
    name: 'match-host-pin',
    component: () => import('@/modules/match/views/HostGameView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/evaluations',
    name: 'evaluations-list',
    component: () => import('@/modules/evaluations/views/EvaluationsListView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/evaluations/quizzes/:quizId',
    name: 'quiz-evaluation',
    component: () => import('@/modules/evaluations/views/QuizEvaluationView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/evaluations/quizzes/:quizId/students/:studentId',
    name: 'student-evaluation',
    component: () => import('@/modules/evaluations/views/StudentEvaluationView.vue'),
    meta: { requiresAuth: true, roles: ['admin', 'teacher'] }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Navigation Guard per protegir rutes
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Inicialitza l'estat d'autenticació des del localStorage si no està inicialitzat
  if (!authStore.isInitialized) {
    await authStore.initAuth()
  }

  const isAuthenticated = authStore.isAuthenticated

  if (to.meta.requiresAuth && !isAuthenticated) {
    next({
      name: 'login',
      query: { redirect: to.fullPath !== '/profile' ? to.fullPath : undefined }
    })
    return
  }

  if (to.meta.requiresGuest && isAuthenticated) {
    next({ name: 'profile' })
    return
  }

  // Comprovació de rols permesos (RBAC)
  if (to.meta.roles && Array.isArray(to.meta.roles)) {
    const userRole = authStore.currentUser?.role
    if (!userRole || !to.meta.roles.includes(userRole)) {
      next({ name: 'profile' })
      return
    }
  }

  next()
})

export default router
