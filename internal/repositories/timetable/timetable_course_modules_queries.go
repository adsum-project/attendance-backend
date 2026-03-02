package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const courseModulesTable = "course_modules"

var ErrCourseModuleNotFound = errors.New("course module assignment not found")

func (r *TimetableRepository) CreateCourseModule(ctx context.Context, courseID, moduleID string, yearOfStudy int) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO `+courseModulesTable+` (module_id, course_id, year_of_study, created_at, updated_at)
		VALUES (CONVERT(uniqueidentifier, @p1), CONVERT(uniqueidentifier, @p2), @p3, SYSUTCDATETIME(), SYSUTCDATETIME())`,
		moduleID,
		courseID,
		yearOfStudy,
	)
	if err != nil {
		return fmt.Errorf("failed to assign module to course: %w", err)
	}
	return nil
}

func (r *TimetableRepository) DeleteCourseModule(ctx context.Context, courseID, moduleID string) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM `+courseModulesTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1) AND `+query.Guid("module_id")+` = LOWER(@p2)`,
		courseID,
		moduleID,
	)
	if err != nil {
		return fmt.Errorf("failed to unassign module from course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCourseModuleNotFound
	}
	return nil
}

func (r *TimetableRepository) GetCourseModules(ctx context.Context, courseID string) ([]timetablemodels.CourseModule, error) {
	var modules []timetablemodels.CourseModule
	err := r.db.SelectContext(
		ctx,
		&modules,
		`SELECT `+query.Guid("m.module_id")+` as module_id, m.module_code, m.module_name, `+query.Guid("m.owner_id")+` as owner_id,
		CONVERT(VARCHAR(10), m.start_date, 23) as start_date, CONVERT(VARCHAR(10), m.end_date, 23) as end_date,
		cm.year_of_study
		FROM `+courseModulesTable+` cm
		INNER JOIN modules m ON cm.module_id = m.module_id
		WHERE `+query.Guid("cm.course_id")+` = LOWER(@p1)
		ORDER BY cm.year_of_study, m.module_code`,
		courseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get course modules: %w", err)
	}
	return modules, nil
}

func (r *TimetableRepository) GetModuleCourses(ctx context.Context, moduleID string) ([]timetablemodels.ModuleCourse, error) {
	var courses []timetablemodels.ModuleCourse
	err := r.db.SelectContext(
		ctx,
		&courses,
		`SELECT `+query.Guid("c.course_id")+` as course_id, c.course_code, c.course_name, c.campus, cm.year_of_study
		FROM `+courseModulesTable+` cm
		INNER JOIN courses c ON cm.course_id = c.course_id
		WHERE `+query.Guid("cm.module_id")+` = LOWER(@p1)
		ORDER BY cm.year_of_study, c.course_code`,
		moduleID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get module courses: %w", err)
	}
	return courses, nil
}

func (r *TimetableRepository) CourseModuleExists(ctx context.Context, courseID, moduleID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+courseModulesTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1) AND `+query.Guid("module_id")+` = LOWER(@p2)`,
		courseID,
		moduleID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check course module assignment: %w", err)
	}
	return true, nil
}
