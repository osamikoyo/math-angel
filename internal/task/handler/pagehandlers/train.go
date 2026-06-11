package pagehandlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) StartTrain(c *echo.Context) error {
	emptyprms := 0

	pageStr := c.QueryParam("page")
	pageIndex := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			pageIndex = p
		}
	}

	sizeStr := c.QueryParam("size")
	pageSize := 12
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 50 {
			pageSize = s
		}
	}

	taskType := c.QueryParam("type")
	if taskType == "" {
		taskType = "all"

		emptyprms++
	}

	level := c.QueryParam("level")
	if level == "" {
		level = "all"

		emptyprms++
	}

	tasks, err := h.service.GetTasks(c.Request().Context(), taskType, level, uint(pageSize), uint(pageIndex))
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound) {
			return renderWithStatus(c, http.StatusOK, pages.NotFoundComponent())
		}

		return c.String(http.StatusInternalServerError, "internal service error")
	}

	if emptyprms >= 2 {
		return renderWithStatus(c, http.StatusOK,
			pages.TrainPage(tasks, pageSize, pageIndex, taskType, level))
	}

	return renderWithStatus(c, http.StatusOK,
		pages.TasksGrid(tasks, pageSize, pageIndex, taskType, level))
}
