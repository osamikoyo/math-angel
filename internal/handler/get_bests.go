package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *Handler) GetBests(c *echo.Context) error {
	taskType := c.Param("type")
	level := c.Param("level")

	pageIndexStr := c.Param("page_index")
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "page_index must be number")
	}

	pageSizeStr := c.Param("page_size")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "page_size must be number")
	}

	tasks, err := h.service.GetBests(c.Request().Context(),
		taskType,
		level,
		uint(pageSize),
		uint(pageIndex))
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound) {
			return c.String(http.StatusNotFound, "not found tasks")
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return renderWithStatus(c, http.StatusOK, pages.TasksPage(tasks, pageSize, pageIndex))
}
