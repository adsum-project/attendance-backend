package timetablemodels

type Module struct {
	ModuleID      string `db:"module_id" json:"moduleId"`
	ModuleCode    string `db:"module_code" json:"moduleCode"`
	ModuleName    string `db:"module_name" json:"moduleName"`
	OwnerID       string `db:"owner_id" json:"ownerId"`
	CreatedByName string `db:"-" json:"createdByName,omitempty"`
	StartDate     string `db:"start_date" json:"startDate"`
	EndDate       string `db:"end_date" json:"endDate"`
	CreatedAt     string `db:"created_at" json:"createdAt"`
	UpdatedAt     string `db:"updated_at" json:"updatedAt"`
}
