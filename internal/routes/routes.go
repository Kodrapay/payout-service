package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/payout-service/internal/config"
	"github.com/kodra-pay/payout-service/internal/handlers"
	"github.com/kodra-pay/payout-service/internal/repositories"
	"github.com/kodra-pay/payout-service/internal/services"
)

func Register(app *fiber.App, cfg config.Config) {
	health := handlers.NewHealthHandler(cfg.ServiceName)
	health.Register(app)

	repo, err := repositories.NewPayoutRepository(cfg.PostgresDSN)
	if err != nil {
		panic(err)
	}
	svc := services.NewPayoutService(repo, cfg.MerchantServiceURL)
	handler := handlers.NewPayoutHandler(svc)

	app.Get("/payouts", handler.List)
	app.Post("/payouts", handler.Create)
	app.Get("/payouts/:id", handler.Get)
	app.Put("/payouts/:id/status", handler.UpdateStatus)
}
