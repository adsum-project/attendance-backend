package timetablemodels

type Course struct {
	CourseID   string `db:"course_id" json:"courseId"`
	CourseCode string `db:"course_code" json:"courseCode"`
	CourseName string `db:"course_name" json:"courseName"`
	Campus     string `db:"campus" json:"campus"`
	OwnerID    string `db:"owner_id" json:"ownerId"`
	CreatedAt  string `db:"created_at" json:"createdAt"`
	UpdatedAt  string `db:"updated_at" json:"updatedAt"`
}
