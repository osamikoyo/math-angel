package middleware

import (
	"github.com/osamikoyo/math-angel/internal/user/service"
)

type Middleware struct {
	service *service.Service
}

func NewMiddleware(svc *service.Service) *Middleware {
	return &Middleware{
		service: svc,
	}
}
