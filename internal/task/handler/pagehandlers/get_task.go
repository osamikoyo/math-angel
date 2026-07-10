package pagehandlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) GetTask(c *echo.Context) error {
	id := c.Param("id")

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		switch err {
		case selferrors.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		case selferrors.ErrNotFound:
			return renderWithStatus(c, http.StatusNotFound, pages.NotFound())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	var (
		solved   = false
		solvedAt = time.Time{}
	)

	user_idAny := c.Get("user_id")
	if user_idAny != nil {
		user_id, err := strconv.Atoi(user_idAny.(string))
		if err != nil {
			return c.String(http.StatusBadRequest, "bad user id")
		}

		sAt, err := h.service.TaskSolvedBy(c.Request().Context(), task.UID, uint(user_id))
		if err == nil {
			solved = true
			solvedAt = sAt
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

		Solved:   solved,
		SolvedAt: solvedAt,
	}))
}
