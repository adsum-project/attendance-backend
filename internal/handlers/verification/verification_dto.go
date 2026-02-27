package verificationhandlers

type VerifyRequest struct {
	ImageBase64 string `json:"imageBase64"`
	ClassID     string `json:"classId"`
}
