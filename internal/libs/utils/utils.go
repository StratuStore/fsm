package utils

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/StratuStore/fsm/internal/libs/ownerrors"
)

func ProcessError(l *slog.Logger, c *fiber.Ctx, err error) error {
	var userErr ownerrors.UserError
	if errors.As(err, &userErr) {
		l.Debug("unable to execute service inside handler", slog.String("err", err.Error()))

		return c.Status(userErr.Status()).JSON(NewResponseByUserError(userErr))
	}
	l.Error("unable to execute service inside handler (unknown error)", slog.String("err", err.Error()))

	return c.Status(http.StatusInternalServerError).JSON(NewErrorResponse("internal error"))
}
