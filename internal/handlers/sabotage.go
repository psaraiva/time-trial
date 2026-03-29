package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

type SabotageHandler struct {
	state *entities.State
}

func NewSabotageHandler(state *entities.State) *SabotageHandler {
	return &SabotageHandler{state: state}
}

type sabotageRequest struct {
	Code     int `json:"code"`
	DelayMin int `json:"delayMin"`
	DelayMax int `json:"delayMax"`
}

// SetSabotage sets or resets the forced HTTP status code and optional delay range.
//
//	@Summary		Set sabotage configuration
//	@Description	Sets the active forced status code and optional delay range. With no body, resets to random behavior (code=0, delays=0).
//	@Tags			time-trial
//	@Accept			json
//	@Produce		json
//	@Param			body	body		sabotageRequest		false	"Sabotage configuration. Omit body to reset."
//	@Success		200		{object}	EnvelopeSetSabotage
//	@Failure		400		{object}	EnvelopeError
//	@Router			/time-trial [post]
func (h *SabotageHandler) SetSabotage(c *fiber.Ctx) error {
	if len(c.Body()) == 0 {
		h.state.Reset()
		return c.JSON(fiber.Map{"sabotaged": false, "code": 0, "delayMin": 0, "delayMax": 0})
	}

	var req sabotageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	switch req.Code {
	case 0:
		h.state.Reset()
		return c.JSON(fiber.Map{"sabotaged": false, "code": 0, "delayMin": 0, "delayMax": 0})
	case 200, 400, 500:
		// ok
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("invalid code %d: use 0 (reset), 200, 400 or 500", req.Code),
		})
	}

	h.state.SetCode(int32(req.Code))
	if err := h.state.SetDelay(int32(req.DelayMin), int32(req.DelayMax)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"sabotaged": true,
		"code":      req.Code,
		"delayMin":  req.DelayMin,
		"delayMax":  req.DelayMax,
	})
}
