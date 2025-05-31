package service

import (
	"context"
	"github.com/mbretter/go-mongodb/types"
)

type Communicator interface {
	Delete(ctx context.Context, id types.ObjectId) error
	Create(ctx context.Context, id types.ObjectId, size uint) (host, connectionID string, err error)
	Open(ctx context.Context, id types.ObjectId) (host, connectionID string, err error)
	Update(ctx context.Context, id types.ObjectId, size uint) (host, connectionID string, err error)
}
