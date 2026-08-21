import apiClient from '@/api/client'
import type {
  EvaluationsListResponse,
  QuizEvaluationResponse,
  StudentEvaluationDetail,
  GradeRequest,
  GradeResponse
} from './types'
import {
  mockEvaluationsList,
  mockQuizEvaluations,
  mockStudentDetails
} from './mockData'

const isMockEnabled =
  import.meta.env.VITE_USE_MOCKS === 'true' ||
  import.meta.env.VITE_USE_MOCKS === undefined

export async function listEvaluations(): Promise<EvaluationsListResponse> {
  if (isMockEnabled) {
    try {
      const response = await apiClient.get<EvaluationsListResponse>('/evaluations')
      return response.data
    } catch {
      return mockEvaluationsList
    }
  }
  const response = await apiClient.get<EvaluationsListResponse>('/evaluations')
  return response.data
}

export async function getQuizEvaluation(quizId: string): Promise<QuizEvaluationResponse> {
  if (isMockEnabled) {
    try {
      const response = await apiClient.get<QuizEvaluationResponse>(`/evaluations/quizzes/${quizId}`)
      return response.data
    } catch {
      const mockData = mockQuizEvaluations[quizId]
      if (mockData) return mockData
      return {
        quizId,
        quizTitle: 'Qüestionari d’Avaluació',
        totalMatches: 1,
        stats: [
          {
            questionId: 'q-demo',
            questionText: 'Pregunta d’exemple',
            questionIndex: 0,
            hitRate: 0.8,
            avgResponseTimeMs: 4000,
            answerDistribution: [
              { answerId: 'a1', answerText: 'Opció A', isCorrect: true, count: 8, percentage: 0.8 },
              { answerId: 'a2', answerText: 'Opció B', isCorrect: false, count: 2, percentage: 0.2 }
            ],
            noAnswerCount: 0
          }
        ],
        students: [
          {
            studentId: 's-demo-1',
            studentName: 'Alumne d’Exemple',
            matchesCount: 1,
            calculatedGrade: 8.0,
            finalGrade: null,
            isGraded: false
          }
        ]
      }
    }
  }
  const response = await apiClient.get<QuizEvaluationResponse>(`/evaluations/quizzes/${quizId}`)
  return response.data
}

export async function getStudentEvaluation(
  quizId: string,
  studentId: string
): Promise<StudentEvaluationDetail> {
  if (isMockEnabled) {
    try {
      const response = await apiClient.get<StudentEvaluationDetail>(
        `/evaluations/quizzes/${quizId}/students/${studentId}`
      )
      return response.data
    } catch {
      const key = `${quizId}_${studentId}`
      const mock = mockStudentDetails[key]
      if (mock) return mock

      return {
        evaluationId: `eval-${studentId}`,
        studentId,
        studentName: 'Alumne d’Exemple',
        calculatedGrade: 7.5,
        finalGrade: null,
        isGraded: false,
        gradedBy: null,
        gradedAt: null,
        matches: [
          {
            matchId: 'm-demo',
            matchDate: new Date().toISOString(),
            score: 3,
            totalQuestions: 4,
            answers: [
              {
                questionId: 'q1',
                questionText: 'Pregunta 1',
                questionIndex: 0,
                selectedAnswerIds: ['ans1'],
                correctAnswerIds: ['ans1'],
                isCorrect: true,
                responseTimeMs: 3500
              }
            ]
          }
        ]
      }
    }
  }
  const response = await apiClient.get<StudentEvaluationDetail>(
    `/evaluations/quizzes/${quizId}/students/${studentId}`
  )
  return response.data
}

export async function gradeStudent(
  quizId: string,
  studentId: string,
  finalGrade: number
): Promise<GradeResponse> {
  if (finalGrade < 0 || finalGrade > 10) {
    throw new Error('La nota ha de ser entre 0.00 i 10.00.')
  }

  const payload: GradeRequest = { finalGrade }

  if (isMockEnabled) {
    try {
      const response = await apiClient.put<GradeResponse>(
        `/evaluations/quizzes/${quizId}/students/${studentId}/grade`,
        payload
      )
      return response.data
    } catch {
      const now = new Date().toISOString()
      return {
        evaluationId: `eval-${studentId}`,
        calculatedGrade: 8.0,
        finalGrade,
        isGraded: true,
        gradedBy: 'teacher-1',
        gradedAt: now
      }
    }
  }

  const response = await apiClient.put<GradeResponse>(
    `/evaluations/quizzes/${quizId}/students/${studentId}/grade`,
    payload
  )
  return response.data
}
