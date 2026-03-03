package usermodels

type User struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Role        string `json:"role,omitempty"` // "admin", "staff", or "default" (student)
}
