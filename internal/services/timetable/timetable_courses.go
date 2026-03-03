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

// CreateCourse creates a course owned by the authenticated user.
func (t *TimetableService) CreateCourse(ctx context.Context, courseCode, courseName, campus string) (string, error) {
	ownerID, _ := ctx.Value("userID").(string)
	var v validation.Errors
	v.Add(validation.Required(courseCode, "courseCode"))
	v.Add(validation.DigitsOnly(courseCode, "courseCode"))
	v.Add(validation.ExactLength(courseCode, "courseCode", 4))
	v.Add(validation.Required(courseName, "courseName"))
	v.Add(validation.Alphanumeric(courseName, "courseName", true))
	v.Add(validation.LengthRange(courseName, "courseName", 3, 50))
	v.Add(validation.Required(campus, "campus"))
	v.Add(validation.LettersOnly(campus, "campus"))
	v.Add(validation.LengthRange(campus, "campus", 3, 30))
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
	if course.OwnerID != "" && t.graph != nil {
		owner, err := t.graph.GetUser(ctx, course.OwnerID)
		if err == nil && owner != nil {
			course.CreatedByName = owner.DisplayName
		}
	}
	return course, nil
}

func (t *TimetableService) GetCourses(ctx context.Context, page, perPage int, search, sortBy, sortOrder string) (*pagination.Result[timetablemodels.Course], error) {
	fetch := func(ctx context.Context, p, pp int) ([]timetablemodels.Course, error) {
		return t.repo.GetCourses(ctx, p, pp, search, sortBy, sortOrder)
	}
	count := func(ctx context.Context) (int, error) {
		return t.repo.GetCoursesCount(ctx, search)
	}
	return pagination.Paginate(ctx, page, perPage, fetch, count)
}

// GetOwnCourses returns courses the authenticated user is enrolled in.
func (t *TimetableService) GetOwnCourses(ctx context.Context) ([]timetablemodels.Course, error) {
	userID, _ := ctx.Value("userID").(string)
	if userID == "" {
		return nil, nil
	}
	return t.repo.GetCoursesByUserId(ctx, userID)
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
		v.Add(validation.LengthRange(*courseName, "courseName", 3, 50))
	}
	if campus != nil {
		v.Add(validation.Required(*campus, "campus"))
		v.Add(validation.LettersOnly(*campus, "campus"))
		v.Add(validation.LengthRange(*campus, "campus", 3, 30))
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
