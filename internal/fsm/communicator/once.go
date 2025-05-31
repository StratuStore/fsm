package communicator

import (
	"context"
	"sync"
)

type Process struct {
	once   sync.Once
	result chan *Response
}

func NewProcess() *Process {
	return &Process{
		result: make(chan *Response),
	}
}

func (r *Process) Set(response *Response) bool {
	var b bool
	r.once.Do(func() {
		b = true

		r.result <- response
		close(r.result)
	})

	return b
}

func (r *Process) WaitAndGet(ctx context.Context) (*Response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-r.result:
		return result, nil
	}
}
