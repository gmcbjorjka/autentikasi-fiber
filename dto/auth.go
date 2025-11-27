package dto

type RegisterRequest struct {
	Nama     string `json:"nama" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Pin      string `json:"pin,omitempty" validate:"required,min=4,max=6"`
	ImgURL   string `json:"img_url,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	OTP      string `json:"otp" validate:"required,len=6"`
	Password string `json:"password" validate:"required,min=6"`
}
