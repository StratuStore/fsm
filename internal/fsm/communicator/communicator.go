package communicator

import (
	"context"
	"errors"
	"fmt"
	"github.com/StratuStore/fsm/internal/libs/config"
	"github.com/StratuStore/fsm/internal/libs/handler"
	"github.com/StratuStore/fsm/internal/libs/ownerrors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2"
	"github.com/mbretter/go-mongodb/types"
	"log/slog"
	"net/http"
	"sync"
)

const serviceAccountID = "fs"

type Communicator struct {
	l     *slog.Logger
	pub   *amqp.Publisher
	topic string
	host  string
	m     sync.Map
	g     GobMarshaler
}

func New(l *slog.Logger, cfg *config.Config) (*Communicator, error) {
	publisher, err := amqp.NewPublisher(
		amqp.NewNonDurablePubSubConfig(cfg.RabbitMQ.URN, nil),
		watermill.NewSlogLogger(l.With(slog.String("module", "watermill-ampq"))),
	)
	if err != nil {
		return nil, err
	}

	return &Communicator{
		l:     l.With(slog.String("module", "internal.fsm.communicator")),
		pub:   publisher,
		topic: cfg.Topic,
		host:  cfg.RabbitMQ.Host,
	}, nil
}

func (c *Communicator) Handler(ctx *fiber.Ctx) error {
	l := c.l.With(slog.String("op", "Handler"))

	if userID, err := handler.GetUserID(l, ctx); err != nil || userID != serviceAccountID {
		l.Error("unauthorised request to communicator occurred", slog.String("userID", userID), slog.Any("err", err))

		return ownerrors.NewUnauthorizedError(slog.New(slog.DiscardHandler), "unauthorized", "unauthorized")
	}

	var r Response
	if err := c.g.Unmarshal(ctx.Body(), &r); err != nil {
		return ownerrors.NewValidationError(l, "unable to parse request body with gob", "wrong data format", err)
	}

	process, ok := c.m.Load(r.ID)
	if !ok {
		return ownerrors.NewValidationError(l, "requestID is not found", "wrong requestID")
	}
	p := process.(*Process)

	if ok := p.Set(&r); !ok {
		return ctx.SendStatus(http.StatusResetContent)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

func (c *Communicator) Delete(ctx context.Context, id types.ObjectId) error {
	request, err := NewRequest(DeleteType, c.host, id, 0)
	if err != nil {
		return fmt.Errorf("unable to create request type: %w", err)
	}

	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return err
	}
	if response.Err != "" {
		return errors.New(response.Err)
	}

	return nil
}

func (c *Communicator) Create(ctx context.Context, id types.ObjectId, size uint) (host, connectionID string, err error) {
	request, err := NewRequest(CreateType, c.host, id, size)
	if err != nil {
		return "", "", fmt.Errorf("unable to create request type: %w", err)
	}

	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return "", "", err
	}

	return response.ToReturn()
}

func (c *Communicator) Open(ctx context.Context, id types.ObjectId) (host, connectionID string, err error) {
	request, err := NewRequest(OpenType, c.host, id, 0)
	if err != nil {
		return "", "", fmt.Errorf("unable to create request type: %w", err)
	}

	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return "", "", err
	}

	return response.ToReturn()
}

func (c *Communicator) Update(ctx context.Context, id types.ObjectId, size uint) (host, connectionID string, err error) {
	request, err := NewRequest(UpdateType, c.host, id, size)
	if err != nil {
		return "", "", fmt.Errorf("unable to create request type: %w", err)
	}

	response, err := c.makeRequest(ctx, request)
	if err != nil {
		return "", "", err
	}

	return response.ToReturn()
}

func (c *Communicator) makeRequest(ctx context.Context, r *Request) (*Response, error) {
	l := c.l.With(slog.String("op", "makeRequest"))

	p := NewProcess()
	if _, loaded := c.m.LoadOrStore(r.ID, p); loaded {
		l.Error("should not be possible: duplicate of request inside map occurred", slog.Any("id", r.ID))

		return nil, fmt.Errorf("should not be possible: duplicate of request inside map occurred")
	}

	payload, err := c.g.Marshal(r)
	if err != nil {
		l.Error("cannot marshal data using gob", slog.Any("data", r), slog.String("err", err.Error()))

		return nil, fmt.Errorf("cannot marshal data using gob: %w", err)
	}
	if err := c.pub.Publish(c.topic, message.NewMessage(r.ID.String(), payload)); err != nil {
		l.Error("unable to send message to queue", slog.String("err", err.Error()))

		return nil, fmt.Errorf("unable to send message to queue: %w", err)
	}

	response, err := p.WaitAndGet(ctx)
	if err != nil {
		l.Error("cannot get response from FS", slog.String("err", err.Error()))

		return nil, err
	}

	return response, nil
}
