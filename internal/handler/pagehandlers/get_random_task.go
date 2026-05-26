package pagehandlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) GetRandomTask(c *echo.Context) error {
	level := c.Param("level")
	taskType := c.Param("type")

	task, err := h.service.GetRandomTask(c.Request().Context(), taskType, level)
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound) {
			return renderWithStatus(c, http.StatusNotFound, pages.NotFound())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return renderWithStatus(c, http.StatusOK, pages.TaskPage(&pages.Task{
		ID:       task.UID,
		Type:     task.Type,
		Boxed:    task.Boxed,
		Level:    task.Level,
		Solution: task.Solution,
		Problem:  task.Problem,
		Likes:    int(task.Likes),
		Dislikes: int(task.Dislikes),
	}))
}
