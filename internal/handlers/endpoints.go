package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/payout-service/internal/dto"
	"github.com/kodra-pay/payout-service/internal/services"
)

type PayoutHandler struct {
	svc *services.PayoutService
}

func NewPayoutHandler(svc *services.PayoutService) *PayoutHandler { return &PayoutHandler{svc: svc} }

func (h *PayoutHandler) Create(c *fiber.Ctx) error {
	var req dto.PayoutRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	return c.JSON(h.svc.Create(c.Context(), req))
}

func (h *PayoutHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(h.svc.Get(c.Context(), id))
}

func (h *PayoutHandler) List(c *fiber.Ctx) error {
	merchantID := c.Query("merchant_id")
	return c.JSON(h.svc.List(c.Context(), merchantID))
}

func (h *PayoutHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(h.svc.Cancel(c.Context(), id))
}
