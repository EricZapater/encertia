import apiClient from '@/api/client'
import type { VideoInfo, VideoInfoRequest, DownloadRequest } from './types'

/**
 * Obtenir metadades d'un vídeo des del servidor
 */
export async function fetchVideoInfo(url: string, cookies?: string): Promise<VideoInfo> {
  const payload: VideoInfoRequest = {
    url,
    cookies: cookies && cookies.trim() ? cookies.trim() : undefined
  }
  const response = await apiClient.post<VideoInfo>('/downloader/info', payload)
  return response.data
}

/**
 * Descarregar el fitxer de vídeo MP4 des del servidor
 */
export async function downloadVideoFile(url: string, cookies?: string): Promise<{ blob: Blob; filename: string }> {
  const payload: DownloadRequest = {
    url,
    cookies: cookies && cookies.trim() ? cookies.trim() : undefined
  }
  const response = await apiClient.post('/downloader/download', payload, {
    responseType: 'blob'
  })

  let filename = 'video.mp4'
  const disposition = response.headers['content-disposition']
  if (disposition) {
    const filenameMatch = disposition.match(/filename="?([^"]+)"?/)
    if (filenameMatch && filenameMatch[1]) {
      filename = filenameMatch[1]
    }
  }

  return { blob: response.data, filename }
}
