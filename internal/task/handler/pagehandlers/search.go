package pagehandlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/task/model"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) Search(c *echo.Context) error {
	query := c.QueryParam("query")

	log.Printf("query is :%s", query)

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

	tasks, err := h.service.Search(c.Request().Context(), query, pageIndex, pageSize)
	if err != nil {
		switch err {
		case errors.ErrEmptyQuery:
			return c.String(http.StatusBadRequest, "empty query")
		case errors.ErrNotFound:
			return renderWithStatus(c, http.StatusNotFound, pages.NotFound())
		default:
			return c.String(http.StatusInternalServerError, "internal server error")
		}
	}

	return renderWithStatus(c, http.StatusOK,
		pages.TasksPage(
			taskSearchResultToTasks(tasks),
			pageSize,
			pageIndex),
	)
}

func taskSearchResultToTasks(sr []model.TaskSearchResult) []model.Task {
	tasks := make([]model.Task, len(sr))

	for i := range sr {
		tasks[i] = sr[i].Task
	}

	return tasks
}
