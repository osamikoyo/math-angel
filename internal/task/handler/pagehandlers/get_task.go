package pagehandlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) GetTask(c *echo.Context) error {
	id := c.Param("id")

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		switch err {
		case errors.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		case errors.ErrNotFound:
			return renderWithStatus(c, http.StatusNotFound, pages.NotFound())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return renderWithStatus(c, http.StatusOK, pages.TaskPage(&pages.Task{
		Type:     task.Type,
		ID:       task.UID,
		Level:    task.Level,
		Problem:  task.Problem,
		Solution: task.Solution,
		Boxed:    task.Boxed,
		Likes:    int(task.Likes),
		Dislikes: int(task.Dislikes),
	}))
}
