package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	verificationmodels "github.com/adsum-project/attendance-backend/internal/models/verification"
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

// ErrNoFaceDetected is returned when the FR backend returns 400 (e.g. no face, low confidence, multiple faces).
var ErrNoFaceDetected = errors.New("no face detected")

type VerificationService struct {
	client           *http.Client
	embeddingsURL    string
	verificationRepo *verificationrepo.VerificationRepository
	timetableService *timetable.TimetableService
	qrBroadcaster    *broadcaster.Broadcaster
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
		client:           &http.Client{Timeout: 60 * time.Second},
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

func (s *VerificationService) IssueQRToken(ctx context.Context, classID string) (string, error) {
	expiresAt := time.Now().UTC().Add(30 * time.Second)
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
	occurrenceDate := time.Now().UTC().Format("2006-01-02")
	already, err := s.verificationRepo.HasSignedInForOccurrence(ctx, userID, classID, occurrenceDate)
	if err != nil {
		return err
	}
	if already {
		return verificationrepo.ErrAlreadySignedIn
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

func (s *VerificationService) CleanupExpiredQRTokens(ctx context.Context) (int64, error) {
	return s.verificationRepo.DeleteExpiredQRTokens(ctx)
}

func (s *VerificationService) GetOwnRecords(ctx context.Context, userID string) ([]*AttendanceRecord, error) {
	rows, err := s.verificationRepo.GetOwnRecords(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapRecords(rows), nil
}

func (s *VerificationService) GetRecordsByClass(ctx context.Context, classID string) ([]*AttendanceRecord, error) {
	rows, err := s.verificationRepo.GetRecordsByClass(ctx, classID)
	if err != nil {
		return nil, err
	}
	return mapRecords(rows), nil
}

func mapRecords(rows []verificationmodels.AttendanceRecord) []*AttendanceRecord {
	out := make([]*AttendanceRecord, len(rows))
	for i := range rows {
		status := rows[i].Status
		if status == "" {
			status = "present"
		}
		out[i] = &AttendanceRecord{
			RecordID:   rows[i].RecordID,
			UserID:     rows[i].UserID,
			ClassID:    rows[i].ClassID,
			ClassName:  rows[i].ClassName,
			ModuleCode: rows[i].ModuleCode,
			ModuleName: rows[i].ModuleName,
			Room:       rows[i].Room,
			SignedInAt: rows[i].SignedInAt,
			Method:     rows[i].Method,
			Status:     status,
		}
	}
	return out
}

func (s *VerificationService) RunQRTokenCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		ctx := context.Background()
		n, err := s.CleanupExpiredQRTokens(ctx)
		if err != nil {
			log.Printf("QR token cleanup: %v", err)
			continue
		}
		if n > 0 {
			log.Printf("QR token cleanup: removed %d expired tokens", n)
		}
	}
}

// ProcessEndedClasses creates absent records for enrolled students who did not sign in when their class ended.
func (s *VerificationService) ProcessEndedClasses(ctx context.Context) (int, error) {
	classes, err := s.timetableService.GetClassesEndedRecently(ctx)
	if err != nil {
		return 0, fmt.Errorf("get classes ended: %w", err)
	}
	var totalCreated int
	for _, c := range classes {
		userIDs, err := s.timetableService.GetEnrolledUserIDsForClass(ctx, c.ClassID)
		if err != nil {
			log.Printf("absent processor: get enrolled %s: %v", c.ClassID, err)
			continue
		}
		for _, userID := range userIDs {
			has, err := s.verificationRepo.HasRecordForClassOccurrence(ctx, userID, c.ClassID, c.OccurrenceDate)
			if err != nil {
				log.Printf("absent processor: check record %s %s: %v", userID, c.ClassID, err)
				continue
			}
			if has {
				continue
			}
			if err := s.verificationRepo.InsertAbsentRecord(ctx, userID, c.ClassID, c.OccurrenceEndAt); err != nil {
				log.Printf("absent processor: insert absent %s %s: %v", userID, c.ClassID, err)
				continue
			}
			totalCreated++
		}
	}
	return totalCreated, nil
}

// RunAbsentRecordProcessor runs every interval, creating absent records for ended classes.
func (s *VerificationService) RunAbsentRecordProcessor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		ctx := context.Background()
		n, err := s.ProcessEndedClasses(ctx)
		if err != nil {
			log.Printf("Absent record processor: %v", err)
			continue
		}
		if n > 0 {
			log.Printf("Absent record processor: created %d absent records", n)
		}
	}
}

// UpdateRecordStatus updates a record's status to excused or absent. Only allowed for records that are already absent/excused.
func (s *VerificationService) UpdateRecordStatus(ctx context.Context, recordID, status string) error {
	if status != "absent" && status != "excused" {
		return fmt.Errorf("status must be absent or excused")
	}
	return s.verificationRepo.UpdateRecordStatus(ctx, recordID, status)
}
