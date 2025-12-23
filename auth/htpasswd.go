package auth

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type HtpasswdAuth struct {
	path string
}

func NewHtpasswdAuth(path string) *HtpasswdAuth {
	return &HtpasswdAuth{path: path}
}

func (h *HtpasswdAuth) Authenticate(user, pass string) error {
	f, err := os.Open(h.path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == user {
			if bcrypt.CompareHashAndPassword([]byte(parts[1]), []byte(pass)) == nil {
				return nil
			}
			return ErrAuthFailed
		}
	}
	return ErrAuthFailed
}
