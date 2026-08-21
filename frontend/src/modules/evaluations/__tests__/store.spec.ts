import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useEvaluationStore } from '../store'

vi.mock('../api', () => ({
  listEvaluations: vi.fn().mockResolvedValue({
    evaluations: [
      {
        quizId: 'q-test-1',
        quizTitle: 'Test Quiz',
        totalMatches: 2,
        totalStudents: 10,
        gradedCount: 5,
        lastMatchAt: '2026-08-21T10:00:00Z'
      }
    ]
  }),
  getQuizEvaluation: vi.fn().mockResolvedValue({
    quizId: 'q-test-1',
    quizTitle: 'Test Quiz',
    totalMatches: 2,
    stats: [],
    students: [
      {
        studentId: 's-1',
        studentName: 'Student 1',
        matchesCount: 1,
        calculatedGrade: 8.0,
        finalGrade: null,
        isGraded: false
      }
    ]
  }),
  getStudentEvaluation: vi.fn().mockResolvedValue({
    evaluationId: 'e-1',
    studentId: 's-1',
    studentName: 'Student 1',
    calculatedGrade: 8.0,
    finalGrade: null,
    isGraded: false,
    matches: []
  }),
  gradeStudent: vi.fn().mockResolvedValue({
    evaluationId: 'e-1',
    calculatedGrade: 8.0,
    finalGrade: 9.5,
    isGraded: true,
    gradedBy: 'teacher-1',
    gradedAt: '2026-08-21T12:00:00Z'
  })
}))

describe('useEvaluationStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('fetches evaluations list correctly', async () => {
    const store = useEvaluationStore()
    await store.fetchEvaluationsList()

    expect(store.evaluationsList.length).toBe(1)
    expect(store.evaluationsList[0].quizTitle).toBe('Test Quiz')
    expect(store.hasEvaluations).toBe(true)
  })

  it('fetches quiz evaluation correctly', async () => {
    const store = useEvaluationStore()
    await store.fetchQuizEvaluation('q-test-1')

    expect(store.activeQuizEvaluation).not.toBeNull()
    expect(store.activeQuizEvaluation?.students.length).toBe(1)
  })

  it('saves student grade correctly and updates reactive state', async () => {
    const store = useEvaluationStore()
    await store.fetchQuizEvaluation('q-test-1')
    await store.fetchStudentEvaluation('q-test-1', 's-1')

    await store.saveStudentGrade('q-test-1', 's-1', 9.5)

    expect(store.activeStudentEvaluation?.finalGrade).toBe(9.5)
    expect(store.activeStudentEvaluation?.isGraded).toBe(true)
    expect(store.activeQuizEvaluation?.students[0].finalGrade).toBe(9.5)
  })
})
