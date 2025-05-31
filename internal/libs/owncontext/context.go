package owncontext

import (
	"context"
)

type Context interface {
	context.Context
	UserID() string
}

type ctx struct {
	context.Context
	userID string
}

func (c *ctx) UserID() string {
	return c.userID
}

func New(c context.Context, userID string) Context {
	return &ctx{
		Context: c,
		userID:  userID,
	}
}
