package authutils

import "golang.org/x/crypto/bcrypt"

// HashPassword возвращает bcrypt-хэш пароля.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword сравнивает обычный пароль и сохранённый хэш.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
