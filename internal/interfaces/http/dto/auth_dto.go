package dto

import "errors"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm" binding:"required,min=6"`
}

func (r RegisterRequest) Validate() error {
	if r.Password != r.Confirm {
		return errors.New("confirm must match password")
	}
	return nil
}
