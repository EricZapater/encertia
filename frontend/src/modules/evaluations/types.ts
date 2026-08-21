export interface EvaluationQuizSummary {
  quizId: string
  quizTitle: string
  totalMatches: number
  totalStudents: number
  gradedCount: number
  lastMatchAt: string
}

export interface EvaluationsListResponse {
  evaluations: EvaluationQuizSummary[]
}

export interface AnswerDistributionItem {
  answerId: string
  answerText: string
  isCorrect: boolean
  count: number
  percentage: number
}

export interface QuestionStats {
  questionId: string
  questionText: string
  questionIndex: number
  hitRate: number
  avgResponseTimeMs: number
  answerDistribution: AnswerDistributionItem[]
  noAnswerCount: number
}

export interface StudentEvaluationSummary {
  studentId: string
  studentName: string
  matchesCount: number
  calculatedGrade: number
  finalGrade?: number | null
  isGraded: boolean
}

export interface QuizEvaluationResponse {
  quizId: string
  quizTitle: string
  totalMatches: number
  stats: QuestionStats[]
  students: StudentEvaluationSummary[]
}

export interface StudentAnswerDetail {
  questionId: string
  questionText: string
  questionIndex: number
  selectedAnswerIds: string[]
  correctAnswerIds: string[]
  isCorrect: boolean
  responseTimeMs: number
}

export interface StudentMatchResult {
  matchId: string
  matchDate: string
  score: number
  totalQuestions: number
  answers: StudentAnswerDetail[]
}

export interface StudentEvaluationDetail {
  evaluationId: string
  studentId: string
  studentName: string
  calculatedGrade: number
  finalGrade?: number | null
  isGraded: boolean
  gradedBy?: string | null
  gradedAt?: string | null
  matches: StudentMatchResult[]
}

export interface GradeRequest {
  finalGrade: number
}

export interface GradeResponse {
  evaluationId: string
  calculatedGrade: number
  finalGrade: number
  isGraded: boolean
  gradedBy: string
  gradedAt: string
}

export interface EvaluationErrorDetails {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface EvaluationErrorResponse {
  error: EvaluationErrorDetails
}
