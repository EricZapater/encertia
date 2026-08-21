import type {
  EvaluationsListResponse,
  QuizEvaluationResponse,
  StudentEvaluationDetail
} from './types'

export const mockEvaluationsList: EvaluationsListResponse = {
  evaluations: [
    {
      quizId: '11111111-1111-1111-1111-111111111111',
      quizTitle: 'Matemàtiques — Tema 3',
      totalMatches: 2,
      totalStudents: 15,
      gradedCount: 8,
      lastMatchAt: '2026-08-21T10:30:00Z'
    },
    {
      quizId: '22222222-2222-2222-2222-222222222222',
      quizTitle: 'Ciències de la Natura — Unitat 1',
      totalMatches: 1,
      totalStudents: 12,
      gradedCount: 12,
      lastMatchAt: '2026-08-20T16:45:00Z'
    }
  ]
}

export const mockQuizEvaluations: Record<string, QuizEvaluationResponse> = {
  '11111111-1111-1111-1111-111111111111': {
    quizId: '11111111-1111-1111-1111-111111111111',
    quizTitle: 'Matemàtiques — Tema 3',
    totalMatches: 2,
    stats: [
      {
        questionId: 'q1',
        questionText: 'Quant és 2 + 2?',
        questionIndex: 0,
        hitRate: 0.85,
        avgResponseTimeMs: 3200,
        answerDistribution: [
          { answerId: 'a1', answerText: '4', isCorrect: true, count: 17, percentage: 0.85 },
          { answerId: 'a2', answerText: '3', isCorrect: false, count: 2, percentage: 0.1 },
          { answerId: 'a3', answerText: '5', isCorrect: false, count: 1, percentage: 0.05 }
        ],
        noAnswerCount: 0
      },
      {
        questionId: 'q2',
        questionText: 'Quina és l’arrel quadrada de 16?',
        questionIndex: 1,
        hitRate: 0.6,
        avgResponseTimeMs: 5400,
        answerDistribution: [
          { answerId: 'a4', answerText: '4', isCorrect: true, count: 12, percentage: 0.6 },
          { answerId: 'a5', answerText: '8', isCorrect: false, count: 6, percentage: 0.3 },
          { answerId: 'a6', answerText: '2', isCorrect: false, count: 2, percentage: 0.1 }
        ],
        noAnswerCount: 1
      }
    ],
    students: [
      {
        studentId: 's1',
        studentName: 'Joan Garcia',
        matchesCount: 2,
        calculatedGrade: 8.0,
        finalGrade: 8.5,
        isGraded: true
      },
      {
        studentId: 's2',
        studentName: 'Maria Soler',
        matchesCount: 1,
        calculatedGrade: 10.0,
        finalGrade: null,
        isGraded: false
      }
    ]
  }
}

export const mockStudentDetails: Record<string, StudentEvaluationDetail> = {
  '11111111-1111-1111-1111-111111111111_s1': {
    evaluationId: 'eval-1',
    studentId: 's1',
    studentName: 'Joan Garcia',
    calculatedGrade: 8.0,
    finalGrade: 8.5,
    isGraded: true,
    gradedBy: 'teacher-uuid',
    gradedAt: '2026-08-21T11:00:00Z',
    matches: [
      {
        matchId: 'm2',
        matchDate: '2026-08-21T10:30:00Z',
        score: 2,
        totalQuestions: 2,
        answers: [
          {
            questionId: 'q1',
            questionText: 'Quant és 2 + 2?',
            questionIndex: 0,
            selectedAnswerIds: ['a1'],
            correctAnswerIds: ['a1'],
            isCorrect: true,
            responseTimeMs: 2500
          },
          {
            questionId: 'q2',
            questionText: 'Quina és l’arrel quadrada de 16?',
            questionIndex: 1,
            selectedAnswerIds: ['a4'],
            correctAnswerIds: ['a4'],
            isCorrect: true,
            responseTimeMs: 4100
          }
        ]
      }
    ]
  }
}
