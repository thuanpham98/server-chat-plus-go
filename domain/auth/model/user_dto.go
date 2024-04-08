package domain_auth_model

type UserDTO struct{
	Id   string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}