package verification

type EmbeddingVerifyResponse struct {
	Success bool           `json:"success"`
	Data    *VerifyData    `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type VerifyData struct {
	UserID     string  `json:"userId"`
	Similarity float64 `json:"similarity"`
}
