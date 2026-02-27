package verificationhandlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adsum-project/attendance-backend/internal/services/verification"
)

type VerificationProvider struct {
	embeddingsApiURL   string
	client             *http.Client
	verificationService *verification.VerificationService
}

func NewVerificationProvider(svc *verification.VerificationService) (*VerificationProvider, error) {
	embeddingsApiURL := strings.TrimSpace(os.Getenv("EMBEDDINGS_API_URL"))
	if embeddingsApiURL == "" {
		return nil, fmt.Errorf("EMBEDDINGS_API_URL environment variable is required")
	}

	if svc == nil {
		return nil, fmt.Errorf("verification service is required")
	}

	return &VerificationProvider{
		embeddingsApiURL:   embeddingsApiURL,
		client:             &http.Client{Timeout: 15 * time.Second},
		verificationService: svc,
	}, nil
}
