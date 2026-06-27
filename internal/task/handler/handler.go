package handler

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v5"

	"github.com/osamikoyo/math-angel/internal/task/handler/pagehandlers"
	"github.com/osamikoyo/math-angel/internal/task/service"
	selfmw "github.com/osamikoyo/math-angel/internal/user/handler/middleware"
)

type Handler struct {
	service *service.TaskService
	pages   *pagehandlers.PageHandler
	mw *selfmw.Middleware
}

func NewHandler(service *service.TaskService, mw *selfmw.Middleware) *Handler {
	return &Handler{
		service: service,
		pages:   pagehandlers.NewPageHandler(service),
		mw: mw,
	}
}

func renderWithStatus(c *echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response())
}

func (h *Handler) RegisterRouters(e *echo.Echo) {
	e.GET("/healthcheck", h.HealthCheck)

	taskGroup := e.Group("/task")

	// page handlers

	e.GET("/", h.pages.Home)
	e.GET("/train", h.pages.StartTrain)
	e.GET("/bests", h.pages.GetInvitationForBests)
	e.GET("/add", h.pages.AddTaskPage)
	e.GET("/search", h.pages.SearchPage)

	e.Static("/static", "static")

	taskGroup.GET("/search/page_index/:page_index/page_size/:page_size", h.pages.Search)

	taskGroup.GET("/get/:id", h.pages.GetTask, h.mw.MayAuth)
	taskGroup.GET("/get/random/:type/level/:level", h.pages.GetRandomTask)
	taskGroup.GET("/get/bests/:type/level/:level/page/:page_index/size/:page_size", h.pages.GetBests)

	// func handlers

	taskGroup.POST("/inc/like/:id", h.IncLike)
	taskGroup.POST("/dec/like/:id", h.DecLike)
	taskGroup.POST("/inc/dislike/:id", h.IncDislike)
	taskGroup.POST("/dec/dislike/:id", h.DecDislike)

	taskGroup.POST("/add", h.AddTask, h.mw.Auth)
}
