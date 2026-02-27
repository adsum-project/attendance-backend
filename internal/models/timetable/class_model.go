package timetablemodels

type Class struct {
	ClassID    string  `db:"class_id" json:"classId"`
	ModuleID   string  `db:"module_id" json:"moduleId"`
	ClassName  string  `db:"class_name" json:"className"`
	Room       string  `db:"room" json:"room"`
	DayOfWeek  int     `db:"day_of_week" json:"dayOfWeek"`
	StartsAt   string  `db:"starts_at" json:"startsAt"`
	EndsAt     string  `db:"ends_at" json:"endsAt"`
	Recurrence string  `db:"recurrence" json:"recurrence"`
	CreatedAt  string  `db:"created_at" json:"createdAt"`
	UpdatedAt  string  `db:"updated_at" json:"updatedAt"`
}
