package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	"github.com/adsum-project/attendance-backend/pkg/utils"
)

// VerifyEmbedding sends the image to the embeddings API and records attendance on match.
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

	if res.StatusCode == http.StatusBadRequest {
		return nil, ErrNoFaceDetected
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
	occurrenceDate := time.Now().UTC().Format("2006-01-02")
	already, err := s.verificationRepo.HasSignedInForOccurrence(ctx, upstream.Data.UserID, classID, occurrenceDate)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, verificationrepo.ErrAlreadySignedIn
	}
	if err := s.verificationRepo.InsertRecord(ctx, upstream.Data.UserID, classID, methodFR); err != nil {
		return nil, fmt.Errorf("insert attendance record: %w", err)
	}

	return upstream.Data, nil
}
