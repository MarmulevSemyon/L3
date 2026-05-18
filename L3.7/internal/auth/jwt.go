package auth

import (
	"errors"
	"fmt"
	"time"

	"warehousecontrol/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// Claims хранит данные пользователя внутри JWT.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`

	jwt.RegisteredClaims
}

// GenerateToken создаёт JWT для пользователя.
func GenerateToken(actor domain.Actor, secret string) (string, error) {
	claims := Claims{
		UserID:   actor.ID,
		Username: actor.Username,
		Role:     actor.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

// ParseToken проверяет JWT и возвращает пользователя из claims.
func ParseToken(tokenString string, secret string) (domain.Actor, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secret), nil
		},
	)
	if err != nil {
		return domain.Actor{}, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return domain.Actor{}, errors.New("invalid token")
	}

	if claims.UserID <= 0 || claims.Username == "" || claims.Role == "" {
		return domain.Actor{}, errors.New("invalid token claims")
	}

	return domain.Actor{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}
