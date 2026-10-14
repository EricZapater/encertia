import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { DeviceSummary } from './types'
import * as devicesApi from './api'

export const useDeviceStore = defineStore('devices', () => {
  // State
  const devices = ref<DeviceSummary[]>([])
  const isLoading = ref(false)
  const isRevoking = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const activeDevices = computed(() => devices.value.filter((d) => d.isActive))
  const revokedDevices = computed(() => devices.value.filter((d) => !d.isActive))
  const totalCount = computed(() => devices.value.length)

  // Actions
  function clearError() {
    error.value = null
  }

  async function fetchDevices(): Promise<DeviceSummary[]> {
    isLoading.value = true
    error.value = null
    try {
      const data = await devicesApi.listDevices()
      devices.value = data || []
      return devices.value
    } catch (err: any) {
      error.value =
        err.response?.data?.error?.message ||
        err.response?.data?.message ||
        err.message ||
        'Error en carregar els dispositius'
      devices.value = []
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function revokeDevice(id: string): Promise<void> {
    isRevoking.value = true
    error.value = null
    try {
      await devicesApi.revokeDevice(id)
      const target = devices.value.find((d) => d.id === id)
      if (target) {
        target.isActive = false
      }
    } catch (err: any) {
      error.value =
        err.response?.data?.error?.message ||
        err.response?.data?.message ||
        err.message ||
        'Error en revocar el dispositiu'
      throw err
    } finally {
      isRevoking.value = false
    }
  }

  return {
    devices,
    isLoading,
    isRevoking,
    error,
    activeDevices,
    revokedDevices,
    totalCount,
    clearError,
    fetchDevices,
    revokeDevice
  }
})
