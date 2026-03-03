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
	q := `DELETE FROM ` + courseStudentsTable + ` WHERE ` + query.GuidWhere("course_id", "@p1") + ` AND ` + query.GuidWhere("user_id", "@p2")
	result, err := r.db.ExecContext(ctx, q, courseID, userID)
	if err != nil {
		return fmt.Errorf("failed to unassign student from course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCourseStudentNotFound
	}
	return nil
}

func (r *TimetableRepository) GetCourseStudents(ctx context.Context, courseID string, page, perPage int) ([]timetablemodels.CourseStudent, error) {
	baseQuery := `SELECT ` + query.Guid("course_id") + ` as course_id, ` + query.Guid("user_id") + ` as user_id, year_of_study
		FROM ` + courseStudentsTable + `
		WHERE ` + query.GuidWhere("course_id", "@p1") + `
		ORDER BY user_id`
	args := []any{courseID}
	if page > 0 && perPage > 0 {
		baseQuery += ` OFFSET @p2 ROWS FETCH NEXT @p3 ROWS ONLY`
		args = append(args, (page-1)*perPage, perPage)
	}
	var students []timetablemodels.CourseStudent
	err := r.db.SelectContext(ctx, &students, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get course students: %w", err)
	}
	return students, nil
}

func (r *TimetableRepository) GetCourseStudentsCount(ctx context.Context, courseID string) (int, error) {
	var total int
	q := `SELECT COUNT(*) FROM ` + courseStudentsTable + ` WHERE ` + query.GuidWhere("course_id", "@p1")
	err := r.db.QueryRowContext(ctx, q, courseID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count course students: %w", err)
	}
	return total, nil
}

func (r *TimetableRepository) CourseStudentExists(ctx context.Context, courseID, userID string) (bool, error) {
	var exists int
	q := `SELECT 1 FROM ` + courseStudentsTable + ` WHERE ` + query.GuidWhere("course_id", "@p1") + ` AND ` + query.GuidWhere("user_id", "@p2")
	err := r.db.QueryRowContext(ctx, q, courseID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check course student assignment: %w", err)
	}
	return true, nil
}
