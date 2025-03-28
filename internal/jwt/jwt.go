package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateJWT(userId string, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"id":  userId,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateJWT(tokenStr string, jwtSecret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	// extract user from jwt fields
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	if err := claims.Valid(); err != nil {
		return "", err
	}

	userID, ok := claims["id"].(string)
	if !ok || userID == "" {
		return "", err
	}

	return userID, nil
}
