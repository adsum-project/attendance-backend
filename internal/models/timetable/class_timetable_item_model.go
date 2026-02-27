package timetablemodels

type ClassTimetableItem struct {
	ClassID         string `db:"class_id" json:"classId"`
	ModuleCode      string `db:"module_code" json:"moduleCode"`
	ModuleName      string `db:"module_name" json:"moduleName"`
	ModuleStartDate string `db:"module_start_date" json:"moduleStartDate"`
	ModuleEndDate   string `db:"module_end_date" json:"moduleEndDate"`
	CourseCode      string `db:"course_code" json:"courseCode"`
	CourseName      string `db:"course_name" json:"courseName"`
	ClassName       string `db:"class_name" json:"className"`
	Room            string `db:"room" json:"room"`
	DayOfWeek       int    `db:"day_of_week" json:"dayOfWeek"`
	StartsAt        string `db:"starts_at" json:"startsAt"`
	EndsAt          string `db:"ends_at" json:"endsAt"`
	Recurrence      string `db:"recurrence" json:"recurrence"`
}
