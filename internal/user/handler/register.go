package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
)

type RegisterRequest struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

func (h *Handler) Register(c *echo.Context) error {
	var req RegisterRequest

	if err := c.Bind(&req);err != nil{
		return c.String(http.StatusBadRequest, "failed parse request")
	}

	if err := h.service.RegisterUser(c.Request().Context(), req.Username, req.Email, req.Password);err != nil{
		if errors.Is(err, selferrors.ErrAlreadyExist) {
			return c.String(http.StatusConflict, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusCreated, "success")
}
