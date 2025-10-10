package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

type userHandler struct {
	userSVC    domain.UserService
	logger *zap.Logger
}

func UserRoutes(g *echo.Group, userSVC domain.UserService, logger *zap.Logger) {
	h := &userHandler{userSVC: userSVC, logger: logger}
	g.POST("/users", h.create)
	g.GET("/users", h.list)
	g.GET("/users/get", h.getByNationalID)
	g.GET("/users/by-username", h.getByUsername)
	g.GET("/users/by-mobile", h.getByMobile)
	g.GET("/users/by-role", h.getByRole)
	g.GET("/users/role", h.getRole)
	g.PATCH("/users", h.update)
	g.DELETE("/users", h.delete)
}

func (h *userHandler) create(c echo.Context) error {
	type createReq struct {
		NationalID   domain.NationalID   `json:"national_id"`
		FirstName    string              `json:"first_name"`
		LastName     string              `json:"last_name"`
		Username     domain.Username     `json:"username"`
		Password     domain.Password     `json:"password"`
		MobileNumber domain.MobileNumber `json:"mobile_number"`
		Gender       string              `json:"gender"`
		Email        domain.Email        `json:"email"`
		Role         string              `json:"role"`
	}
	var req createReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("user.create: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	u := domain.User{
		NationalID:   req.NationalID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserName:     req.Username,
		Password:     req.Password,
		MobileNumber: req.MobileNumber,
		Gender:       req.Gender,
		Email:        req.Email,
		Role:         req.Role,
		IsActive:     true,
	}
	if err := h.userSVC.Create(c.Request().Context(), u); err != nil {
		h.logger.Warn("user.create failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *userHandler) list(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	users, err := h.userSVC.GetAllUsers(c.Request().Context(), limit, offset)
	if err != nil {
		h.logger.Error("user.list failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *userHandler) getByNationalID(c echo.Context) error {
	id := domain.NationalID(c.QueryParam("id"))
	u, err := h.userSVC.GetUserByNationalID(c.Request().Context(), id)
	if err != nil {
		h.logger.Warn("user.get_by_id failed", zap.String("id", string(id)), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, u)
}

func (h *userHandler) getByUsername(c echo.Context) error {
	username := domain.Username(c.QueryParam("username"))
	u, err := h.userSVC.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		h.logger.Warn("user.get_by_username failed", zap.String("username", string(username)), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, u)
}

func (h *userHandler) getByMobile(c echo.Context) error {
	mobile := domain.MobileNumber(c.QueryParam("mobile"))
	u, err := h.userSVC.GetUserByMobileNumber(c.Request().Context(), mobile)
	if err != nil {
		h.logger.Warn("user.get_by_mobile failed", zap.String("mobile", string(mobile)), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, u)
}

func (h *userHandler) getByRole(c echo.Context) error {
	role := c.QueryParam("role")
	users, err := h.userSVC.GetUsersByRole(c.Request().Context(), role)
	if err != nil {
		h.logger.Warn("user.get_by_role failed", zap.String("role", role), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *userHandler) getRole(c echo.Context) error {
	id := domain.NationalID(c.QueryParam("id"))
	u, err := h.userSVC.GetUserRole(c.Request().Context(), id)
	if err != nil {
		h.logger.Warn("user.get_role failed", zap.String("id", string(id)), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"role": u.Role})
}

func (h *userHandler) update(c echo.Context) error {
	var req struct {
		ID           domain.NationalID   `json:"id"`
		FirstName    string              `json:"first_name"`
		LastName     string              `json:"last_name"`
		UserName     domain.Username     `json:"username"`
		MobileNumber domain.MobileNumber `json:"mobile_number"`
		Gender       string              `json:"gender"`
		Email        domain.Email        `json:"email"`
		Role         string              `json:"role"`
		Password     domain.Password     `json:"password"`
		IsActive     *bool               `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("user.update: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	u := domain.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserName:     req.UserName,
		MobileNumber: req.MobileNumber,
		Gender:       req.Gender,
		Email:        req.Email,
		Role:         req.Role,
		Password:     req.Password,
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	if err := h.userSVC.Update(c.Request().Context(), u, req.ID); err != nil {
		h.logger.Warn("user.update failed", zap.String("id", string(req.ID)), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *userHandler) delete(c echo.Context) error {
	id := domain.NationalID(c.QueryParam("id"))
	if err := h.userSVC.Delete(c.Request().Context(), id); err != nil {
		h.logger.Warn("user.delete failed", zap.String("id", string(id)), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
