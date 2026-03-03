package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const classesTable = "classes"

var ErrClassNotFound = errors.New("class not found")

func (r *TimetableRepository) GetClasses(ctx context.Context, moduleID string, page, perPage int) ([]timetablemodels.Class, error) {
	baseQuery := `SELECT ` + query.Guid("class_id") + ` as class_id, ` + query.Guid("module_id") + ` as module_id, class_name, ` + query.Room("room") + ` as room,
		day_of_week, ` + query.Time("starts_at") + ` as starts_at, ` + query.Time("ends_at") + ` as ends_at,
		recurrence,
		` + query.DateTimeISO("created_at") + ` as created_at, ` + query.DateTimeISO("updated_at") + ` as updated_at
		FROM ` + classesTable + `
		WHERE ` + query.GuidWhere("module_id", "@p1") + `
		ORDER BY day_of_week, starts_at`
	args := []any{moduleID}
	if page > 0 && perPage > 0 {
		baseQuery += ` OFFSET @p2 ROWS FETCH NEXT @p3 ROWS ONLY`
		args = append(args, (page-1)*perPage, perPage)
	}
	var classes []timetablemodels.Class
	err := r.db.SelectContext(ctx, &classes, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	return classes, nil
}

func (r *TimetableRepository) GetClassesCount(ctx context.Context, moduleID string) (int, error) {
	var total int
	q := `SELECT COUNT(*) FROM ` + classesTable + ` WHERE ` + query.GuidWhere("module_id", "@p1")
	err := r.db.QueryRowContext(ctx, q, moduleID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count classes: %w", err)
	}
	return total, nil
}

func (r *TimetableRepository) GetClassByID(ctx context.Context, moduleID, classID string) (*timetablemodels.Class, error) {
	var class timetablemodels.Class
	q := `SELECT ` + query.Guid("class_id") + ` as class_id, ` + query.Guid("module_id") + ` as module_id, class_name, ` + query.Room("room") + ` as room,
		day_of_week, ` + query.Time("starts_at") + ` as starts_at, ` + query.Time("ends_at") + ` as ends_at,
		recurrence,
		` + query.DateTimeISO("created_at") + ` as created_at, ` + query.DateTimeISO("updated_at") + ` as updated_at
		FROM ` + classesTable + `
		WHERE ` + query.GuidWhere("module_id", "@p1") + ` AND ` + query.GuidWhere("class_id", "@p2")
	err := r.db.GetContext(ctx, &class, q, moduleID, classID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClassNotFound
		}
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	return &class, nil
}

func (r *TimetableRepository) HasRoomConflict(ctx context.Context, room string, dayOfWeek int, startsAt, endsAt string, excludeClassID string) (bool, error) {
	queryStr := `SELECT 1 FROM ` + classesTable + `
		WHERE room = @p1 AND day_of_week = @p2
		AND starts_at < CAST(@p3 AS TIME) AND ends_at > CAST(@p4 AS TIME)`
	args := []any{room, dayOfWeek, endsAt, startsAt}
	if excludeClassID != "" {
		queryStr += ` AND class_id != CONVERT(uniqueidentifier, @p5)`
		args = append(args, excludeClassID)
	}
	var exists int
	err := r.db.GetContext(ctx, &exists, queryStr, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check room conflict: %w", err)
	}
	return true, nil
}

func (r *TimetableRepository) CreateClass(ctx context.Context, moduleID, className, room string, dayOfWeek int, startsAt, endsAt, recurrence string) (string, error) {
	var classID string
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO `+classesTable+` (class_id, module_id, class_name, room, day_of_week, starts_at, ends_at, recurrence, created_at, updated_at)
		OUTPUT `+query.Guid("INSERTED.class_id")+`
		VALUES (NEWID(), CONVERT(uniqueidentifier, @p1), @p2, @p3, @p4, CAST(@p5 AS TIME), CAST(@p6 AS TIME), @p7, SYSUTCDATETIME(), SYSUTCDATETIME())`,
		moduleID, className, room, dayOfWeek, startsAt, endsAt, recurrence,
	).Scan(&classID)
	if err != nil {
		return "", fmt.Errorf("failed to create class: %w", err)
	}
	return classID, nil
}

func (r *TimetableRepository) UpdateClass(ctx context.Context, moduleID, classID string, className, room *string, dayOfWeek *int, startsAt, endsAt, recurrence *string) error {
	clause, args, nextParam := query.UpdateAndCast(map[string]any{
		"class_name":  className,
		"room":        room,
		"day_of_week": dayOfWeek,
		"starts_at":   startsAt,
		"ends_at":     endsAt,
		"recurrence":  recurrence,
	}, map[string]string{
		"starts_at": "TIME",
		"ends_at":   "TIME",
	})
	if clause == "" {
		return errors.New("no fields to update")
	}
	nextParam2 := "@p" + fmt.Sprint(len(args)+2)
	result, err := r.db.ExecContext(ctx,
		`UPDATE `+classesTable+` SET `+clause+`, updated_at = SYSUTCDATETIME() WHERE `+query.GuidWhere("module_id", nextParam)+` AND `+query.GuidWhere("class_id", nextParam2),
		append(args, moduleID, classID)...,
	)
	if err != nil {
		return fmt.Errorf("failed to update class: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrClassNotFound
	}
	return nil
}

func (r *TimetableRepository) DeleteClass(ctx context.Context, moduleID, classID string) error {
	q := `DELETE FROM ` + classesTable + ` WHERE ` + query.GuidWhere("module_id", "@p1") + ` AND ` + query.GuidWhere("class_id", "@p2")
	result, err := r.db.ExecContext(ctx, q, moduleID, classID)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrClassNotFound
	}
	return nil
}
