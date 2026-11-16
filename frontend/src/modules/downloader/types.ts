export interface VideoFormat {
  formatId: string
  resolution?: string
  extension: string
  filesize?: number
}

export interface VideoInfo {
  id: string
  title: string
  thumbnail: string
  duration: number
  extractor: string
  uploader?: string
  description?: string
  formats?: VideoFormat[]
}

export interface VideoInfoRequest {
  url: string
}

export interface DownloadRequest {
  url: string
}
