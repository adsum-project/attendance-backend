package timetable

import (
	"context"
	"errors"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/authorization"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/validation"
)

func (t *TimetableService) GetModuleCourses(ctx context.Context, moduleID string) ([]timetablemodels.ModuleCourse, error) {
	_, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return nil, errs.NotFound("Module not found")
		}
		return nil, err
	}
	return t.repo.GetModuleCourses(ctx, moduleID)
}

func (t *TimetableService) GetCourseModules(ctx context.Context, courseID string) ([]timetablemodels.CourseModule, error) {
	_, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return nil, errs.NotFound("Course not found")
		}
		return nil, err
	}
	return t.repo.GetCourseModules(ctx, courseID)
}

func (t *TimetableService) AssignModuleToCourse(ctx context.Context, courseID, moduleID string, yearOfStudy int) error {
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

	_, err = t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return errs.NotFound("Module not found")
		}
		return err
	}

	var v validation.Errors
	v.Add(validation.IntRange(yearOfStudy, "yearOfStudy", 1, 6))
	if err := v.Result(); err != nil {
		return err
	}

	exists, err := t.repo.CourseModuleExists(ctx, courseID, moduleID)
	if err != nil {
		return err
	}
	if exists {
		return errs.Error(409, "module is already assigned to this course")
	}

	return t.repo.CreateCourseModule(ctx, courseID, moduleID, yearOfStudy)
}

func (t *TimetableService) UnassignModuleFromCourse(ctx context.Context, courseID, moduleID string) error {
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

	err = t.repo.DeleteCourseModule(ctx, courseID, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseModuleNotFound) {
			return errs.NotFound("Module assignment not found")
		}
		return err
	}
	return nil
}
