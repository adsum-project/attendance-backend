package timetable

import (
	"context"
	"errors"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/authorization"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
	"github.com/adsum-project/attendance-backend/pkg/utils/validation"
)

func (t *TimetableService) CreateCourse(ctx context.Context, courseCode, courseName, campus string) (string, error) {
	ownerID, _ := ctx.Value("userID").(string)
	var v validation.Errors
	v.Add(validation.Required(courseCode, "courseCode"))
	v.Add(validation.DigitsOnly(courseCode, "courseCode"))
	v.Add(validation.ExactLength(courseCode, "courseCode", 4))
	v.Add(validation.Required(courseName, "courseName"))
	v.Add(validation.Alphanumeric(courseName, "courseName", true))
	v.Add(validation.LengthRange(courseName, "courseName", 1, 100))
	v.Add(validation.Required(campus, "campus"))
	v.Add(validation.LettersOnly(campus, "campus"))
	v.Add(validation.LengthRange(campus, "campus", 1, 20))
	if err := v.Result(); err != nil {
		return "", err
	}

	exists, err := t.repo.CourseCodeExists(ctx, courseCode)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errs.Error(409, "courseCode already exists")
	}

	return t.repo.CreateCourse(ctx, courseCode, courseName, campus, ownerID)
}

func (t *TimetableService) GetCourse(ctx context.Context, courseID string) (*timetablemodels.Course, error) {
	course, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return nil, errs.NotFound("Course not found")
		}
		return nil, err
	}
	return course, nil
}

func (t *TimetableService) GetCourses(ctx context.Context, page, perPage int) ([]timetablemodels.Course, int, int, int, error) {
	return pagination.Paginate(ctx, page, perPage, t.repo.GetCourses, t.repo.GetCoursesCount)
}

func (t *TimetableService) UpdateCourse(ctx context.Context, courseID string, courseCode, courseName, campus *string) error {
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
	if courseCode == nil && courseName == nil && campus == nil {
		return errs.BadRequest("at least one field must be provided")
	}

	var v validation.Errors
	if courseCode != nil {
		v.Add(validation.Required(*courseCode, "courseCode"))
		v.Add(validation.DigitsOnly(*courseCode, "courseCode"))
		v.Add(validation.ExactLength(*courseCode, "courseCode", 4))
	}
	if courseName != nil {
		v.Add(validation.Required(*courseName, "courseName"))
		v.Add(validation.Alphanumeric(*courseName, "courseName", true))
		v.Add(validation.LengthRange(*courseName, "courseName", 1, 100))
	}
	if campus != nil {
		v.Add(validation.Required(*campus, "campus"))
		v.Add(validation.LettersOnly(*campus, "campus"))
		v.Add(validation.LengthRange(*campus, "campus", 1, 20))
	}
	if err := v.Result(); err != nil {
		return err
	}

	if courseCode != nil && *courseCode != course.CourseCode {
		exists, err := t.repo.CourseCodeExists(ctx, *courseCode)
		if err != nil {
			return err
		}
		if exists {
			return errs.Error(409, "courseCode already exists")
		}
	}

	return t.repo.UpdateCourse(ctx, courseID, courseCode, courseName, campus)
}

func (t *TimetableService) DeleteCourse(ctx context.Context, courseID string) error {
	course, err := t.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrCourseNotFound) {
			return errs.NotFound("Course not found")
		}
		return err
	}
	if isOwner := authorization.IsOwnerOrAdmin(ctx, course.OwnerID); !isOwner {
		return errs.Forbidden("")
	}
	return t.repo.DeleteCourse(ctx, courseID)
}
