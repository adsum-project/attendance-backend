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
		`+query.Date("m.start_date")+` as module_start_date, `+query.Date("m.end_date")+` as module_end_date,
		c.course_code, c.course_name,
		cl.class_name, `+query.Room("cl.room")+` as room, cl.day_of_week,
		`+query.Time("cl.starts_at")+` as starts_at, `+query.Time("cl.ends_at")+` as ends_at,
		cl.recurrence
		FROM `+courseStudentsTable+` cs
		INNER JOIN `+courseModulesTable+` cm ON cs.course_id = cm.course_id AND cs.year_of_study = cm.year_of_study
		INNER JOIN `+modulesTable+` m ON cm.module_id = m.module_id
		INNER JOIN `+classesTable+` cl ON m.module_id = cl.module_id
		INNER JOIN `+coursesTable+` c ON cm.course_id = c.course_id
		WHERE `+query.GuidWhere("cs.user_id", "@p1")+`
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
		`+query.Date("m.start_date")+` as module_start_date, `+query.Date("m.end_date")+` as module_end_date,
		c.course_code, c.course_name,
		cl.class_name, `+query.Room("cl.room")+` as room, cl.day_of_week,
		`+query.Time("cl.starts_at")+` as starts_at, `+query.Time("cl.ends_at")+` as ends_at,
		cl.recurrence
		FROM `+classesTable+` cl
		INNER JOIN `+modulesTable+` m ON cl.module_id = m.module_id
		INNER JOIN `+courseModulesTable+` cm ON m.module_id = cm.module_id
		INNER JOIN `+coursesTable+` c ON cm.course_id = c.course_id
		WHERE `+query.Room("cl.room")+` = UPPER(LTRIM(RTRIM(@p1)))
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
	q := `SELECT ` + query.Room("room") + ` as room FROM ` + nodeRoomTable + ` WHERE ` + query.GuidWhere("user_id", "@p1")
	err := r.db.GetContext(ctx, &room, q, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get node room: %w", err)
	}
	return room, nil
}

func (r *TimetableRepository) UpsertNodeRoom(ctx context.Context, userID, room string) error {
	room = strings.ToUpper(strings.TrimSpace(room))
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
	q := `DELETE FROM ` + nodeRoomTable + ` WHERE ` + query.GuidWhere("user_id", "@p1")
	result, err := r.db.ExecContext(ctx, q, userID)
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
	q := `SELECT 1 FROM ` + classesTable + ` cl
		INNER JOIN ` + courseModulesTable + ` cm ON cl.module_id = cm.module_id
		INNER JOIN ` + courseStudentsTable + ` cs ON cm.course_id = cs.course_id AND cm.year_of_study = cs.year_of_study
		WHERE ` + query.GuidWhere("cl.class_id", "@p1") + ` AND ` + query.GuidWhere("cs.user_id", "@p2")
	err := r.db.QueryRowContext(ctx, q, classID, userID).Scan(&exists)
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
	q := `SELECT 1 FROM ` + classesTable + ` cl
		INNER JOIN ` + modulesTable + ` m ON cl.module_id = m.module_id
		WHERE ` + query.GuidWhere("cl.class_id", "@p1") + `
		AND cl.day_of_week = @p2
		AND CAST(@p3 AS TIME) >= cl.starts_at AND CAST(@p3 AS TIME) <= cl.ends_at
		AND CAST(SYSUTCDATETIME() AS DATE) >= m.start_date AND CAST(SYSUTCDATETIME() AS DATE) <= m.end_date`
	err := r.db.QueryRowContext(ctx, q, classID, dayOfWeek, currentTime).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check class running: %w", err)
	}
	return true, nil
}

// ClassEndedItem represents a class that has recently ended (for absent processing).
type ClassEndedItem struct {
	ClassID         string `db:"class_id"`
	EndsAt          string `db:"ends_at"`
	DayOfWeek       int    `db:"day_of_week"`
	OccurrenceDate  string `db:"occurrence_date"`
	OccurrenceEndAt string `db:"occurrence_end_at"`
}

// GetClassesEndedRecently returns classes whose occurrence ended in the last 10 minutes.
// windowStart and nowStr are ISO datetime strings. dayOfWeek: 1=Mon .. 7=Sun (matches class.day_of_week).
func (r *TimetableRepository) GetClassesEndedRecently(ctx context.Context, windowStart, nowStr string, dayOfWeek int) ([]ClassEndedItem, error) {
	var rows []ClassEndedItem
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT `+query.Guid("cl.class_id")+` as class_id,
		`+query.Time("cl.ends_at")+` as ends_at,
		cl.day_of_week,
		`+query.Date("CAST(SYSUTCDATETIME() AS DATE)")+` as occurrence_date,
		`+query.DateTimeISO("occurrence_end")+` as occurrence_end_at
		FROM `+classesTable+` cl
		INNER JOIN `+modulesTable+` m ON cl.module_id = m.module_id
		CROSS APPLY (
			SELECT CAST(CAST(CAST(SYSUTCDATETIME() AS DATE) AS DATETIME) + CAST(cl.ends_at AS DATETIME) AS DATETIME2) AS occurrence_end
		) oe
		WHERE cl.day_of_week = @p3
		AND occurrence_end >= CAST(@p1 AS DATETIME2) AND occurrence_end <= CAST(@p2 AS DATETIME2)
		AND CAST(SYSUTCDATETIME() AS DATE) >= m.start_date AND CAST(SYSUTCDATETIME() AS DATE) <= m.end_date`,
		windowStart, nowStr, dayOfWeek,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes ended recently: %w", err)
	}
	return rows, nil
}

// GetEnrolledUserIDsForClass returns all user IDs enrolled in the given class (via course-module-class).
func (r *TimetableRepository) GetEnrolledUserIDsForClass(ctx context.Context, classID string) ([]string, error) {
	var rows []struct {
		UserID string `db:"user_id"`
	}
	q := `SELECT ` + query.Guid("cs.user_id") + ` as user_id
		FROM ` + classesTable + ` cl
		INNER JOIN ` + courseModulesTable + ` cm ON cl.module_id = cm.module_id
		INNER JOIN ` + courseStudentsTable + ` cs ON cm.course_id = cs.course_id AND cm.year_of_study = cm.year_of_study
		WHERE ` + query.GuidWhere("cl.class_id", "@p1")
	err := r.db.SelectContext(ctx, &rows, q, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrolled users: %w", err)
	}
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = rows[i].UserID
	}
	return ids, nil
}
