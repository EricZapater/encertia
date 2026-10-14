/**
 * Tipus de dades per al mòdul d'autorització de dispositius (device_auth).
 * Respecta contracts/device_auth.openapi.yaml
 */

export interface DeviceSummary {
  id: string
  name: string
  teacherId: string
  teacherName: string
  teacherEmail: string
  isActive: boolean
  lastUsedAt: string | null
  createdAt: string
}

export type DevicePlatform = 'windows' | 'darwin' | 'linux' | 'cert'

export interface PlatformDownloadInfo {
  platform: DevicePlatform
  filename: string
  icon: string
  titleKey: string
  descKey: string
  instructionKey: string
  command?: string
  buttonKey: string
  badge?: string
}
