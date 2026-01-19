package mocks

import (
	"context"

	"github.com/cdriehuys/stuff2/internal/models"
)

type UserModel struct {
	RegisterError  error
	RegisteredUser models.NewUser

	VerifyEmailToken string
	VerifyEmailError error
}

func (m *UserModel) Register(_ context.Context, user models.NewUser) error {
	m.RegisteredUser = user

	return m.RegisterError
}

func (m *UserModel) VerifyEmail(_ context.Context, token string) error {
	m.VerifyEmailToken = token

	return m.VerifyEmailError
}
