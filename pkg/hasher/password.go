package hasher

import "golang.org/x/crypto/bcrypt"

type PasswordHash struct {
}

func NewPasswordHash() *PasswordHash {
	return &PasswordHash{}
}

func (h *PasswordHash) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (h *PasswordHash) ComparePassword(passsword, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passsword))
	return err == nil
}
