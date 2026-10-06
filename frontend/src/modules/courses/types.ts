export type CourseStatus = 'draft' | 'active' | 'archived'

export interface Course {
  id: string
  title: string
  code: string
  description?: string | null
  status: CourseStatus
  startDate?: string | null
  endDate?: string | null
  teacherId: string
  teacherName?: string | null
  enrolledStudentsCount?: number
  unitsCount?: number
  createdAt: string
  updatedAt: string
}

export interface CourseListResponse {
  items: Course[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface CourseFilters {
  page?: number
  pageSize?: number
  search?: string
  status?: CourseStatus
}

export interface CreateCourseRequest {
  title: string
  code: string
  description?: string
  status?: CourseStatus
  startDate?: string
  endDate?: string
}

export interface UpdateCourseRequest {
  title?: string
  code?: string
  description?: string
  status?: CourseStatus
  startDate?: string
  endDate?: string
}

export interface CourseUnit {
  id: string
  courseId: string
  title: string
  description?: string | null
  orderIndex: number
  quizzesCount?: number
  blocksCount?: number
  createdAt: string
  updatedAt: string
}

export interface CreateCourseUnitRequest {
  title: string
  description?: string
  orderIndex?: number
}

export interface UpdateCourseUnitRequest {
  title?: string
  description?: string
  orderIndex?: number
}

export interface LinkedQuiz {
  id: string
  title: string
  questionsCount?: number
}

export type ScriptBlockType = 'material' | 'quiz' | 'break'

export interface ScriptBlock {
  id: string
  unitId: string
  blockType: ScriptBlockType
  orderIndex: number
  title: string
  description?: string | null
  materialId?: string | null
  pdfUrl?: string | null
  startPage?: number | null
  endPage?: number | null
  quizId?: string | null
  quizTitle?: string | null
  durationMinutes?: number | null
  createdAt: string
}

export interface CreateScriptBlockRequest {
  blockType: ScriptBlockType
  orderIndex: number
  title: string
  description?: string
  materialId?: string
  pdfUrl?: string
  startPage?: number
  endPage?: number
  quizId?: string
  durationMinutes?: number
}

export interface CourseUnitDetail extends CourseUnit {
  linkedQuizzes?: LinkedQuiz[]
  scriptBlocks?: ScriptBlock[]
}

export interface CourseDetail extends Course {
  units?: CourseUnit[]
}

export interface EnrolledStudent {
  id: string
  firstName?: string
  lastName?: string
  email: string
  enrolledAt: string
}

export interface CourseStudentsResponse {
  courseId: string
  total: number
  students: EnrolledStudent[]
}

export interface EnrollStudentsRequest {
  studentIds: string[]
}
