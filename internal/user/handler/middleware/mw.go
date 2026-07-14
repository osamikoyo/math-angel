package middleware

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/user/service"
)

type Middleware struct {
	service *service.Service
}

func renderWithStatus(c *echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}

func NewMiddleware(svc *service.Service) *Middleware {
	return &Middleware{
		service: svc,
	}
}
