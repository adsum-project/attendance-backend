package timetablemodels

type ModuleCourse struct {
	CourseID    string `db:"course_id" json:"courseId"`
	CourseCode  string `db:"course_code" json:"courseCode"`
	CourseName  string `db:"course_name" json:"courseName"`
	Campus      string `db:"campus" json:"campus"`
	YearOfStudy int    `db:"year_of_study" json:"yearOfStudy"`
}
