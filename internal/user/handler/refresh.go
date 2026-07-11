package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) RefreshToken(c *echo.Context) error {
	refreshCookie, err := c.Cookie("refresh_token")
	if err != nil || refreshCookie.Value == "" {
		return c.String(http.StatusUnauthorized, "refresh token not found")
	}

	newAccessToken, err := h.service.Refresh(c.Request().Context(), refreshCookie.Value)
	if err != nil {
		return c.String(http.StatusUnauthorized, "invalid refresh token")
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24,
	})

	return c.String(http.StatusOK, "token refreshed")
}
