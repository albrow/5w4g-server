package models

type AdminUser struct {
	Email          string `json:"email" zoom:"index"`
	HashedPassword string `json:"-" zoom:"index"`
	Identifier
}
