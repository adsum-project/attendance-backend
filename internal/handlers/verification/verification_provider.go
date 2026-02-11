package verificationhandlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type VerificationProvider struct {
	embeddingsApiURL string
	client           *http.Client
}

func NewVerificationProvider() (*VerificationProvider, error) {
	embeddingsApiURL := strings.TrimSpace(os.Getenv("EMBEDDINGS_API_URL"))
	if embeddingsApiURL == "" {
		return nil, fmt.Errorf("EMBEDDINGS_API_URL environment variable is required")
	}

	return &VerificationProvider{
		embeddingsApiURL: embeddingsApiURL,
		client:           &http.Client{Timeout: 15 * time.Second},
	}, nil
}
