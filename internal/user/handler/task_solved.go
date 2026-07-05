package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func (h *Handler) TaskSolved(c *echo.Context) error {
	log.Printf("routing task solved")

	idStr := c.Get("user_id")
	if idStr == nil {
		return c.String(http.StatusBadRequest, "empty user id")
	}

	id, err := strconv.Atoi(idStr.(string))
	if err != nil {
		return c.String(http.StatusBadRequest, "bad user id")
	}

	taskIDStr := c.Param("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad task uid")
	}

	if err = h.service.TaskSolved(c.Request().Context(), uint(id), taskID); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "success")
}
