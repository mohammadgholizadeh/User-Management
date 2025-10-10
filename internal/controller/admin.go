package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/internal/domain"
)

type adminHandler struct {
	adminSVC domain.AdminService
	logger   *zap.Logger
}

func AdminRoutes(g *echo.Group, adminSVC domain.AdminService, logger *zap.Logger) {
	h := &adminHandler{adminSVC: adminSVC, logger: logger}
	g.POST("/admins", h.create)
	g.GET("/admins", h.get)
	g.DELETE("/admins", h.delete)
}

func (h *adminHandler) create(c echo.Context) error {
	var a domain.Admin
	if err := c.Bind(&a); err != nil {
		h.logger.Warn("admin.create: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := h.adminSVC.Create(c.Request().Context(), a); err != nil {
		h.logger.Warn("admin.create failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *adminHandler) get(c echo.Context) error {
	id := domain.NationalID(c.QueryParam("id"))
	a, err := h.adminSVC.Get(c.Request().Context(), id)
	if err != nil {
		h.logger.Warn("admin.get failed", zap.String("id", string(id)), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, a)
}

func (h *adminHandler) delete(c echo.Context) error {
	id := domain.NationalID(c.QueryParam("id"))
	if err := h.adminSVC.Delete(c.Request().Context(), id); err != nil {
		h.logger.Warn("admin.delete failed", zap.String("id", string(id)), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}
