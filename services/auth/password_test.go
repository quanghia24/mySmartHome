package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Errorf("error hasing password %v", err)
	}
	if hash == "" {
		t.Errorf("expected has to be not empty")
	}
	if hash == "password" {
		t.Errorf("expected has to be different from password")
	}
}

func TestComparePasswords(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Errorf("error hasing password %v", err)
	}

	if !ComparePasswords(hash, []byte("password")) {
		t.Errorf("expect password to match hash")
	}

	if ComparePasswords(hash, []byte("notpassword")) {
		t.Errorf("expect password to not match hash")
	}

}
