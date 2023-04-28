package hashPass

import "golang.org/x/crypto/bcrypt"

func CipherPassword(password string) (string, error) {
	ps, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(ps), nil
}
