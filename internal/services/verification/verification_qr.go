package verification

import (
	"context"
	"fmt"
	"time"

	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
)

// IssueQRToken creates a short-lived sign-in token and broadcasts it to QR subscribers.
func (s *VerificationService) IssueQRToken(ctx context.Context, classID string) (string, error) {
	expiresAt := time.Now().UTC().Add(30 * time.Second)
	token, err := s.verificationRepo.CreateSignInToken(ctx, classID, expiresAt)
	if err != nil {
		return "", err
	}
	s.qrBroadcaster.Broadcast(classID, token)
	return token, nil
}

// SignInWithQRToken validates token, checks enrollment, and records attendance.
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
