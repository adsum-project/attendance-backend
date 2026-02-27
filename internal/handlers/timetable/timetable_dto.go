package timetablehandlers

type CreateCourseRequest struct {
	CourseCode string `json:"courseCode"`
	CourseName string `json:"courseName"`
	Campus     string `json:"campus"`
}

type UpdateCourseRequest struct {
	CourseCode *string `json:"courseCode"`
	CourseName *string `json:"courseName"`
	Campus     *string `json:"campus"`
}

type CreateModuleRequest struct {
	ModuleCode string `json:"moduleCode"`
	ModuleName string `json:"moduleName"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

type UpdateModuleRequest struct {
	ModuleCode *string `json:"moduleCode"`
	ModuleName *string `json:"moduleName"`
	StartDate  *string `json:"startDate"`
	EndDate    *string `json:"endDate"`
}

type AssignModuleToCourseRequest struct {
	YearOfStudy int `json:"yearOfStudy"`
}

type CreateClassRequest struct {
	ClassName  string  `json:"className"`
	Room       string  `json:"room"`
	DayOfWeek  int     `json:"dayOfWeek"`
	StartsAt   string  `json:"startsAt"`
	EndsAt     string  `json:"endsAt"`
	Recurrence string  `json:"recurrence"`
}

type UpdateClassRequest struct {
	ClassName  *string `json:"className"`
	Room       *string `json:"room"`
	DayOfWeek  *int    `json:"dayOfWeek"`
	StartsAt   *string `json:"startsAt"`
	EndsAt     *string `json:"endsAt"`
	Recurrence *string `json:"recurrence"`
}

type NodeRoomAssignRequest struct {
	Room string `json:"room"`
}
