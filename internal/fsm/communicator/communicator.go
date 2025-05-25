package communicator

import "context"

type Communicator struct {
}

func New() *Communicator {
	return &Communicator{}
}

func (c *Communicator) Delete(ctx context.Context, id string) error {
	return nil
}

func (c *Communicator) Create(ctx context.Context, id string) (host, connectionID string, err error) {
	return "localhost:5675", "3f122594-a43a-474b-85d2-36a8ce98ef2e", nil
}

func (c *Communicator) Open(ctx context.Context, id string) (host, connectionID string, err error) {
	return "localhost:5675", "3f122594-a43a-474b-85d2-36a8ce98ef2e", nil
}

func (c *Communicator) Update(ctx context.Context, id string) (host, connectionID string, err error) {
	return "localhost:5675", "3f122594-a43a-474b-85d2-36a8ce98ef2e", nil
}
