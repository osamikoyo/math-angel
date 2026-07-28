package pagehandlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) GetTop(c *echo.Context) error {
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

	if pageIndex < 1 {
		pageIndex = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	top, total, err := h.service.GetUserTop(c.Request().Context(), uint(pageIndex), uint(pageSize))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return renderWithStatus(c, http.StatusOK, pages.TopPage(top, pageIndex, pageSize, totalPages))
}
