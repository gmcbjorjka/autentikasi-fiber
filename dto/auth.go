package dto

type RegisterRequest struct {
	Nama     string `json:"nama" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Pin      string `json:"pin" validate:"required,min=4,max=6"`
	ImgURL   string `json:"img_url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
