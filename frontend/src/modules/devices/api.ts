/**
 * Crides HTTP per al mòdul d'autorització de dispositius (Devices).
 * Utilitza el client Axios centralitzat src/api/client.ts i respecta contracts/device_auth.openapi.yaml.
 */
import apiClient from '@/api/client'
import type { DeviceSummary } from './types'

/**
 * Llistar tots els dispositius autoritzats (només admin)
 */
export async function listDevices(): Promise<DeviceSummary[]> {
  const response = await apiClient.get<DeviceSummary[]>('/admin/devices')
  return response.data
}

/**
 * Revocar o desactivar un dispositiu autoritzat (només admin)
 */
export async function revokeDevice(id: string): Promise<void> {
  await apiClient.delete(`/admin/devices/${id}`)
}

/**
 * Descarregar executable de l'agent Encertia o certificat públic (només admin)
 */
export async function downloadAgent(platform: string): Promise<Blob> {
  const response = await apiClient.get(`/admin/devices/download/${platform}`, {
    responseType: 'blob'
  })
  return response.data
}
