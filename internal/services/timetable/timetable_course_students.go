package timetable

import (
	"context"
	"errors"
	"slices"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/authorization"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/validation"
)

func (t *TimetableService) GetCourseStudents(ctx context.Context, courseID string) ([]timetablemodels.CourseStudentEnrollment, error) {
	_, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return nil, errs.NotFound("Course not found")
		}
		return nil, err
	}

	assignments, err := t.repo.GetCourseStudents(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if len(assignments) == 0 {
		return []timetablemodels.CourseStudentEnrollment{}, nil
	}

	ids := make([]string, len(assignments))
	for i := range assignments {
		ids[i] = assignments[i].UserID
	}

	graphUsers, err := t.graph.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	userByID := make(map[string]*timetablemodels.CourseStudentEnrollment)
	for _, u := range graphUsers {
		userByID[u.ID] = &timetablemodels.CourseStudentEnrollment{
			UserID:      u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Mail,
		}
	}

	result := make([]timetablemodels.CourseStudentEnrollment, 0, len(assignments))
	for _, a := range assignments {
		if u, ok := userByID[a.UserID]; ok {
			u.YearOfStudy = a.YearOfStudy
			result = append(result, *u)
		}
	}
	return result, nil
}

func (t *TimetableService) AssignStudentToCourse(ctx context.Context, courseID, userID string, yearOfStudy int) error {
	course, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return errs.NotFound("Course not found")
		}
		return err
	}
	if !authorization.IsOwnerOrAdmin(ctx, course.OwnerID) {
		return errs.Forbidden("")
	}

	var v validation.Errors
	v.Add(validation.IntRange(yearOfStudy, "yearOfStudy", 1, 6))
	if err := v.Result(); err != nil {
		return err
	}

	graphUser, err := t.graph.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if graphUser == nil {
		return errs.NotFound("User not found in directory")
	}

	roles, err := t.graph.GetUserRoles(ctx, userID)
	if err != nil {
		return err
	}
	if slices.Contains(roles, "admin") || slices.Contains(roles, "staff") {
		return errs.BadRequest("Cannot assign admin or staff as student")
	}

	exists, err := t.repo.CourseStudentExists(ctx, courseID, userID)
	if err != nil {
		return err
	}
	if exists {
		return errs.Error(409, "student is already assigned to this course")
	}

	return t.repo.CreateCourseStudent(ctx, courseID, userID, yearOfStudy)
}

func (t *TimetableService) UnassignStudentFromCourse(ctx context.Context, courseID, userID string) error {
	course, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return errs.NotFound("Course not found")
		}
		return err
	}
	if !authorization.IsOwnerOrAdmin(ctx, course.OwnerID) {
		return errs.Forbidden("")
	}

	err = t.repo.DeleteCourseStudent(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseStudentNotFound) {
			return errs.NotFound("Student assignment not found")
		}
		return err
	}
	return nil
}
