package security

import "crypto/rand"

type TokenGenerator struct{}

func (g TokenGenerator) Generate() string {
	return rand.Text()
}
