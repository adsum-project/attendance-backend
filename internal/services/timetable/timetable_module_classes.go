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

func (t *TimetableService) CreateClass(ctx context.Context, moduleID, className, room, startsAt, endsAt, recurrence string, untilDate *string) (string, error) {
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
	v.Add(validation.Required(startsAt, "startsAt"))
	v.Add(validation.DateTimeFormat(startsAt, "startsAt"))
	v.Add(validation.Required(endsAt, "endsAt"))
	v.Add(validation.DateTimeFormat(endsAt, "endsAt"))
	v.Add(validation.DateTimeRange(startsAt, endsAt, "endsAt"))
	v.Add(validation.Required(recurrence, "recurrence"))
	v.Add(validation.Recurrence(recurrence, "recurrence"))
	if untilDate != nil && *untilDate != "" {
		v.Add(validation.DateFormat(*untilDate, "untilDate"))
	}
	if err := v.Result(); err != nil {
		return "", err
	}

	return t.repo.CreateClass(ctx, moduleID, className, room, startsAt, endsAt, recurrence, untilDate)
}

func (t *TimetableService) UpdateClass(ctx context.Context, moduleID, classID string, className, room, startsAt, endsAt, recurrence, untilDate *string) error {
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

	if className == nil && room == nil && startsAt == nil && endsAt == nil && recurrence == nil && untilDate == nil {
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
	if startsAt != nil {
		v.Add(validation.Required(*startsAt, "startsAt"))
		v.Add(validation.DateTimeFormat(*startsAt, "startsAt"))
	}
	if endsAt != nil {
		v.Add(validation.Required(*endsAt, "endsAt"))
		v.Add(validation.DateTimeFormat(*endsAt, "endsAt"))
	}
	if startsAt != nil || endsAt != nil {
		startVal := validation.OptionalString(startsAt, class.StartsAt)
		endVal := validation.OptionalString(endsAt, class.EndsAt)
		v.Add(validation.DateTimeRange(startVal, endVal, "endsAt"))
	}
	if recurrence != nil {
		v.Add(validation.Required(*recurrence, "recurrence"))
		v.Add(validation.Recurrence(*recurrence, "recurrence"))
	}
	if untilDate != nil && *untilDate != "" {
		v.Add(validation.DateFormat(*untilDate, "untilDate"))
	}
	if err := v.Result(); err != nil {
		return err
	}

	return t.repo.UpdateClass(ctx, moduleID, classID, className, room, startsAt, endsAt, recurrence, untilDate)
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
