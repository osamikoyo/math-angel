package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/errors"
)

func (h *Handler) DecDislike(c *echo.Context) error {
	id := c.Param("id")

	if err := h.service.DecDislike(c.Request().Context(), id);err != nil{
		switch err {
		case errors.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		case errors.ErrNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return c.String(http.StatusOK, "success")
}