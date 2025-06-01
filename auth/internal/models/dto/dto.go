package dto

type UserRegistrationCredentials struct {
	Email    string `json:"email" binding:"required" binding:"email"`
	Password string `json:"password" binding:"required" binding:"min=8"`
	Username string `json:"username" binding:"required" binding:"min=5"`
}

type UserLoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
