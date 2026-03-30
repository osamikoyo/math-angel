package handler

import (
	"github.com/labstack/echo/v5"

	"github.com/a-h/templ"
	"github.com/osamikoyo/math-angel/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRouters(e *echo.Echo) {
	e.GET("/healthcheck", h.HealthCheck)
	
	e.GET("/", h.Home)
	e.GET("/train", h.StartTrain)

	e.Static("/static", "static")

	taskGroup := e.Group("/task")

	taskGroup.PUT("/inc/like", h.IncLike)
	taskGroup.PUT("/dec/like", h.DecLike)
	taskGroup.PUT("/inc/dislike", h.IncLike)
	taskGroup.PUT("/dec/dislike", h.DecDislike)

	taskGroup.GET("/get/bests/:type/level/:level", h.GetBests)
	taskGroup.GET("/get/task/:id", h.GetTask)
	taskGroup.GET("/get/random/:type/level/:level", h.GetRandomTask)
}

func renderWithStatus(c *echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}
