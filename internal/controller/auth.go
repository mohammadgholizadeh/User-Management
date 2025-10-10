package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
	"github.com/mohammadgholizadeh/user-management/internal/service"
)

type authHandler struct {
	authSVC    domain.AuthenticationService
	logger *zap.Logger
}

func AuthenticationRoutes(g *echo.Group, authSVC domain.AuthenticationService, logger *zap.Logger) {
	AuthenticationPublicRoutes(g, authSVC, logger)
	AuthenticationProtectedRoutes(g, authSVC, logger)
}

func AuthenticationPublicRoutes(g *echo.Group, authSVC domain.AuthenticationService, logger *zap.Logger) {
	h := &authHandler{authSVC: authSVC, logger: logger}
	g.POST("/auth/login", h.login)
	g.POST("/auth/register", h.register)
	g.POST("/auth/send-reset-password-token", h.sendResetPasswordToken)
	g.POST("/auth/reset-password", h.resetPassword)
}

func AuthenticationProtectedRoutes(g *echo.Group, authSVC domain.AuthenticationService, logger *zap.Logger) {
	h := &authHandler{authSVC: authSVC, logger: logger}
	g.POST("/auth/logout", h.logout)
}

func (h *authHandler) login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("login: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	resp, err := h.authSVC.Login(c.Request().Context(), req)
	if err != nil {
		h.logger.Warn("login failed", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *authHandler) register(c echo.Context) error {
	var req domain.RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("register: invalid request body", zap.Error(err), zap.Any("request_body", c.Request().Body))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid request",
			"details": err.Error(),
		})
	}

	// Log the parsed request for debugging
	h.logger.Debug("register request",
		zap.String("username", string(req.Username)),
		zap.String("email", string(req.Email)),
		zap.String("national_id", string(req.NationalID)),
	)

	resp, err := h.authSVC.Register(c.Request().Context(), req)
	if err != nil {
		h.logger.Warn("register failed",
			zap.Error(err),
			zap.String("username", string(req.Username)),
		)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *authHandler) logout(c echo.Context) error {
	authz := c.Request().Header.Get("Authorization")
	token := ""
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		token = strings.TrimSpace(authz[7:])
	}
	ctx := service.PutTokenInContext(c.Request().Context(), token)
	if err := h.authSVC.Logout(ctx); err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *authHandler) sendResetPasswordToken(c echo.Context) error {
	var req struct {
		Username   domain.Username   `json:"username"`
		NationalID domain.NationalID `json:"national_id"`
	}
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("send reset token: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Username == "" && req.NationalID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username or national_id required"})
	}
	if err := h.authSVC.SendResetPasswordToken(c.Request().Context(), req.Username, req.NationalID); err != nil {
		h.logger.Warn("send reset token failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *authHandler) resetPassword(c echo.Context) error {
	var req struct {
		Username    domain.Username   `json:"username"`
		NationalID  domain.NationalID `json:"national_id"`
		Token       string            `json:"token"`
		NewPassword string            `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("reset password: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Token == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
	}
	if err := h.authSVC.ResetPassword(c.Request().Context(), req.Username, req.NationalID, req.Token, req.NewPassword); err != nil {
		h.logger.Warn("reset password failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
