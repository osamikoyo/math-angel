package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/model"
)

func (h *Handler) AddTask(c *echo.Context) error {
	var task model.Task
	if err := c.Bind(&task); err != nil {
		return c.String(http.StatusBadRequest, "bad task: "+err.Error())
	}

	err := h.service.CreateTask(
		c.Request().Context(),
		task.Type,
		task.Problem,
		task.Solution,
		task.Boxed,
		task.Level)
	
	if err != nil{
		if errors.Is(err, selferrors.ErrAlreadyExist) {
			return c.String(http.StatusBadRequest, "already exist")
		}

		return c.String(http.StatusInternalServerError, "internal error")
	}

	return c.String(http.StatusOK, "ok")
}
