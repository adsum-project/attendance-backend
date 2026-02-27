package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const nodeRoomTable = "node_room"

func (r *TimetableRepository) GetClassesByUserId(ctx context.Context, userID string) ([]timetablemodels.ClassTimetableItem, error) {
	var rows []timetablemodels.ClassTimetableItem
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT `+query.Guid("cl.class_id")+` as class_id,
		m.module_code, m.module_name,
		CONVERT(VARCHAR(10), m.start_date, 23) as module_start_date, CONVERT(VARCHAR(10), m.end_date, 23) as module_end_date,
		c.course_code, c.course_name,
		cl.class_name, cl.room, cl.day_of_week,
		CONVERT(VARCHAR(8), cl.starts_at, 108) as starts_at, CONVERT(VARCHAR(8), cl.ends_at, 108) as ends_at,
		cl.recurrence
		FROM `+courseStudentsTable+` cs
		INNER JOIN `+courseModulesTable+` cm ON cs.course_id = cm.course_id
		INNER JOIN `+modulesTable+` m ON cm.module_id = m.module_id
		INNER JOIN `+classesTable+` cl ON m.module_id = cl.module_id
		INNER JOIN `+coursesTable+` c ON cm.course_id = c.course_id
		WHERE `+query.Guid("cs.user_id")+` = LOWER(@p1)
		ORDER BY cl.day_of_week, cl.starts_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes for user: %w", err)
	}
	return rows, nil
}

func (r *TimetableRepository) GetClassesByRoom(ctx context.Context, room string, dayOfWeek int, currentTime string) ([]timetablemodels.ClassTimetableItem, error) {
	var rows []timetablemodels.ClassTimetableItem
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT `+query.Guid("cl.class_id")+` as class_id,
		m.module_code, m.module_name,
		CONVERT(VARCHAR(10), m.start_date, 23) as module_start_date, CONVERT(VARCHAR(10), m.end_date, 23) as module_end_date,
		c.course_code, c.course_name,
		cl.class_name, cl.room, cl.day_of_week,
		CONVERT(VARCHAR(8), cl.starts_at, 108) as starts_at, CONVERT(VARCHAR(8), cl.ends_at, 108) as ends_at,
		cl.recurrence
		FROM `+classesTable+` cl
		INNER JOIN `+modulesTable+` m ON cl.module_id = m.module_id
		INNER JOIN `+courseModulesTable+` cm ON m.module_id = cm.module_id
		INNER JOIN `+coursesTable+` c ON cm.course_id = c.course_id
		WHERE LOWER(LTRIM(RTRIM(cl.room))) = LOWER(LTRIM(RTRIM(@p1)))
		AND cl.day_of_week = @p2
		AND CAST(@p3 AS TIME) >= cl.starts_at AND CAST(@p3 AS TIME) <= cl.ends_at
		AND CAST(SYSUTCDATETIME() AS DATE) >= m.start_date AND CAST(SYSUTCDATETIME() AS DATE) <= m.end_date
		ORDER BY cl.starts_at`,
		room,
		dayOfWeek,
		currentTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes by room: %w", err)
	}
	return rows, nil
}

func (r *TimetableRepository) GetNodeRoomByUserID(ctx context.Context, userID string) (string, error) {
	var room string
	err := r.db.GetContext(
		ctx,
		&room,
		`SELECT room FROM `+nodeRoomTable+`
		WHERE `+query.Guid("user_id")+` = LOWER(@p1)`,
		userID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get node room: %w", err)
	}
	return room, nil
}

func (r *TimetableRepository) UpsertNodeRoom(ctx context.Context, userID, room string) error {
	room = strings.TrimSpace(room)
	if room == "" {
		return fmt.Errorf("room cannot be empty")
	}
	_, err := r.db.ExecContext(
		ctx,
		`MERGE `+nodeRoomTable+` AS target
		USING (SELECT CONVERT(uniqueidentifier, @p1) AS user_id) AS source
		ON target.user_id = source.user_id
		WHEN MATCHED THEN UPDATE SET room = @p2, updated_at = SYSUTCDATETIME()
		WHEN NOT MATCHED THEN INSERT (user_id, room, created_at, updated_at)
		VALUES (CONVERT(uniqueidentifier, @p1), @p2, SYSUTCDATETIME(), SYSUTCDATETIME());`,
		userID,
		room,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert node room: %w", err)
	}
	return nil
}

func (r *TimetableRepository) DeleteNodeRoom(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM `+nodeRoomTable+`
		WHERE `+query.Guid("user_id")+` = LOWER(@p1)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete node room: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}
	return nil
}

func (r *TimetableRepository) StudentEnrolledInClass(ctx context.Context, userID, classID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+classesTable+` cl
		INNER JOIN `+courseModulesTable+` cm ON cl.module_id = cm.module_id
		INNER JOIN `+courseStudentsTable+` cs ON cm.course_id = cs.course_id
		WHERE `+query.Guid("cl.class_id")+` = LOWER(@p1) AND `+query.Guid("cs.user_id")+` = LOWER(@p2)`,
		classID,
		userID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check student enrollment: %w", err)
	}
	return true, nil
}

func (r *TimetableRepository) ClassCurrentlyRunning(ctx context.Context, classID string, dayOfWeek int, currentTime string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+classesTable+` cl
		INNER JOIN `+modulesTable+` m ON cl.module_id = m.module_id
		WHERE `+query.Guid("cl.class_id")+` = LOWER(@p1)
		AND cl.day_of_week = @p2
		AND CAST(@p3 AS TIME) >= cl.starts_at AND CAST(@p3 AS TIME) <= cl.ends_at
		AND CAST(SYSUTCDATETIME() AS DATE) >= m.start_date AND CAST(SYSUTCDATETIME() AS DATE) <= m.end_date`,
		classID,
		dayOfWeek,
		currentTime,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check class running: %w", err)
	}
	return true, nil
}
