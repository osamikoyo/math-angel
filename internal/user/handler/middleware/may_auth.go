package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

func (mw *Middleware) MayAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString.Value == "" {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				tokenString = &http.Cookie{Value: strings.TrimPrefix(authHeader, "Bearer ")}
			} else {
				return next(c)
			}
		}

		ctx := c.Request().Context()

		userID, err := mw.service.Validate(ctx, tokenString.Value)
		if err == nil {
			c.Set("user_id", userID)
		}

		return next(c)
	}
}
