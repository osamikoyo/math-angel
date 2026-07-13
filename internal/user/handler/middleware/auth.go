package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (mw *Middleware) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var tokenValue string

		cookie, err := c.Cookie("access_token")
		if err == nil && cookie != nil && cookie.Value != "" {
			tokenValue = cookie.Value
			log.Print("token loaded from cookie: " + tokenValue)
		} else {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				tokenValue = strings.TrimPrefix(authHeader, "Bearer ")
				tokenValue = strings.TrimSpace(tokenValue)
				log.Print("token loaded from header: " + tokenValue)
			}
		}

		if tokenValue == "" {
			return next(c)
		}

		ctx := c.Request().Context()
		userID, err := mw.service.Validate(ctx, tokenValue)
		if err != nil {
			log.Printf("validate error: %s", err.Error())

			return renderWithStatus(c, http.StatusUnauthorized, pages.Register())
		}

		c.Set("user_id", userID)
		return next(c)
	}
}
