package pagehandlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) Search(c *echo.Context) error {
	query := c.Param("query")
	
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
	if err != nil{
		switch err{
		case errors.ErrEmptyQuery:
			return c.String(http.StatusBadRequest, "empty query")
		case errors.ErrNotFound:
			return c.String(http.StatusNotFound, "not found tasks")
		default:
			return c.String(http.StatusInternalServerError, "internal server error")
		}
	}

	return renderWithStatus(c, http.StatusOK, pages.TasksPage(tasks, pageSize, pageIndex))
}
