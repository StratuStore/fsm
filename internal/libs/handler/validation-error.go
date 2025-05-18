package handler

import (
	"errors"
	"fmt"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	next error
	validator.FieldError
}

func (e ValidationError) Error() string {
	msg := e.FieldError.Error()

	if e.next != nil {
		msg += "\n " + e.next.Error()
	}

	return msg
}

func (e ValidationError) Unwrap() error {
	return e.next
}

func (e ValidationError) UserMessage() string {
	msg := fmt.Sprintf("%v validation failed, must be %v, got %v", e.Field(), e.ActualTag(), e.Param())

	var userError ownerrors.UserError
	if errors.As(e.next, &userError) {
		msg += ", " + userError.UserMessage()
	}

	return msg
}

func (e ValidationError) Status() int {
	return http.StatusBadRequest
}

func NewValidationError(err error) ownerrors.UserError {
	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		err := ownerrors.New(
			err,
			"provided data is empty",
			"provided data is empty",
			http.StatusBadRequest,
		)

		return err
	}

	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		for _, e := range validateErrs {
			err = &ValidationError{
				next:       err,
				FieldError: e,
			}
		}
	}

	return ownerrors.New(err, "unknown error", "internal server error", http.StatusInternalServerError)
}
