package verification

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	"github.com/adsum-project/attendance-backend/internal/services/graph"
	"github.com/adsum-project/attendance-backend/internal/services/timetable"
	"github.com/adsum-project/attendance-backend/pkg/broadcaster"
)

const (
	methodFR = "fr"
	methodQR = "qr"
)

var ErrNoMatch = errors.New("no match")
var ErrNoFaceDetected = errors.New("no face detected")

type VerificationService struct {
	client           *http.Client
	embeddingsURL    string
	verificationRepo *verificationrepo.VerificationRepository
	timetableService *timetable.TimetableService
	graph            *graph.GraphService
	qrBroadcaster    *broadcaster.Broadcaster
}

// NewVerificationService creates the verification service (QR sign-in, face recognition, attendance records).
func NewVerificationService(verificationRepo *verificationrepo.VerificationRepository, timetableService *timetable.TimetableService, graphService *graph.GraphService) (*VerificationService, error) {
	embeddingsURL := strings.TrimSpace(os.Getenv("EMBEDDINGS_API_URL"))
	if embeddingsURL == "" {
		return nil, fmt.Errorf("EMBEDDINGS_API_URL environment variable is required")
	}
	if timetableService == nil {
		return nil, fmt.Errorf("timetable service is required")
	}
	return &VerificationService{
		client:           &http.Client{Timeout: 60 * time.Second},
		embeddingsURL:    embeddingsURL,
		verificationRepo: verificationRepo,
		timetableService: timetableService,
		graph:            graphService, // nil ok: display names omitted for class records
		qrBroadcaster:    broadcaster.New(),
	}, nil
}
