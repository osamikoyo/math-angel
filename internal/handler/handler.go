package handler

import (
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/service"
)

type Handler struct {
	service *service.Service
}

func (h *Handler) NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRouters(e *echo.Echo) {
	e.GET("/healthcheck", h.HealthCheck)

	taskGroup := e.Group("/task")

	taskGroup.PUT("/inc/like", h.IncLike)
	taskGroup.PUT("/dec/like", h.DecLike)
	taskGroup.PUT("/inc/dislike", h.IncLike)
	taskGroup.PUT("/dec/dislike", h.DecDislike)

	taskGroup.GET("/get/random/:type/level/:level", h.GetRandomTask)
}
