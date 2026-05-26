package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *Handler) AddTask(c *echo.Context) error {
	var req struct {
		Type     string `form:"type"`
		Problem  string `form:"problem"`
		Solution string `form:"solution"`
		Boxed    string `form:"boxed"`
		Level    string `form:"level"`
	}

	if err := c.Bind(&req); err != nil {
		return renderWithStatus(c, http.StatusBadRequest,
			pages.ResultMessage("Error in form data", true))
	}

	err := h.service.CreateTask(c.Request().Context(), req.Type, req.Problem, req.Solution, req.Boxed, req.Level)
	if err != nil {
		if errors.Is(err, selferrors.ErrAlreadyExist) {
			return renderWithStatus(c, http.StatusBadRequest,
				pages.ResultMessage("Already exist", true))
		}
		return renderWithStatus(c, http.StatusInternalServerError,
			pages.ResultMessage("Internal server error", true))
	}

	c.Request().Header.Set("HX-Redirect", "/")

	return renderWithStatus(c, http.StatusOK,
		pages.ResultMessage("Task was successfully added", false))
}
