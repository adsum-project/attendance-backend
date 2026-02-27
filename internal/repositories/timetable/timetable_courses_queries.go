package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const coursesTable = "courses"

var ErrCourseNotFound = errors.New("course not found")

func (r *TimetableRepository) GetCourses(ctx context.Context, page, perPage int) ([]timetablemodels.Course, error) {
	var courses []timetablemodels.Course
	offset := (page - 1) * perPage
	err := r.db.SelectContext(
		ctx,
		&courses,
		`SELECT `+query.Guid("course_id")+` as course_id, course_code, course_name, campus, `+query.Guid("owner_id")+` as owner_id, created_at, updated_at
		FROM `+coursesTable+`
		ORDER BY created_at
		OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY`,
		offset,
		perPage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %w", err)
	}
	return courses, nil
}

func (r *TimetableRepository) GetCourseByID(ctx context.Context, courseID string) (*timetablemodels.Course, error) {
	var course timetablemodels.Course
	err := r.db.GetContext(
		ctx,
		&course,
		`SELECT `+query.Guid("course_id")+` as course_id, course_code, course_name, campus, `+query.Guid("owner_id")+` as owner_id, created_at, updated_at
		FROM `+coursesTable+`
		WHERE `+query.Guid("course_id")+` = LOWER(@p1)`,
		courseID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCourseNotFound
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}
	return &course, nil
}

func (r *TimetableRepository) GetCoursesByUserId(ctx context.Context, userID string) ([]timetablemodels.Course, error) {
	var courses []timetablemodels.Course
	err := r.db.SelectContext(
		ctx,
		&courses,
		`SELECT `+query.Guid("c.course_id")+` as course_id, c.course_code, c.course_name, c.campus, `+query.Guid("c.owner_id")+` as owner_id, c.created_at, c.updated_at
		FROM `+courseStudentsTable+` cs
		INNER JOIN `+coursesTable+` c ON cs.course_id = c.course_id
		WHERE `+query.Guid("cs.user_id")+` = LOWER(@p1)
		ORDER BY c.course_code`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses for user: %w", err)
	}
	return courses, nil
}

func (r *TimetableRepository) GetCoursesCount(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+coursesTable).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count courses: %w", err)
	}
	return total, nil
}

func (r *TimetableRepository) CourseCodeExists(ctx context.Context, courseCode string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+coursesTable+` WHERE course_code = @p1`,
		courseCode,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check course code: %w", err)
	}
	return true, nil
}

func (r *TimetableRepository) CreateCourse(ctx context.Context, courseCode, courseName, campus, ownerID string) (string, error) {
	var courseID string
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO `+coursesTable+` (course_id, course_code, course_name, campus, owner_id, created_at, updated_at)
		OUTPUT `+query.Guid("INSERTED.course_id")+`
		VALUES (NEWID(), @p1, @p2, @p3, @p4, SYSUTCDATETIME(), SYSUTCDATETIME())`,
		courseCode,
		courseName,
		campus,
		ownerID,
	).Scan(&courseID)
	if err != nil {
		return "", fmt.Errorf("failed to create course: %w", err)
	}
	return courseID, nil
}

func (r *TimetableRepository) UpdateCourse(ctx context.Context, courseID string, courseCode, courseName, campus *string) error {
	clause, args, nextParam := query.Update(map[string]any{
		"course_code": courseCode,
		"course_name": courseName,
		"campus":      campus,
	})
	result, err := r.db.ExecContext(ctx,
		`UPDATE `+coursesTable+` SET `+clause+`, updated_at = SYSUTCDATETIME() WHERE `+query.Guid("course_id")+` = LOWER(`+nextParam+`)`,
		append(args, courseID)...,
	)
	if err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *TimetableRepository) DeleteCourse(ctx context.Context, courseID string) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM `+coursesTable+` WHERE `+query.Guid("course_id")+` = LOWER(@p1)`,
		courseID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCourseNotFound
	}
	return nil
}
