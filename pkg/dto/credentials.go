package dto

import "github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"

//easyjson:json
type UserCredentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type RegisterUserCredentials struct {
	Username string    `json:"login"`
	Password string    `json:"password"`
	Role     user.Role `json:"role"`
}
