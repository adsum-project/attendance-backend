package timetablerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/query"
)

const modulesTable = "modules"

var ErrModuleNotFound = errors.New("module not found")

func (r *TimetableRepository) GetModules(ctx context.Context, page, perPage int, search, sortBy, sortOrder string) ([]timetablemodels.Module, error) {
	var modules []timetablemodels.Module
	offset := (page - 1) * perPage
	where := ""
	args := []any{offset, perPage}
	if search != "" {
		where = " WHERE (module_code LIKE '%' + @p3 + '%' OR module_name LIKE '%' + @p3 + '%')"
		args = append(args, search)
	}
	order := query.OrderBy(sortBy, sortOrder, "ORDER BY created_at", map[string]string{
		"moduleCode": "module_code",
		"moduleName": "module_name",
		"startDate":  "start_date",
		"endDate":    "end_date",
	})
	err := r.db.SelectContext(
		ctx,
		&modules,
		`SELECT `+query.Guid("module_id")+` as module_id, module_code, module_name, `+query.Guid("owner_id")+` as owner_id,
		`+query.Date("start_date")+` as start_date, `+query.Date("end_date")+` as end_date,
		created_at, updated_at
		FROM `+modulesTable+`
		`+where+`
		`+order+`
		OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}
	return modules, nil
}

func (r *TimetableRepository) GetModuleByID(ctx context.Context, moduleID string) (*timetablemodels.Module, error) {
	var module timetablemodels.Module
	err := r.db.GetContext(
		ctx,
		&module,
		`SELECT `+query.Guid("module_id")+` as module_id, module_code, module_name, `+query.Guid("owner_id")+` as owner_id,
		`+query.Date("start_date")+` as start_date, `+query.Date("end_date")+` as end_date,
		created_at, updated_at
		FROM `+modulesTable+`
		WHERE `+query.GuidWhere("module_id", "@p1")+``,
		moduleID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrModuleNotFound
		}
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	return &module, nil
}

func (r *TimetableRepository) GetModulesCount(ctx context.Context, search string) (int, error) {
	var total int
	q := `SELECT COUNT(*) FROM ` + modulesTable
	args := []any{}
	if search != "" {
		q += ` WHERE (module_code LIKE '%' + @p1 + '%' OR module_name LIKE '%' + @p1 + '%')`
		args = append(args, search)
	}
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count modules: %w", err)
	}
	return total, nil
}

func (r *TimetableRepository) ModuleCodeExists(ctx context.Context, moduleCode string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT 1 FROM `+modulesTable+` WHERE module_code = @p1`,
		moduleCode,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check module code: %w", err)
	}
	return true, nil
}

func (r *TimetableRepository) CreateModule(ctx context.Context, moduleCode, moduleName, ownerID, startDate, endDate string) (string, error) {
	var moduleID string
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO `+modulesTable+` (module_id, module_code, module_name, owner_id, start_date, end_date, created_at, updated_at)
		OUTPUT `+query.Guid("INSERTED.module_id")+`
		VALUES (NEWID(), @p1, @p2, @p3, CAST(@p4 AS DATE), CAST(@p5 AS DATE), SYSUTCDATETIME(), SYSUTCDATETIME())`,
		moduleCode,
		moduleName,
		ownerID,
		startDate,
		endDate,
	).Scan(&moduleID)
	if err != nil {
		return "", fmt.Errorf("failed to create module: %w", err)
	}
	return moduleID, nil
}

func (r *TimetableRepository) UpdateModule(ctx context.Context, moduleID string, moduleCode, moduleName, startDate, endDate *string) error {
	clause, args, nextParam := query.UpdateAndCast(map[string]any{
		"module_code": moduleCode,
		"module_name": moduleName,
		"start_date":  startDate,
		"end_date":    endDate,
	}, map[string]string{"start_date": "DATE", "end_date": "DATE"})
	result, err := r.db.ExecContext(ctx,
		`UPDATE `+modulesTable+` SET `+clause+`, updated_at = SYSUTCDATETIME() WHERE `+query.GuidWhere("module_id", nextParam),
		append(args, moduleID)...,
	)
	if err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrModuleNotFound
	}
	return nil
}

func (r *TimetableRepository) DeleteModule(ctx context.Context, moduleID string) error {
	q := `DELETE FROM ` + modulesTable + ` WHERE ` + query.GuidWhere("module_id", "@p1")
	result, err := r.db.ExecContext(ctx, q, moduleID)
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrModuleNotFound
	}
	return nil
}
