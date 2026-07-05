package pagehandlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/user/service"
)

type PageHandler struct {
	service *service.Service
}

func NewPageHandler(service *service.Service) *PageHandler {
	return &PageHandler{
		service: service,
	}
}

func renderWithStatus(c *echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}
