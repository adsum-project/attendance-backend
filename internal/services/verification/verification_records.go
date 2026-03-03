package verification

import (
	"context"
	"fmt"

	verificationmodels "github.com/adsum-project/attendance-backend/internal/models/verification"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
)

func (s *VerificationService) GetOwnRecords(ctx context.Context, userID string, page, perPage int) (*pagination.Result[*AttendanceRecord], error) {
	fetch, count := pagination.BindIDAndMap(userID, s.verificationRepo.GetOwnRecords, s.verificationRepo.GetOwnRecordsCount, recordsMapFn)
	return pagination.Paginate(ctx, page, perPage, fetch, count)
}

func (s *VerificationService) GetRecordsByClass(ctx context.Context, classID string, page, perPage int) (*pagination.Result[*AttendanceRecord], error) {
	page, perPage = pagination.Normalize(page, perPage)
	rows, err := s.verificationRepo.GetRecordsByClass(ctx, classID, page, perPage)
	if err != nil {
		return nil, err
	}
	total, err := s.verificationRepo.GetRecordsByClassCount(ctx, classID)
	if err != nil {
		return nil, err
	}
	records := mapRecords(rows)
	if len(records) > 0 && s.graph != nil {
		userIDs := make([]string, 0, len(records))
		seen := make(map[string]bool)
		for _, r := range records {
			if r.UserID != "" && !seen[r.UserID] {
				seen[r.UserID] = true
				userIDs = append(userIDs, r.UserID)
			}
		}
		if len(userIDs) > 0 {
			graphUsers, err := s.graph.GetUsersByIDs(ctx, userIDs)
			if err == nil {
				byID := make(map[string]string)
				for _, u := range graphUsers {
					byID[u.ID] = u.DisplayName
				}
				for _, r := range records {
					r.DisplayName = byID[r.UserID]
				}
			}
		}
	}
	return &pagination.Result[*AttendanceRecord]{Data: records, Total: total, Page: page, PerPage: perPage}, nil
}

func recordsMapFn(_ context.Context, rows []verificationmodels.AttendanceRecord) ([]*AttendanceRecord, error) {
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

// UpdateRecordStatus sets attendance record status to absent or excused.
func (s *VerificationService) UpdateRecordStatus(ctx context.Context, recordID, status string) error {
	if status != "absent" && status != "excused" {
		return fmt.Errorf("status must be absent or excused")
	}
	return s.verificationRepo.UpdateRecordStatus(ctx, recordID, status)
}
