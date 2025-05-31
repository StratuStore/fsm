package ownerrors

import (
	"errors"
	"log/slog"
	"net/http"
)

const (
	InternalErrorMessage = "service temporary unavailable"
)

func NewError(l *slog.Logger, status int, debugMessage, userMessage string, errs ...error) error {
	err := errors.Join(errs...)
	l.Debug(debugMessage, slog.Any("err", err))

	return &OwnError{error: err, internalMsg: debugMessage, msg: userMessage, status: status}
}

func NewInternalError(l *slog.Logger, message string, errs ...error) error {
	err := errors.Join(errs...)
	l.Error(message, slog.Any("err", err))

	return &OwnError{error: err, internalMsg: message, msg: InternalErrorMessage, status: http.StatusInternalServerError}
}

func NewValidationError(l *slog.Logger, internalMessage, userMessage string, errs ...error) error {
	return NewError(l, http.StatusBadRequest, internalMessage, userMessage, errs...)
}

func NewNotFoundError(l *slog.Logger, internalMessage, userMessage string, errs ...error) error {
	return NewError(l, http.StatusNotFound, internalMessage, userMessage, errs...)
}

func NewUnauthorizedError(l *slog.Logger, internalMessage, userMessage string, errs ...error) error {
	return NewError(l, http.StatusUnauthorized, internalMessage, userMessage, errs...)
}
