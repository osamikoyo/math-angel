package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
)

type LoginRequest struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

func (h *Handler) Login(c *echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "failed parse request")
	}

	token, err := h.service.LoginUser(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound) {
			return c.String(http.StatusNotFound, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24 * 7,
	})

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"token": token,
		})
}
