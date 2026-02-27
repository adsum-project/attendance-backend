package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	"github.com/adsum-project/attendance-backend/internal/services/timetable"
	"github.com/adsum-project/attendance-backend/pkg/broadcaster"
	"github.com/adsum-project/attendance-backend/pkg/utils"
)

const (
	methodFR = "fr"
	methodQR = "qr"
)

var ErrNoMatch = errors.New("no match")

type VerificationService struct {
	client           *http.Client
	embeddingsURL    string
	verificationRepo *verificationrepo.VerificationRepository
	timetableService *timetable.TimetableService
	qrBroadcaster *broadcaster.Broadcaster
}

func NewVerificationService(verificationRepo *verificationrepo.VerificationRepository, timetableService *timetable.TimetableService) (*VerificationService, error) {
	embeddingsURL := strings.TrimSpace(os.Getenv("EMBEDDINGS_API_URL"))
	if embeddingsURL == "" {
		return nil, fmt.Errorf("EMBEDDINGS_API_URL environment variable is required")
	}
	if timetableService == nil {
		return nil, fmt.Errorf("timetable service is required")
	}
	return &VerificationService{
		client:           &http.Client{Timeout: 15 * time.Second},
		embeddingsURL:    embeddingsURL,
		verificationRepo: verificationRepo,
		timetableService: timetableService,
		qrBroadcaster:    broadcaster.New(),
	}, nil
}

func (s *VerificationService) VerifyEmbedding(ctx context.Context, imageBase64, classID string) (*VerifyData, error) {
	body := map[string]string{"imageBase64": imageBase64}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	res, err := utils.RequestUpstream(
		ctx,
		s.client,
		http.MethodPost,
		s.embeddingsURL,
		"/embeddings/verify",
		url.Values{},
		http.Header{"Content-Type": []string{"application/json"}},
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("upstream request: %w", err)
	}

	var upstream EmbeddingVerifyResponse
	if err := json.Unmarshal(res.Body, &upstream); err != nil {
		return nil, fmt.Errorf("invalid upstream response: %w", err)
	}

	if res.StatusCode == http.StatusNotFound || !upstream.Success || upstream.Data == nil {
		return nil, ErrNoMatch
	}

	if err := s.timetableService.CanStudentSignIntoClass(ctx, upstream.Data.UserID, classID); err != nil {
		return nil, err
	}

	if err := s.verificationRepo.InsertRecord(ctx, upstream.Data.UserID, classID, methodFR); err != nil {
		return nil, fmt.Errorf("insert attendance record: %w", err)
	}

	return upstream.Data, nil
}

func (s *VerificationService) IssueQRToken(ctx context.Context, classID string) (string, error) {
	expiresAt := time.Now().UTC().Add(2 * time.Minute)
	token, err := s.verificationRepo.CreateSignInToken(ctx, classID, expiresAt)
	if err != nil {
		return "", err
	}
	s.qrBroadcaster.Broadcast(classID, token)
	return token, nil
}

func (s *VerificationService) SignInWithQRToken(ctx context.Context, userID, token string) error {
	classID, err := s.verificationRepo.ConsumeToken(ctx, token)
	if err != nil {
		return err
	}
	if err := s.timetableService.CanStudentSignIntoClass(ctx, userID, classID); err != nil {
		return err
	}
	if err := s.verificationRepo.InsertRecord(ctx, userID, classID, methodQR); err != nil {
		return fmt.Errorf("insert attendance record: %w", err)
	}
	_, err = s.IssueQRToken(ctx, classID)
	return err
}

func (s *VerificationService) QRTokenStream(classID string) (chan string, func()) {
	ch := s.qrBroadcaster.Subscribe(classID)
	unsub := func() { s.qrBroadcaster.Unsubscribe(classID, ch) }
	return ch, unsub
}

func (s *VerificationService) IsQRTokenValid(ctx context.Context, token string) (bool, error) {
	return s.verificationRepo.IsTokenValid(ctx, token)
}
