package errors

import (
	"errors"
	"fmt"
)

type UserError interface {
	error
	Status() int
	UserMessage() string
}

type OwnError struct {
	error
	internalMsg string
	msg         string
	status      int
}

func New(err error, internalMsg string, msg string, status int) UserError {
	return &OwnError{
		error:       err,
		internalMsg: internalMsg,
		msg:         msg,
		status:      status,
	}
}

func (e *OwnError) Unwrap() error {
	return e.error
}

func (e *OwnError) Error() string {
	msg := e.internalMsg

	if e.msg != "" {
		msg += fmt.Sprintf(" (for user: %s with status code %v)", e.msg, e.status)
	}

	if e.error != nil {
		msg += fmt.Sprintf("\n %v", e.error)
	}

	return msg
}

func (e *OwnError) Status() int {
	return e.status
}

func (e *OwnError) UserMessage() string {
	msg := e.msg

	var userError UserError
	if errors.As(e.error, &userError) {
		msg += "\n " + userError.UserMessage()
	}

	return msg
}
