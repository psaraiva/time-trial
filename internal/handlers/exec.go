package handlers

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

var randomCodes = []int{200, 400, 500}

type ExecHandler struct {
	state *entities.State
	plan  *entities.Plan
}

func NewExecHandler(state *entities.State, plan *entities.Plan) *ExecHandler {
	return &ExecHandler{state: state, plan: plan}
}

// Exec simulates a dependent service response based on the active entities configuration.
//
// GET /exec
//   - code=0   → random response (200, 400 or 500)
//   - code=200 → returns 200
//   - code=400 → returns 400
//   - code=500 → returns 500
//
// If delayMin/delayMax are set, a random delay in that range is applied before responding.
func (h *ExecHandler) Exec(c *fiber.Ctx) error {
	code := int(h.state.GetCode())
	if code == 0 {
		code = randomCodes[rand.Intn(len(randomCodes))]
	}

	applyDelay(h.state)

	return c.Status(code).JSON(fiber.Map{
		"sabotaged": h.state.GetCode() != 0,
		"code":      code,
	})
}

// ExecPlan executes the next step of the active plan.
// Returns 404 if no plan is loaded or all steps have been consumed.
//
// GET /exec/plan
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

	return c.Status(int(s.GetCode())).JSON(fiber.Map{
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
