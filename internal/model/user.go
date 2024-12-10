package model

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
	Email       string `json:"email"`
}

type RegisterReqeust struct {
	LoginRequest
	Username  string `json:"username" validate:"required,min=4,max=20"`
	EmailCode string `json:"email_code" validate:"required,len=6"`
}

type ForgetPasswordReqeust struct {
	LoginRequest
	EmailCode string `json:"email_code" validate:"required,len=6"`
}

type SendCodeRequest struct {
	Email string `validate:"required,email"`
}
