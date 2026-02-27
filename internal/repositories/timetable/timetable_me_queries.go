package timetablerepo

import (
	"context"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

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
