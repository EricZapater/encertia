import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  EvaluationQuizSummary,
  QuizEvaluationResponse,
  StudentEvaluationDetail
} from './types'
import {
  listEvaluations,
  getQuizEvaluation,
  getStudentEvaluation,
  gradeStudent as gradeStudentApi
} from './api'

export const useEvaluationStore = defineStore('evaluations', () => {
  const evaluationsList = ref<EvaluationQuizSummary[]>([])
  const activeQuizEvaluation = ref<QuizEvaluationResponse | null>(null)
  const activeStudentEvaluation = ref<StudentEvaluationDetail | null>(null)

  const isLoading = ref(false)
  const isSavingGrade = ref(false)
  const error = ref<string | null>(null)

  const hasEvaluations = computed(() => evaluationsList.value.length > 0)

  async function fetchEvaluationsList() {
    isLoading.value = true
    error.value = null
    try {
      const response = await listEvaluations()
      evaluationsList.value = response.evaluations || []
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Error en carregar les avaluacions.'
    } finally {
      isLoading.value = false
    }
  }

  async function fetchQuizEvaluation(quizId: string) {
    isLoading.value = true
    error.value = null
    try {
      const data = await getQuizEvaluation(quizId)
      activeQuizEvaluation.value = data
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Error en carregar l’avaluació del quiz.'
    } finally {
      isLoading.value = false
    }
  }

  async function fetchStudentEvaluation(quizId: string, studentId: string) {
    isLoading.value = true
    error.value = null
    try {
      const data = await getStudentEvaluation(quizId, studentId)
      activeStudentEvaluation.value = data
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || 'Error en carregar el detall de l’alumne.'
    } finally {
      isLoading.value = false
    }
  }

  async function saveStudentGrade(quizId: string, studentId: string, finalGrade: number) {
    isSavingGrade.value = true
    error.value = null
    try {
      const res = await gradeStudentApi(quizId, studentId, finalGrade)

      if (activeStudentEvaluation.value && activeStudentEvaluation.value.studentId === studentId) {
        activeStudentEvaluation.value.finalGrade = res.finalGrade
        activeStudentEvaluation.value.isGraded = res.isGraded
        activeStudentEvaluation.value.gradedBy = res.gradedBy
        activeStudentEvaluation.value.gradedAt = res.gradedAt
      }

      if (activeQuizEvaluation.value && activeQuizEvaluation.value.quizId === quizId) {
        const studentItem = activeQuizEvaluation.value.students.find(s => s.studentId === studentId)
        if (studentItem) {
          studentItem.finalGrade = res.finalGrade
          studentItem.isGraded = res.isGraded
        }
      }
      return res
    } catch (err: any) {
      const errMsg = err.response?.data?.error?.message || err.message || 'Error en desar la nota.'
      error.value = errMsg
      throw new Error(errMsg)
    } finally {
      isSavingGrade.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    evaluationsList,
    activeQuizEvaluation,
    activeStudentEvaluation,
    isLoading,
    isSavingGrade,
    error,
    hasEvaluations,
    fetchEvaluationsList,
    fetchQuizEvaluation,
    fetchStudentEvaluation,
    saveStudentGrade,
    clearError
  }
})
