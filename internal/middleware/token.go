package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/pkg/jwt"
)

type TokenMiddleware struct {
	jwt *jwt.Manager
	lg  *zap.Logger
}

func NewTokenMiddleware(jwtManager *jwt.Manager, logger *zap.Logger) *TokenMiddleware {
	return &TokenMiddleware{jwt: jwtManager, lg: logger}
}

func (m *TokenMiddleware) Parse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, _ := m.extractClaims(c)
		if claims != nil {
			c.Set("claims", claims)
		}
		return next(c)
	}
}

func (m *TokenMiddleware) ParseAndVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := m.extractClaims(c)
		if err != nil {
			m.lg.Warn("invalid token", zap.Error(err))
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		if claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}
		c.Set("claims", claims)
		return next(c)
	}
}

func (m *TokenMiddleware) extractClaims(c echo.Context) (*jwt.Claims, error) {
	authz := c.Request().Header.Get("Authorization")
	if authz == "" {
		return nil, nil
	}
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
	}
	token := strings.TrimSpace(authz[7:])
	claims, err := m.jwt.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
