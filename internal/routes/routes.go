package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/payout-service/internal/handlers"
	"github.com/kodra-pay/payout-service/internal/services"
)

func Register(app *fiber.App, service string) {
	health := handlers.NewHealthHandler(service)
	health.Register(app)

	svc := services.NewPayoutService()
	h := handlers.NewPayoutHandler(svc)
	api := app.Group("/payouts")
	api.Post("/", h.Create)
	api.Get("/:id", h.Get)
	api.Post("/:id/cancel", h.Cancel)
}
