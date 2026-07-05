package pagehandlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/ui/pages"
)

func (h *PageHandler) GetProfile(c *echo.Context) error {
	userIDAny := c.Get("user_id")
	if userIDAny == nil {
		return renderWithStatus(c, http.StatusOK, pages.Register())
	}

	userIDStr, ok := userIDAny.(string)
	if !ok {
		return c.String(http.StatusBadRequest, "bad user id")
	}

	log.Print("user id: ", userIDAny)

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.String(http.StatusBadGateway, err.Error())
	}

	profile, err := h.service.GetProfile(c.Request().Context(), uint(userID))
	if err != nil {
		if errors.Is(err, selferrors.ErrNotFound) {
			return c.String(http.StatusNotFound, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return renderWithStatus(c, http.StatusOK, pages.ProfilePage(profile))
}
