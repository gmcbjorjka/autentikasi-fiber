package utils

import "golang.org/x/crypto/bcrypt"

func Hash(plain string) (string, error) {
	if plain == "" {
		return "", nil // biar kosong tidak di-hash
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func Check(plain, hashed string) bool {
	if plain == "" || hashed == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
