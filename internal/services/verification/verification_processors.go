package verification

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (s *VerificationService) CleanupExpiredQRTokens(ctx context.Context) (int64, error) {
	return s.verificationRepo.DeleteExpiredQRTokens(ctx)
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
