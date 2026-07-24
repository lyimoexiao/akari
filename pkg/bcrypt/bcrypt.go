package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the recommended cost factor for bcrypt hashing.
const DefaultCost = bcrypt.DefaultCost

// HashPassword generates a bcrypt hash from a plain-text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// HashPasswordWithCost generates a bcrypt hash with a custom cost factor.
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password with cost: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a plain-text password against a bcrypt hash.
// Returns true if they match, false otherwise.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidateHash checks if a bcrypt hash is valid (correct format and cost).
func ValidateHash(hash string) error {
	_, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return fmt.Errorf("invalid bcrypt hash: %w", err)
	}
	return nil
}
