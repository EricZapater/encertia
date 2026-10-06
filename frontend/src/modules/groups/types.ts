export interface Group {
  id: string
  name: string
  academicYear: string
  courseId?: string | null
  courseTitle?: string | null
  teacherId: string
  studentCount?: number
  createdAt: string
  updatedAt: string
}

export interface CreateGroupRequest {
  name: string
  academicYear?: string
  courseId?: string | null
}

export interface UpdateGroupRequest {
  name?: string
  academicYear?: string
  courseId?: string | null
}

export interface GroupResponse {
  data: Group
}

export interface GroupListResponse {
  items: Group[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface GroupFilters {
  page?: number
  pageSize?: number
  search?: string
  academicYear?: string
  courseId?: string
}

export interface AddStudentsToGroupRequest {
  studentIds: string[]
}

export interface GroupStudent {
  id: string
  email: string
  fullName: string
  assignedAt: string
}

export interface GroupStudentListResponse {
  items: GroupStudent[]
  total: number
}
