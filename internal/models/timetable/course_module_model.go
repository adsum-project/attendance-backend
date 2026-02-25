package timetablemodels

type CourseModule struct {
	ModuleID    string `db:"module_id" json:"moduleId"`
	ModuleCode  string `db:"module_code" json:"moduleCode"`
	ModuleName  string `db:"module_name" json:"moduleName"`
	OwnerID     string `db:"owner_id" json:"ownerId"`
	StartDate   string `db:"start_date" json:"startDate"`
	EndDate     string `db:"end_date" json:"endDate"`
	YearOfStudy int    `db:"year_of_study" json:"yearOfStudy"`
}
