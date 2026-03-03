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

func (t *TimetableService) CreateModule(ctx context.Context, moduleCode, moduleName, startDate, endDate string) (string, error) {
	ownerID, _ := ctx.Value("userID").(string)
	var v validation.Errors
	v.Add(validation.Required(moduleCode, "moduleCode"))
	v.Add(validation.ExactLength(moduleCode, "moduleCode", 6))
	v.Add(validation.ModuleCodeFormat(moduleCode, "moduleCode"))
	v.Add(validation.Required(moduleName, "moduleName"))
	v.Add(validation.Alphanumeric(moduleName, "moduleName", true))
	v.Add(validation.LengthRange(moduleName, "moduleName", 3, 50))
	v.Add(validation.Required(startDate, "startDate"))
	v.Add(validation.DateFormat(startDate, "startDate"))
	v.Add(validation.Required(endDate, "endDate"))
	v.Add(validation.DateFormat(endDate, "endDate"))
	v.Add(validation.DateRange(startDate, endDate, "endDate"))
	if err := v.Result(); err != nil {
		return "", err
	}

	exists, err := t.repo.ModuleCodeExists(ctx, moduleCode)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errs.Error(409, "moduleCode already exists")
	}

	return t.repo.CreateModule(ctx, moduleCode, moduleName, ownerID, startDate, endDate)
}

func (t *TimetableService) GetModule(ctx context.Context, moduleID string) (*timetablemodels.Module, error) {
	module, err := t.repo.GetModuleByID(ctx, moduleID)
	if err != nil {
		if errors.Is(err, timetablerepo.ErrModuleNotFound) {
			return nil, errs.NotFound("Module not found")
		}
		return nil, err
	}
	if module.OwnerID != "" && t.graph != nil {
		owner, err := t.graph.GetUser(ctx, module.OwnerID)
		if err == nil && owner != nil {
			module.CreatedByName = owner.DisplayName
		}
	}
	return module, nil
}

func (t *TimetableService) GetModules(ctx context.Context, page, perPage int, search, sortBy, sortOrder string) (*pagination.Result[timetablemodels.Module], error) {
	fetch := func(ctx context.Context, p, pp int) ([]timetablemodels.Module, error) {
		return t.repo.GetModules(ctx, p, pp, search, sortBy, sortOrder)
	}
	count := func(ctx context.Context) (int, error) {
		return t.repo.GetModulesCount(ctx, search)
	}
	return pagination.Paginate(ctx, page, perPage, fetch, count)
}

func (t *TimetableService) UpdateModule(ctx context.Context, moduleID string, moduleCode, moduleName, startDate, endDate *string) error {
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
	if moduleCode == nil && moduleName == nil && startDate == nil && endDate == nil {
		return errs.BadRequest("at least one field must be provided")
	}

	var v validation.Errors
	if moduleCode != nil {
		v.Add(validation.Required(*moduleCode, "moduleCode"))
		v.Add(validation.ExactLength(*moduleCode, "moduleCode", 6))
		v.Add(validation.ModuleCodeFormat(*moduleCode, "moduleCode"))
	}
	if moduleName != nil {
		v.Add(validation.Required(*moduleName, "moduleName"))
		v.Add(validation.Alphanumeric(*moduleName, "moduleName", true))
		v.Add(validation.LengthRange(*moduleName, "moduleName", 3, 50))
	}
	if startDate != nil {
		v.Add(validation.Required(*startDate, "startDate"))
		v.Add(validation.DateFormat(*startDate, "startDate"))
	}
	if endDate != nil {
		v.Add(validation.Required(*endDate, "endDate"))
		v.Add(validation.DateFormat(*endDate, "endDate"))
	}
	if startDate != nil || endDate != nil {
		startVal := validation.OptionalString(startDate, module.StartDate)
		endVal := validation.OptionalString(endDate, module.EndDate)
		v.Add(validation.DateRange(startVal, endVal, "endDate"))
	}
	if err := v.Result(); err != nil {
		return err
	}

	if moduleCode != nil && *moduleCode != module.ModuleCode {
		exists, err := t.repo.ModuleCodeExists(ctx, *moduleCode)
		if err != nil {
			return err
		}
		if exists {
			return errs.Error(409, "moduleCode already exists")
		}
	}

	return t.repo.UpdateModule(ctx, moduleID, moduleCode, moduleName, startDate, endDate)
}

func (t *TimetableService) DeleteModule(ctx context.Context, moduleID string) error {
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
	return t.repo.DeleteModule(ctx, moduleID)
}
