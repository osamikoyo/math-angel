package handler

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/user/handler/middleware"
	"github.com/osamikoyo/math-angel/internal/user/service"
)

type Handler struct{
	mw *middleware.Middleware
	service *service.Service
}

func NewHandler(mw *middleware.Middleware, service *service.Service) *Handler {
	return &Handler{
		mw: mw,
		service: service,
	}
}

func renderWithStatus(c *echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}

func (h *Handler) RegisterRouters(e *echo.Echo) {
	e.GET("/profile", h.GetProfile, h.mw.Auth)
}