package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *Handler) IncLike(c *echo.Context) error {
	id := c.Param("id")

	if err := h.service.IncLike(c.Request().Context(), id); err != nil {
		switch err {
		case errors.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		case errors.ErrNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	task, err := h.service.GetTask(c.Request().Context(), id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error")
	}

	return renderWithStatus(c, http.StatusOK, pages.LikeGroup(&pages.Task{
		ID:    task.UID,
		Likes: int(task.Likes),
	}))
}
