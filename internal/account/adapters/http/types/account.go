package types

type AccountSignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AccountLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AccountUpdateRequest struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type AccountResponse struct {
	Id       uint   `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type AccountCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}
