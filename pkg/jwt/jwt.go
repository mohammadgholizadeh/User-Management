package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string
	jwt.RegisteredClaims
}

type Manager struct {
	SecretKey string
}

func NewManager(SecretKey string) *Manager {
	return &Manager{SecretKey: SecretKey}
}

func (m *Manager) GenerateToken(UserID, role string, exiration time.Duration) (string, error) {
	claims := &Claims{
		UserID: UserID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   UserID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.SecretKey))
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
