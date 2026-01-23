package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	if hash == password {
		t.Errorf("Hash should not be the same as the password")
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	if !match {
		t.Errorf("Expected password to match hash, but it didn't")
	}

	match, err = CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("Error checking wrong password: %v", err)
	}
	if match {
		t.Errorf("Expected wrong password NOT to match, but it did")
	}
}

func TestJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret-key-12345"
	duration := time.Hour

	tokenString, err := MakeJWT(userID, secret, duration)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}

	if tokenString == "" {
		t.Error("Expected token string to be non-empty")
	}

	parsedID, err := ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}

	if parsedID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, parsedID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiredDuration := -time.Second 

	tokenString, err := MakeJWT(userID, secret, expiredDuration)
	if err != nil {
		t.Fatalf("Error making expired JWT: %v", err)
	}

	_, err = ValidateJWT(tokenString, secret)
	if err == nil {
		t.Error("Expected error for expired token, but got nil")
	}
}

func TestWrongSecretJWT(t *testing.T) {
	userID := uuid.New()
	tokenString, err := MakeJWT(userID, "secret-A", time.Hour)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}

	_, err = ValidateJWT(tokenString, "secret-B")
	if err == nil {
		t.Error("Expected error when validating with wrong secret, got nil")
	}
}