package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// func GenerateToken(length int) string {
// 	bytes := make([]byte, length)
// 	if _, err := rand.Read(bytes); err != nil {
// 		log.Fatalf("Failed to generate token: %s", err)
// 	}
// 	return base64.URLEncoding.EncodeToString(bytes)
// }

type CustomClaims struct {
	TenantID string `json:"tenantID"`
	UserType string `json:"userType,omitempty"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID uuid.UUID, tenantID uuid.UUID, jwtSecret string) (string, error) {
	claims := CustomClaims{
		TenantID: tenantID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ccms-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateToken(tokenString string, jwtSecret string) (uuid.UUID, uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Parse userID from Subject
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, uuid.Nil, errors.New("invalid user ID in token")
		}

		// Parse tenantID from custom claim
		tenantID, err := uuid.Parse(claims.TenantID)
		if err != nil {
			return uuid.Nil, uuid.Nil, errors.New("invalid tenant ID in token")
		}

		return userID, tenantID, nil
	}

	return uuid.Nil, uuid.Nil, errors.New("invalid token")
}
