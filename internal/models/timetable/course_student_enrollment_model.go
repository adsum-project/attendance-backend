package timetablemodels

type CourseStudentEnrollment struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	YearOfStudy int    `json:"yearOfStudy"`
}
