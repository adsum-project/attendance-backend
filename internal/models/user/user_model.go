package usermodels

type User struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}
