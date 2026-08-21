export type QuizStatus = 'draft' | 'published' | 'archived'

export type QuestionType = 'single_choice' | 'multiple_choice'

export const TIME_LIMIT_OPTIONS = [5, 10, 20, 30, 60, 90, 120] as const
export type TimeLimitSeconds = (typeof TIME_LIMIT_OPTIONS)[number]

export interface QuizAnswer {
  id: string
  text: string
  isCorrect: boolean
  orderIndex: number
}

export interface QuizQuestion {
  id: string
  text: string
  imageUrl?: string | null
  questionType: QuestionType
  timeLimitSeconds: number
  orderIndex: number
  answers: QuizAnswer[]
}

export interface Quiz {
  id: string
  creatorId: string
  creatorName?: string
  title: string
  description?: string | null
  coverImageUrl?: string | null
  status: QuizStatus
  tags: string[]
  questionCount: number
  createdAt: string
  updatedAt: string
}

export interface QuizDetail extends Quiz {
  questions: QuizQuestion[]
}

export interface SaveAnswerInput {
  id?: string | null
  text: string
  isCorrect: boolean
  orderIndex: number
}

export interface SaveQuestionInput {
  id?: string | null
  text: string
  imageUrl?: string | null
  questionType: QuestionType
  timeLimitSeconds: number
  orderIndex: number
  answers: SaveAnswerInput[]
}

export interface CreateQuizRequest {
  title: string
  description?: string | null
  coverImageUrl?: string | null
  status?: QuizStatus
  tags?: string[]
  questions?: SaveQuestionInput[]
}

export interface UpdateQuizRequest {
  title: string
  description?: string | null
  coverImageUrl?: string | null
  status?: QuizStatus
  tags?: string[]
  questions?: SaveQuestionInput[]
}

export interface DuplicateQuizRequest {
  includeAnswers?: boolean
  title?: string | null
}

export interface UploadImageResponse {
  url: string
  key: string
}

export interface QuizListResponse {
  items: Quiz[]
  pagination: {
    page: number
    pageSize: number
    totalCount: number
    totalPages: number
  }
}

export interface QuizFilters {
  page?: number
  pageSize?: number
  search?: string
  status?: QuizStatus
  tag?: string
}

export interface ErrorResponse {
  error: string
  code?: string
  details?: string[]
}

export interface KahootThemeShape {
  name: string
  symbol: string
  color: string
  bgLight: string
  borderColor: string
}

export const KAHOOT_THEME_SHAPES: KahootThemeShape[] = [
  {
    name: 'Triangle',
    symbol: '▲',
    color: '#e21b3c',
    bgLight: '#ffebee',
    borderColor: '#c62828'
  },
  {
    name: 'Rombe',
    symbol: '◆',
    color: '#1368ce',
    bgLight: '#e3f2fd',
    borderColor: '#1565c0'
  },
  {
    name: 'Cercle',
    symbol: '●',
    color: '#d89e00',
    bgLight: '#fffde7',
    borderColor: '#f57f17'
  },
  {
    name: 'Quadrat',
    symbol: '■',
    color: '#26890c',
    bgLight: '#e8f5e9',
    borderColor: '#2e7d32'
  },
  {
    name: 'Estrella',
    symbol: '★',
    color: '#864cbf',
    bgLight: '#f3e5f5',
    borderColor: '#6a1b9a'
  },
  {
    name: 'Hexàgon',
    symbol: '⬡',
    color: '#e67e22',
    bgLight: '#fff3e0',
    borderColor: '#d35400'
  }
]
