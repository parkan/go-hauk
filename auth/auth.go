package auth

import "errors"

var ErrAuthFailed = errors.New("authentication failed")

type Authenticator interface {
	Authenticate(user, pass string) error
}
