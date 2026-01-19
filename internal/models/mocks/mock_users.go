package mocks

import (
	"context"

	"github.com/cdriehuys/stuff2/internal/models"
)

type UserModel struct {
	AuthenticatedEmail    string
	AuthenticatedPassword string
	AuthenticateUser      models.User
	AuthenticateError     error

	RegisterError  error
	RegisteredUser models.NewUser

	VerifyEmailToken string
	VerifyEmailError error
}

func (m *UserModel) Authenticate(_ context.Context, email string, password string) (models.User, error) {
	m.AuthenticatedEmail = email
	m.AuthenticatedPassword = password

	return m.AuthenticateUser, m.AuthenticateError
}

func (m *UserModel) Register(_ context.Context, user models.NewUser) error {
	m.RegisteredUser = user

	return m.RegisterError
}

func (m *UserModel) VerifyEmail(_ context.Context, token string) error {
	m.VerifyEmailToken = token

	return m.VerifyEmailError
}
