package service

import "context"

type Communicator interface {
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, id string) (host, connectionID string, err error)
	Open(ctx context.Context, id string) (host, connectionID string, err error)
	Update(ctx context.Context, id string) (host, connectionID string, err error)
}
