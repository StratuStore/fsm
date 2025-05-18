package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

type InputFunc func(*fiber.Ctx, any) error

func BodyInput(c *fiber.Ctx, input any) error {
	return c.BodyParser(input)
}

func QueryInput(c *fiber.Ctx, input any) error {
	return c.QueryParser(input)
}

func ParamsInput(c *fiber.Ctx, input any) error {
	return c.ParamsParser(input)
}

func ParamAndQueryInput(c *fiber.Ctx, input any) error {
	return errors.Join(c.QueryParser(input), c.ParamsParser(input))
}

func NoInput(c *fiber.Ctx, input any) error {
	return nil
}
