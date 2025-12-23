package auth

import "golang.org/x/crypto/bcrypt"

type PasswordAuth struct {
	hash []byte
}

func NewPasswordAuth(hash string) *PasswordAuth {
	return &PasswordAuth{hash: []byte(hash)}
}

func (p *PasswordAuth) Authenticate(_, pass string) error {
	if err := bcrypt.CompareHashAndPassword(p.hash, []byte(pass)); err != nil {
		return ErrAuthFailed
	}
	return nil
}
