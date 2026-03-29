package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

type ParamRespHandler struct {
	paramResp *entities.ParamResp
}

func NewParamRespHandler(paramResp *entities.ParamResp) *ParamRespHandler {
	return &ParamRespHandler{paramResp: paramResp}
}

// SetParamResp configures the dynamic response-body schema.
//
//	@Summary		Set param-resp configuration
//	@Description	Configures the schema used to generate response bodies when a 200 is returned by /sabotage or /plan/sabotage. Only statusCode=200 is supported. With no body, clears the active configuration (returns {active:false}).
//	@Tags			param-resp
//	@Accept			json
//	@Produce		json
//	@Param			body	body		entities.ResponseConfig		false	"Response schema configuration. Omit body to clear."
//	@Success		200		{object}	EnvelopeSetParamResp
//	@Failure		400		{object}	EnvelopeError
//	@Router			/param-resp [post]
func (h *ParamRespHandler) SetParamResp(c *fiber.Ctx) error {
	if len(c.Body()) == 0 {
		h.paramResp.Clear()
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"active": false,
		})
	}

	var config entities.ResponseConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := entities.ValidateResponseConfig(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	h.paramResp.Set(&config)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"active":     true,
		"statusCode": config.StatusCode,
		"properties": len(config.Item.Properties),
	})
}

// GetConfig returns the current param-resp configuration.
//
//	@Summary		Get param-resp configuration
//	@Description	Returns the active response-body schema exactly as submitted. Returns 404 if no configuration is active.
//	@Tags			param-resp
//	@Produce		json
//	@Success		200	{object}	EnvelopeParamRespConfig
//	@Failure		404	{object}	EnvelopeError
//	@Router			/param-resp/config [get]
func (h *ParamRespHandler) GetConfig(c *fiber.Ctx) error {
	config := h.paramResp.Get()
	if config == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no active param-resp configuration",
		})
	}
	return c.Status(fiber.StatusOK).JSON(config)
}
