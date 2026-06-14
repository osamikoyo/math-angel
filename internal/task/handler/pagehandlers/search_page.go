package pagehandlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) SearchPage(c *echo.Context) error {
	return renderWithStatus(c, http.StatusOK, pages.SearchPage())
}
