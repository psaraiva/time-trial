package handlers

import (
	"math/rand"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
	"github.com/psaraiva/time-trial/internal/generator"
)

var randomCodes = []int{200, 400, 500}

type ExecHandler struct {
	state     *entities.State
	plan      *entities.Plan
	paramResp *entities.ParamResp
}

func NewExecHandler(state *entities.State, plan *entities.Plan, paramResp *entities.ParamResp) *ExecHandler {
	return &ExecHandler{state: state, plan: plan, paramResp: paramResp}
}

// Exec simulates a dependent service response based on the active entities configuration.
//
//	@Summary		Execute sabotage
//	@Description	Responds with the configured HTTP code (or random 200/400/500 when code=0). Applies delay if configured.
//	@Description	When code=200 and a param-resp config is active, "data" is dynamically generated — its schema depends on the active configuration. See GET /param-resp/config.
//	@Tags			sabotage
//	@Produce		json
//	@Success		200	{object}	EnvelopeDynamic			"data is either SabotageStateResponse or dynamically generated (see GET /param-resp/config)"
//	@Failure		400	{object}	EnvelopeSabotageState
//	@Failure		500	{object}	EnvelopeSabotageState
//	@Router			/sabotage [get]
func (h *ExecHandler) Exec(c *fiber.Ctx) error {
	code := int(h.state.GetCode())
	if code == 0 {
		code = randomCodes[rand.Intn(len(randomCodes))]
	}

	applyDelay(h.state)

	if cfg := h.paramResp.Get(); cfg != nil && code == cfg.StatusCode {
		return c.Status(code).JSON(generator.Generate(cfg))
	}

	return c.Status(code).JSON(fiber.Map{
		"sabotaged": h.state.GetCode() != 0,
		"code":      code,
	})
}
