package security

import "github.com/alexedwards/argon2id"

type Argon2IDHasher struct{}

func (h Argon2IDHasher) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (h Argon2IDHasher) ComparePasswordAndHash(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
