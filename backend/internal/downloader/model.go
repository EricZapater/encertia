package downloader

// VideoInfoRequest defineix la petició d'inspecció d'un enllaç de vídeo.
type VideoInfoRequest struct {
	URL string `json:"url" binding:"required"`
}

// VideoFormat representa una opció de qualitat disponible.
type VideoFormat struct {
	FormatID   string `json:"formatId"`
	Resolution string `json:"resolution,omitempty"`
	Extension  string `json:"extension"`
	Filesize   int64  `json:"filesize,omitempty"`
}

// VideoInfoResponse conté les metadades obtingudes del vídeo.
type VideoInfoResponse struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Thumbnail   string        `json:"thumbnail"`
	Duration    int           `json:"duration"` // Durada en segons
	Extractor   string        `json:"extractor"` // Plataforma: youtube, instagram, linkedin, etc.
	Uploader    string        `json:"uploader,omitempty"`
	Description string        `json:"description,omitempty"`
	Formats     []VideoFormat `json:"formats,omitempty"`
}

// DownloadRequest defineix la petició per descarregar el fitxer de vídeo.
type DownloadRequest struct {
	URL string `json:"url" binding:"required"`
}
