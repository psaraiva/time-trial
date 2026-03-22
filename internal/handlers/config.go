package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

type ConfigHandler struct {
	state *entities.State
}

func NewConfigHandler(state *entities.State) *ConfigHandler {
	return &ConfigHandler{state: state}
}

// GetConfig returns the current sabotage configuration.
//
// GET /config
func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	code := h.state.GetCode()
	delayMin, delayMax := h.state.GetDelay()
	return c.JSON(fiber.Map{
		"sabotaged": code != 0,
		"code":      code,
		"delayMin":  delayMin,
		"delayMax":  delayMax,
	})
}
