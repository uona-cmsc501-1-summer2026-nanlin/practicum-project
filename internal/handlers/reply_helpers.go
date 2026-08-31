package handlers

import (
	"errors"
	"net/http"

	"github.com/gobeetle/reply"
	"github.com/gofiber/fiber/v2"
)

func respond(c *fiber.Ctx, payload any) error {
	return reply.NewFiber(c).JSON(payload)
}

func respondCreated(c *fiber.Ctx, data any) error {
	return respond(c, reply.NewData(data).WithCode(http.StatusCreated))
}

func respondOK(c *fiber.Ctx, data any) error {
	return respond(c, reply.NewData(data))
}

func respondNoContent(c *fiber.Ctx) error {
	return reply.NewFiber(c).Empty()
}

func errBadRequest(msg string) error {
	return reply.InvalidRequest(errors.New(msg))
}

func errValidation(msg string) error {
	return reply.ValidationFailed(errors.New(msg))
}

func errNotFound(msg string) error {
	return reply.NotFound(errors.New(msg))
}

func errInternal(msg string) error {
	return reply.Internal(errors.New(msg))
}
