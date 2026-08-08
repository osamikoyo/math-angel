package handler

import (
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/user/handler/middleware"
	"github.com/osamikoyo/math-angel/internal/user/handler/pagehandlers"
	"github.com/osamikoyo/math-angel/internal/user/service"
)

type Handler struct {
	mw           *middleware.Middleware
	service      *service.Service
	pagehandlers *pagehandlers.PageHandler
}

func NewHandler(mw *middleware.Middleware, service *service.Service) *Handler {
	return &Handler{
		mw:           mw,
		service:      service,
		pagehandlers: pagehandlers.NewPageHandler(service),
	}
}

func (h *Handler) RegisterRouters(e *echo.Echo) {
	// page handlers
	e.GET("/profile", h.pagehandlers.GetProfile, h.mw.Auth)
	e.GET("/register", h.pagehandlers.RegisterPage)
	e.GET("/login", h.pagehandlers.Login)
	e.GET("/top/page/:page_index/size/:page_size", h.pagehandlers.GetTop)

	// func handlers
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)
	e.POST("/task/solved/:task_id", h.TaskSolved, h.mw.MayAuth)
	e.POST("/refresh", h.RefreshToken, h.mw.Auth)
}
