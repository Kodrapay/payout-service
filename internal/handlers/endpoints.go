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
	resp, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *PayoutHandler) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payout ID")
	}
	return c.JSON(h.svc.Get(c.Context(), id))
}

func (h *PayoutHandler) List(c *fiber.Ctx) error {
	merchantID := c.QueryInt("merchant_id", 0) // Use c.QueryInt for query parameters
	if merchantID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "merchant_id query parameter is required")
	}
	return c.JSON(h.svc.List(c.Context(), merchantID))
}

func (h *PayoutHandler) Cancel(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payout ID")
	}
	return c.JSON(h.svc.Cancel(c.Context(), id))
}

func (h *PayoutHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payout ID")
	}
	var req dto.PayoutStatusUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.svc.UpdateStatus(c.Context(), id, req.Status)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(resp)
}
