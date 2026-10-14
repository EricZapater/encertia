import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import { listDevices, revokeDevice, downloadAgent } from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    delete: vi.fn()
  }
}))

describe('Devices API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('listDevices calls GET /admin/devices', async () => {
    const mockDevices = [
      {
        id: 'dev-1',
        name: 'Portàtil Aula 3B',
        teacherId: 'teacher-1',
        teacherName: 'Joan Docent',
        teacherEmail: 'joan@encertia.cat',
        isActive: true,
        lastUsedAt: '2026-10-14T09:00:00Z',
        createdAt: '2026-09-01T08:00:00Z'
      }
    ]
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockDevices })

    const result = await listDevices()
    expect(apiClient.get).toHaveBeenCalledWith('/admin/devices')
    expect(result).toEqual(mockDevices)
  })

  it('revokeDevice calls DELETE /admin/devices/:id', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({ data: null })

    await revokeDevice('dev-123')
    expect(apiClient.delete).toHaveBeenCalledWith('/admin/devices/dev-123')
  })

  it('downloadAgent calls GET /admin/devices/download/:platform with responseType blob', async () => {
    const mockBlob = new Blob(['binary'], { type: 'application/octet-stream' })
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockBlob })

    const result = await downloadAgent('windows')
    expect(apiClient.get).toHaveBeenCalledWith('/admin/devices/download/windows', {
      responseType: 'blob'
    })
    expect(result).toEqual(mockBlob)
  })
})
