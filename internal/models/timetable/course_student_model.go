package timetablemodels

type CourseStudent struct {
	CourseID    string `db:"course_id" json:"courseId"`
	UserID      string `db:"user_id" json:"userId"`
	YearOfStudy int    `db:"year_of_study" json:"yearOfStudy"`
}
