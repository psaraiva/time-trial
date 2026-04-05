package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
	"github.com/psaraiva/time-trial/internal/generator"
)

type PlanHandler struct {
	plan *entities.Plan
}

func NewPlanHandler(plan *entities.Plan) *PlanHandler {
	return &PlanHandler{plan: plan}
}

type planStepRequest struct {
	Code     int `json:"code"`
	DelayMin int `json:"delayMin"`
	DelayMax int `json:"delayMax"`
}

type planRequest struct {
	Plan []planStepRequest `json:"plan"`
}

// SetPlan loads a new plan into memory or clears the active one if body is empty.
//
//	@Summary		Set or clear plan
//	@Description	Loads an ordered list of sabotage steps. Steps are consumed in order on GET /plan/sabotage. With no body, clears the active plan (returns {active:false}).
//	@Tags			plan
//	@Accept			json
//	@Produce		json
//	@Param			body	body		planRequest		false	"Ordered plan. Omit body to clear."
//	@Success		200		{object}	EnvelopeSetPlan
//	@Failure		400		{object}	EnvelopeError
//	@Router			/plan [post]
func (h *PlanHandler) SetPlan(c *fiber.Ctx) error {
	if len(c.Body()) == 0 {
		h.plan.Clear()
		return c.JSON(fiber.Map{"active": false})
	}

	var req planRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Plan) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan must have at least one step",
		})
	}

	states := make([]*entities.State, len(req.Plan))
	for i, step := range req.Plan {
		switch step.Code {
		case 200, 400, 500:
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("step %d: invalid code %d: use 200, 400 or 500", i, step.Code),
			})
		}
		s := entities.NewState()
		s.SetCode(int32(step.Code))
		if err := s.SetDelay(int32(step.DelayMin), int32(step.DelayMax)); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("step %d: %s", i, err.Error()),
			})
		}
		states[i] = s
	}

	h.plan.Set(states)
	return c.JSON(fiber.Map{"active": true, "steps": len(states)})
}

// GetConfig returns the full active plan if one exists, or 404 if not.
//
//	@Summary		Get plan configuration
//	@Description	Returns all steps of the active plan regardless of how many have been consumed. Returns 404 if no plan is loaded.
//	@Tags			plan
//	@Produce		json
//	@Success		200	{object}	EnvelopePlanConfig
//	@Failure		404	{object}	EnvelopeError
//	@Router			/plan/config [get]
func (h *PlanHandler) GetConfig(c *fiber.Ctx) error {
	states, ok := h.plan.IsActive()
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no active plan",
		})
	}

	steps := make([]fiber.Map, len(states))
	for i, s := range states {
		delayMin, delayMax := s.GetDelay()
		steps[i] = fiber.Map{
			"code":     s.GetCode(),
			"delayMin": delayMin,
			"delayMax": delayMax,
		}
	}

	return c.JSON(fiber.Map{
		"active": true,
		"steps":  steps,
	})
}

// ExecPlan executes the next step of the active plan.
//
//	@Summary		Execute next plan step
//	@Description	Consumes the next step from the active plan. Returns 404 if no plan is loaded, all steps are consumed, or the plan was interrupted.
//	@Description	When step code=200 and a param-resp config is active, "data" is dynamically generated — its schema depends on the active configuration. See GET /param-resp/config.
//	@Tags			plan
//	@Produce		json
//	@Success		200	{object}	EnvelopeDynamic			"data is either SabotageStateResponse or dynamically generated (see GET /param-resp/config)"
//	@Failure		400	{object}	EnvelopeSabotageState
//	@Failure		404	{object}	EnvelopeError
//	@Failure		500	{object}	EnvelopeSabotageState
//	@Router			/plan/sabotage [get]
func (h *ExecHandler) ExecPlan(c *fiber.Ctx) error {
	s, ok := h.plan.Next()
	if !ok {
		if h.plan.IsCancelled() {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "plan interrupted",
			})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no active plan or all steps have been consumed",
		})
	}

	applyDelay(s)

	code := int(s.GetCode())
	if cfg := h.paramResp.Get(); cfg != nil && code == cfg.StatusCode {
		return c.Status(code).JSON(generator.Generate(cfg))
	}

	return c.Status(code).JSON(fiber.Map{
		"sabotaged": true,
		"code":      s.GetCode(),
	})
}

// applyDelay sleeps for a random duration within the state's delay range, if set.
func applyDelay(s *entities.State) {
	delayMin, delayMax := s.GetDelay()
	if delayMax == 0 {
		return
	}
	delay := delayMin
	if delayMax > delayMin {
		delay = delayMin + int32(rand.Intn(int(delayMax-delayMin+1)))
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
}
