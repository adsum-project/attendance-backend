package verificationmodels

type AttendanceRecord struct {
	RecordID   string  `db:"record_id" json:"recordId"`
	UserID     string  `db:"user_id" json:"userId,omitempty"`
	ClassID    string  `db:"class_id" json:"classId"`
	ClassName  string  `db:"class_name" json:"className"`
	ModuleCode string  `db:"module_code" json:"moduleCode"`
	ModuleName string  `db:"module_name" json:"moduleName"`
	Room       string  `db:"room" json:"room"`
	SignedInAt string  `db:"signed_in_at" json:"signedInAt"`
	Method     string  `db:"method" json:"method"`
	Status     string  `db:"status" json:"status"`
}
