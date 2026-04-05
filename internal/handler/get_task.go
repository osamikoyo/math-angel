package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *Handler) GetTask(c *echo.Context) error {
	id := c.Param("id")

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		switch err {
		case errors.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		case errors.ErrNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return renderWithStatus(c, http.StatusOK, pages.TaskPage(&pages.Task{
		Type:     task.Type,
		ID:       task.ID.String(),
		Level:    task.Level,
		Problem:  task.Problem,
		Solution: task.Solution,
		Boxed:    task.Boxed,
		Likes:    int(task.Likes),
		Dislikes: int(task.Dislikes),
	}))
}
