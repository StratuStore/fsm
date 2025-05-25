package handler

import (
	"context"
	"github.com/StratuStore/fsm/internal/libs/config"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log/slog"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Handler struct {
	app              *fiber.App
	l                *slog.Logger
	cfg              *config.Config
	fileHandler      *FileHandler
	directoryHandler *DirectoryHandler
}

func New(
	l *slog.Logger,
	cfg *config.Config,
	fileHandler *FileHandler,
	directoryHandler *DirectoryHandler,
) *Handler {
	h := &Handler{
		app: fiber.New(fiber.Config{
			IdleTimeout:  cfg.IdleTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		}),
		l:                l.With(slog.String("module", "internal.fsm.handler")),
		cfg:              cfg,
		fileHandler:      fileHandler,
		directoryHandler: directoryHandler,
	}

	h.Register()

	return h
}

func (h *Handler) Register() {
	h.registerDefaults()

	h.fileHandler.Register(h.app, "/file")
	h.directoryHandler.Register(h.app, "/directory")
}

func (h *Handler) registerDefaults() {
	if h.cfg.Env == "dev" {
		h.app.Use(cors.New(cors.ConfigDefault))
	} else {
		h.app.Use(cors.New(cors.Config{
			AllowOrigins: h.cfg.CORSOrigins,
		}))
	}

	h.app.Use(recover.New(recover.Config{
		StackTraceHandler: func(c *fiber.Ctx, r any) {
			h.l.Error("fiber panicked", slog.Any("err", r))
		},
	}))

	h.app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true // TODO: service check
		},
		ReadinessEndpoint: "/ready",
	}))

	h.app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.HS512,
			Key:    []byte(h.cfg.JWTSecret),
		},
	}))
}

func (h *Handler) Start(_ context.Context) error {
	l := h.l.With("op", "internal.fsm.handler.Start")

	addr := net.JoinHostPort(h.cfg.Handler.Host, h.cfg.Handler.Port)

	go func() {
		if err := h.app.Listen(addr); err != nil {
			l.Error("server error", slog.String("err", err.Error()))
		}
	}()

	return nil
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.app.ShutdownWithContext(ctx)
}
