package auth

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordAuth(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	auth := NewPasswordAuth(string(hash))

	if err := auth.Authenticate("", "secret"); err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if err := auth.Authenticate("", "wrong"); err == nil {
		t.Error("expected failure for wrong password")
	}
}

func TestHtpasswdAuth(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("mypass"), bcrypt.DefaultCost)
	content := "alice:" + string(hash) + "\n"

	f, err := os.CreateTemp("", "htpasswd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	auth := NewHtpasswdAuth(f.Name())

	if err := auth.Authenticate("alice", "mypass"); err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if err := auth.Authenticate("alice", "wrong"); err == nil {
		t.Error("expected failure for wrong password")
	}

	if err := auth.Authenticate("bob", "mypass"); err == nil {
		t.Error("expected failure for unknown user")
	}
}
