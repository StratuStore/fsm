package utils

import (
	"github.com/StratuStore/fsm/internal/libs/errors"
)

type Response[T any] struct {
	Error string `json:"error,omitempty"`
	Body  T      `json:"body,omitempty"`
}

func NewResponseByUserError(err ownerrors.UserError) *Response[any] {
	return &Response[any]{
		Error: err.UserMessage(),
	}
}

func NewErrorResponse(msg string) *Response[any] {
	return &Response[any]{
		Error: msg,
	}
}

func NewOKResponse[T any](body T) *Response[T] {
	return &Response[T]{
		Body: body,
	}
}
