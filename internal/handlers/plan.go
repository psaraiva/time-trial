package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
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
// POST /entities/plan
//
//	{
//	  "plan": [
//	    { "code": 500, "delayMin": 500,  "delayMax": 900 },
//	    { "code": 200, "delayMin": 500,  "delayMax": 900 }
//	  ]
//	}
//
// POST /entities/plan (sem body) → limpa o plano ativo
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
// GET /plan/config
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
