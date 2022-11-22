package utils

import (
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"golang.org/x/crypto/bcrypt"
)

func ValidateLogPass(creds sharedTypes.Credentials) (err error) {
	if creds.Login == "" || len(creds.Login) < 5 {
		return ErrBadCredentials
	}
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
