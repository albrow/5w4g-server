package models

type AdminUser struct {
	Email          string `json:"email"`
	HashedPassword string `json:"-"`
	Identifier
}
