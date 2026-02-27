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

func (t *TimetableService) GetClasses(ctx context.Context, moduleID string) ([]timetablemodels.Class, error) {
	_, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return nil, errs.NotFound("Module not found")
		}
		return nil, err
	}
	return t.repo.GetClasses(ctx, moduleID)
}

func (t *TimetableService) GetClass(ctx context.Context, moduleID, classID string) (*timetablemodels.Class, error) {
	_, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return nil, errs.NotFound("Module not found")
		}
		return nil, err
	}
	class, err := t.repo.GetClassByID(ctx, moduleID, classID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrClassNotFound) {
			return nil, errs.NotFound("Class not found")
		}
		return nil, err
	}
	return class, nil
}

func (t *TimetableService) CreateClass(ctx context.Context, moduleID, className, room string, dayOfWeek int, startsAt, endsAt, recurrence string) (string, error) {
	module, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return "", errs.NotFound("Module not found")
		}
		return "", err
	}
	if !authorization.IsOwnerOrAdmin(ctx, module.OwnerID) {
		return "", errs.Forbidden("")
	}

	var v validation.Errors
	v.Add(validation.Required(className, "className"))
	v.Add(validation.LengthRange(className, "className", 1, 100))
	v.Add(validation.Required(room, "room"))
	v.Add(validation.LengthRange(room, "room", 1, 100))
	v.Add(validation.IntRange(dayOfWeek, "dayOfWeek", 1, 7))
	v.Add(validation.Required(startsAt, "startsAt"))
	v.Add(validation.TimeFormat(startsAt, "startsAt"))
	v.Add(validation.Required(endsAt, "endsAt"))
	v.Add(validation.TimeFormat(endsAt, "endsAt"))
	v.Add(validation.TimeRange(startsAt, endsAt, "endsAt"))
	v.Add(validation.Required(recurrence, "recurrence"))
	v.Add(validation.Recurrence(recurrence, "recurrence"))
	if err := v.Result(); err != nil {
		return "", err
	}

	conflict, err := t.repo.HasRoomConflict(ctx, room, dayOfWeek, startsAt, endsAt, "")
	if err != nil {
		return "", err
	}
	if conflict {
		return "", errs.Conflict("room is already booked at this time")
	}

	return t.repo.CreateClass(ctx, moduleID, className, room, dayOfWeek, startsAt, endsAt, recurrence)
}

func (t *TimetableService) UpdateClass(ctx context.Context, moduleID, classID string, className, room *string, dayOfWeek *int, startsAt, endsAt, recurrence *string) error {
	module, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return errs.NotFound("Module not found")
		}
		return err
	}
	if !authorization.IsOwnerOrAdmin(ctx, module.OwnerID) {
		return errs.Forbidden("")
	}

	class, err := t.repo.GetClassByID(ctx, moduleID, classID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrClassNotFound) {
			return errs.NotFound("Class not found")
		}
		return err
	}

	if className == nil && room == nil && dayOfWeek == nil && startsAt == nil && endsAt == nil && recurrence == nil {
		return errs.BadRequest("at least one field must be provided")
	}

	var v validation.Errors
	if className != nil {
		v.Add(validation.Required(*className, "className"))
		v.Add(validation.LengthRange(*className, "className", 1, 100))
	}
	if room != nil {
		v.Add(validation.Required(*room, "room"))
		v.Add(validation.LengthRange(*room, "room", 1, 100))
	}
	if dayOfWeek != nil {
		v.Add(validation.IntRange(*dayOfWeek, "dayOfWeek", 1, 7))
	}
	if startsAt != nil {
		v.Add(validation.Required(*startsAt, "startsAt"))
		v.Add(validation.TimeFormat(*startsAt, "startsAt"))
	}
	if endsAt != nil {
		v.Add(validation.Required(*endsAt, "endsAt"))
		v.Add(validation.TimeFormat(*endsAt, "endsAt"))
	}
	if startsAt != nil || endsAt != nil {
		startVal := validation.OptionalString(startsAt, class.StartsAt)
		endVal := validation.OptionalString(endsAt, class.EndsAt)
		v.Add(validation.TimeRange(startVal, endVal, "endsAt"))
	}
	if recurrence != nil {
		v.Add(validation.Required(*recurrence, "recurrence"))
		v.Add(validation.Recurrence(*recurrence, "recurrence"))
	}
	if err := v.Result(); err != nil {
		return err
	}

	effectiveRoom := validation.OptionalString(room, class.Room)
	effectiveDayOfWeek := class.DayOfWeek
	if dayOfWeek != nil {
		effectiveDayOfWeek = *dayOfWeek
	}
	effectiveStartsAt := validation.OptionalString(startsAt, class.StartsAt)
	effectiveEndsAt := validation.OptionalString(endsAt, class.EndsAt)

	conflict, err := t.repo.HasRoomConflict(ctx, effectiveRoom, effectiveDayOfWeek, effectiveStartsAt, effectiveEndsAt, classID)
	if err != nil {
		return err
	}
	if conflict {
		return errs.Conflict("room is already booked at this time")
	}

	return t.repo.UpdateClass(ctx, moduleID, classID, className, room, dayOfWeek, startsAt, endsAt, recurrence)
}

func (t *TimetableService) DeleteClass(ctx context.Context, moduleID, classID string) error {
	module, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return errs.NotFound("Module not found")
		}
		return err
	}
	if !authorization.IsOwnerOrAdmin(ctx, module.OwnerID) {
		return errs.Forbidden("")
	}

	err = t.repo.DeleteClass(ctx, moduleID, classID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrClassNotFound) {
			return errs.NotFound("Class not found")
		}
		return err
	}
	return nil
}
