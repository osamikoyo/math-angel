package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) IncLike(c *echo.Context) error {
	id := c.Param("id")

	if err := h.service.IncLike(c.Request().Context(), id);err != nil{
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "success")
}