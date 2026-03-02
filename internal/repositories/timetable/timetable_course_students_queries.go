package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const courseStudentsTable = "course_students"

var ErrCourseStudentNotFound = errors.New("course student assignment not found")

func (r *TimetableRepository) CreateCourseStudent(ctx context.Context, courseID, userID string, yearOfStudy int) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO `+courseStudentsTable+` (course_id, user_id, year_of_study, created_at, updated_at)
		VALUES (CONVERT(uniqueidentifier, @p1), CONVERT(uniqueidentifier, @p2), @p3, SYSUTCDATETIME(), SYSUTCDATETIME())`,
		courseID,
		userID,
		yearOfStudy,
	)
	if err != nil {
		return fmt.Errorf("failed to assign student to course: %w", err)
	}
	return nil
}

func (r *TimetableRepository) DeleteCourseStudent(ctx context.Context, courseID, userID string) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM `+courseStudentsTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1) AND `+query.Guid("user_id")+` = LOWER(@p2)`,
		courseID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to unassign student from course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCourseStudentNotFound
	}
	return nil
}

func (r *TimetableRepository) GetCourseStudents(ctx context.Context, courseID string) ([]timetablemodels.CourseStudent, error) {
	var students []timetablemodels.CourseStudent
	err := r.db.SelectContext(
		ctx,
		&students,
		`SELECT `+query.Guid("course_id")+` as course_id, `+query.Guid("user_id")+` as user_id, year_of_study
		FROM `+courseStudentsTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1)
		ORDER BY user_id`,
		courseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get course students: %w", err)
	}
	return students, nil
}

func (r *TimetableRepository) CourseStudentExists(ctx context.Context, courseID, userID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+courseStudentsTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1) AND `+query.Guid("user_id")+` = LOWER(@p2)`,
		courseID,
		userID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check course student assignment: %w", err)
	}
	return true, nil
}
