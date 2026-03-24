package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
)

func (h *Handler) GetRandomTask(c *echo.Context) error {
	levelStr := c.Param("level")
	taskType := c.Param("type")

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "level must be number")
	}

	task, err := h.service.GetRandomTask(c.Request().Context(), taskType, uint(level))
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound ) {
			return c.String(http.StatusNotFound, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, task)
}
