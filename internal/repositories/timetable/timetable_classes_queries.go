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
		`SELECT `+query.Guid("class_id")+` as class_id, `+query.Guid("module_id")+` as module_id, class_name, room,
		CONVERT(VARCHAR(33), starts_at, 127) as starts_at, CONVERT(VARCHAR(33), ends_at, 127) as ends_at,
		recurrence, CONVERT(VARCHAR(10), until_date, 23) as until_date,
		CONVERT(VARCHAR(33), created_at, 127) as created_at, CONVERT(VARCHAR(33), updated_at, 127) as updated_at
		FROM `+classesTable+`
		WHERE `+query.Guid("module_id")+` = LOWER(@p1)
		ORDER BY starts_at`,
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
		`SELECT `+query.Guid("class_id")+` as class_id, `+query.Guid("module_id")+` as module_id, class_name, room,
		CONVERT(VARCHAR(33), starts_at, 127) as starts_at, CONVERT(VARCHAR(33), ends_at, 127) as ends_at,
		recurrence, CONVERT(VARCHAR(10), until_date, 23) as until_date,
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

func (r *TimetableRepository) CreateClass(ctx context.Context, moduleID, className, room, startsAt, endsAt, recurrence string, untilDate *string) (string, error) {
	var classID string
	var untilVal any
	if untilDate != nil && *untilDate != "" {
		untilVal = *untilDate
	}
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO `+classesTable+` (class_id, module_id, class_name, room, starts_at, ends_at, recurrence, until_date, created_at, updated_at)
		OUTPUT `+query.Guid("INSERTED.class_id")+`
		VALUES (NEWID(), CONVERT(uniqueidentifier, @p1), @p2, @p3, CONVERT(datetime2, @p4), CONVERT(datetime2, @p5), @p6, TRY_CAST(@p7 AS DATE), SYSUTCDATETIME(), SYSUTCDATETIME())`,
		moduleID, className, room, startsAt, endsAt, recurrence, untilVal,
	).Scan(&classID)
	if err != nil {
		return "", fmt.Errorf("failed to create class: %w", err)
	}
	return classID, nil
}

func (r *TimetableRepository) UpdateClass(ctx context.Context, moduleID, classID string, className, room, startsAt, endsAt, recurrence *string, untilDate *string) error {
	clause, args, _ := query.UpdateAndCast(map[string]any{
		"class_name": className,
		"room":       room,
		"starts_at":  startsAt,
		"ends_at":    endsAt,
		"recurrence": recurrence,
		"until_date": untilDate,
	}, map[string]string{"starts_at": "datetime2", "ends_at": "datetime2", "until_date": "DATE"})
	p1 := len(args) + 1
	p2 := len(args) + 2
	result, err := r.db.ExecContext(ctx,
		`UPDATE `+classesTable+` SET `+clause+`, updated_at = SYSUTCDATETIME() WHERE `+query.Guid("module_id")+` = LOWER(@p`+fmt.Sprint(p1)+`) AND `+query.Guid("class_id")+` = LOWER(@p`+fmt.Sprint(p2)+`)`,
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
