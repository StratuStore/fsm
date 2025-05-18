package service

import (
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"log/slog"
)

func NewWrongUserError(l *slog.Logger, errs ...error) error {
	return ownerrors.NewValidationError(l, "userID is not equal", "wrong user", errs...)
}

func NewDBError(l *slog.Logger, err error) error {
	return err
}
