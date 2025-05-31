package service

import (
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"log/slog"
)

func NewWrongUserError(l *slog.Logger, errs ...error) error {
	return ownerrors.NewValidationError(l, "userID is not equal", "wrong user", errs...)
}

func NewDBError(l *slog.Logger, err error) error {
	if err != nil {
		return ownerrors.NewNotFoundError(l, "db error", "not found", err)
	}

	return nil
}
