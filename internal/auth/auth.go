package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	isTrue, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return isTrue, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(header http.Header) (string, error) {
	bearerString := header.Get("Authorization")
	if bearerString == "" {
		return "", fmt.Errorf("no authorization provided")
	}
	if !strings.HasPrefix(bearerString, "Bearer ") {
		return "", fmt.Errorf("incorrect type authorization header provided")
	}
	token := strings.TrimPrefix(bearerString, "Bearer ")
	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(key)
	return token, nil
}

func GetAPIKey(header http.Header) (string, error) {
	apiString := header.Get("Authorization")
	if apiString == "" {
		return "", fmt.Errorf("no authorization provided")
	}
	if !strings.HasPrefix(apiString, "ApiKey ") {
		return "", fmt.Errorf("incorrect type authorization header provided")
	}
	key := strings.TrimPrefix(apiString, "ApiKey ")
	return key, nil
}

