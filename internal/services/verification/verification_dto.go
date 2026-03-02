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

type AttendanceRecord struct {
	RecordID   string `json:"recordId"`
	UserID     string `json:"userId,omitempty"`
	ClassID    string `json:"classId"`
	ClassName  string `json:"className"`
	ModuleCode string `json:"moduleCode"`
	ModuleName string `json:"moduleName"`
	Room       string `json:"room"`
	SignedInAt string `json:"signedInAt"`
	Method     string `json:"method"`
	Status     string `json:"status"`
}
