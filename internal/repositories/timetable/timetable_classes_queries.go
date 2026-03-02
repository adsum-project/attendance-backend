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

func (r *TimetableRepository) GetClasses(ctx context.Context, moduleID string) ([]timetablemodels.Class, error) {
	var classes []timetablemodels.Class
	err := r.db.SelectContext(
		ctx,
		&classes,
		`SELECT `+query.Guid("class_id")+` as class_id, `+query.Guid("module_id")+` as module_id, class_name, UPPER(LTRIM(RTRIM(room))) as room,
		day_of_week, CONVERT(VARCHAR(8), starts_at, 108) as starts_at, CONVERT(VARCHAR(8), ends_at, 108) as ends_at,
		recurrence,
		CONVERT(VARCHAR(33), created_at, 127) as created_at, CONVERT(VARCHAR(33), updated_at, 127) as updated_at
		FROM `+classesTable+`
		WHERE `+query.Guid("module_id")+` = LOWER(@p1)
		ORDER BY day_of_week, starts_at`,
		moduleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	return classes, nil
}

func (r *TimetableRepository) GetClassByID(ctx context.Context, moduleID, classID string) (*timetablemodels.Class, error) {
	var class timetablemodels.Class
	err := r.db.GetContext(
		ctx,
		&class,
		`SELECT `+query.Guid("class_id")+` as class_id, `+query.Guid("module_id")+` as module_id, class_name, UPPER(LTRIM(RTRIM(room))) as room,
		day_of_week, CONVERT(VARCHAR(8), starts_at, 108) as starts_at, CONVERT(VARCHAR(8), ends_at, 108) as ends_at,
		recurrence,
		CONVERT(VARCHAR(33), created_at, 127) as created_at, CONVERT(VARCHAR(33), updated_at, 127) as updated_at
		FROM `+classesTable+`
		WHERE `+query.Guid("module_id")+` = LOWER(@p1) AND `+query.Guid("class_id")+` = LOWER(@p2)`,
		moduleID,
		classID,
	)
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
		queryStr += ` AND ` + query.Guid("class_id") + ` != LOWER(@p5)`
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
		`UPDATE `+classesTable+` SET `+clause+`, updated_at = SYSUTCDATETIME() WHERE `+query.Guid("module_id")+` = LOWER(`+nextParam+`) AND `+query.Guid("class_id")+` = LOWER(`+nextParam2+`)`,
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
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM `+classesTable+` WHERE `+query.Guid("module_id")+` = LOWER(@p1) AND `+query.Guid("class_id")+` = LOWER(@p2)`,
		moduleID,
		classID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrClassNotFound
	}
	return nil
}
