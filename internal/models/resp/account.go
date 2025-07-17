package resp

import "github.com/google/uuid"

type Account struct {
	ID       uuid.UUID `json:"id"`
	Avatar   string    `json:"avatar"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
}
