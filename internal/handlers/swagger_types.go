package handlers

// This file contains types used exclusively for Swagger schema generation.
// These structs are never instantiated at runtime — the actual responses are
// produced by handlers using fiber.Map and wrapped by the envelope middleware.

import "github.com/psaraiva/time-trial/internal/entities"

// --- Inner response types ---

// SabotageStateResponse is returned by GET /sabotage and GET /plan/sabotage
// when no param-resp configuration is active.
type SabotageStateResponse struct {
	Sabotaged bool `json:"sabotaged"`
	Code      int  `json:"code"`
}

// SabotageConfigResponse is returned by GET /time-trial/config.
type SabotageConfigResponse struct {
	Sabotaged bool  `json:"sabotaged"`
	Code      int32 `json:"code"`
	DelayMin  int32 `json:"delayMin"`
	DelayMax  int32 `json:"delayMax"`
}

// SetSabotageResponse is returned by POST /time-trial.
type SetSabotageResponse struct {
	Sabotaged bool `json:"sabotaged"`
	Code      int  `json:"code"`
	DelayMin  int  `json:"delayMin"`
	DelayMax  int  `json:"delayMax"`
}

// SetPlanResponse is returned by POST /plan when a plan is successfully loaded.
type SetPlanResponse struct {
	Active bool `json:"active"`
	Steps  int  `json:"steps"`
}

// PlanClearedResponse is returned by POST /plan with no body.
type PlanClearedResponse struct {
	Active bool `json:"active"`
}

// PlanStepConfig represents a single step inside PlanConfigResponse.
type PlanStepConfig struct {
	Code     int32 `json:"code"`
	DelayMin int32 `json:"delayMin"`
	DelayMax int32 `json:"delayMax"`
}

// PlanConfigResponse is returned by GET /plan/config.
type PlanConfigResponse struct {
	Active bool             `json:"active"`
	Steps  []PlanStepConfig `json:"steps"`
}

// SetParamRespResponse is returned by POST /param-resp when a config is successfully set.
type SetParamRespResponse struct {
	Active     bool `json:"active"`
	StatusCode int  `json:"statusCode"`
	Properties int  `json:"properties"`
}

// ParamRespClearedResponse is returned by POST /param-resp with no body.
type ParamRespClearedResponse struct {
	Active bool `json:"active"`
}

// ErrorResponse is returned by all endpoints on validation or not-found errors.
type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Envelope wrappers ---
// One wrapper per distinct Data type so that Swagger renders the full nested schema.
// The actual wrapping at runtime is performed by the envelope middleware in cmd/envelope.go.

// EnvelopeSabotageState wraps SabotageStateResponse.
type EnvelopeSabotageState struct {
	Data      SabotageStateResponse `json:"data"`
	Duration  int64                 `json:"duration"`
	Timestamp string                `json:"timestamp"`
}

// EnvelopeSabotageConfig wraps SabotageConfigResponse.
type EnvelopeSabotageConfig struct {
	Data      SabotageConfigResponse `json:"data"`
	Duration  int64                  `json:"duration"`
	Timestamp string                 `json:"timestamp"`
}

// EnvelopeSetSabotage wraps SetSabotageResponse.
type EnvelopeSetSabotage struct {
	Data      SetSabotageResponse `json:"data"`
	Duration  int64               `json:"duration"`
	Timestamp string              `json:"timestamp"`
}

// EnvelopeSetPlan wraps SetPlanResponse.
type EnvelopeSetPlan struct {
	Data      SetPlanResponse `json:"data"`
	Duration  int64           `json:"duration"`
	Timestamp string          `json:"timestamp"`
}

// EnvelopePlanCleared wraps PlanClearedResponse.
type EnvelopePlanCleared struct {
	Data      PlanClearedResponse `json:"data"`
	Duration  int64               `json:"duration"`
	Timestamp string              `json:"timestamp"`
}

// EnvelopePlanConfig wraps PlanConfigResponse.
type EnvelopePlanConfig struct {
	Data      PlanConfigResponse `json:"data"`
	Duration  int64              `json:"duration"`
	Timestamp string             `json:"timestamp"`
}

// EnvelopeSetParamResp wraps SetParamRespResponse.
type EnvelopeSetParamResp struct {
	Data      SetParamRespResponse `json:"data"`
	Duration  int64                `json:"duration"`
	Timestamp string               `json:"timestamp"`
}

// EnvelopeParamRespCleared wraps ParamRespClearedResponse.
type EnvelopeParamRespCleared struct {
	Data      ParamRespClearedResponse `json:"data"`
	Duration  int64                    `json:"duration"`
	Timestamp string                   `json:"timestamp"`
}

// EnvelopeParamRespConfig wraps entities.ResponseConfig.
type EnvelopeParamRespConfig struct {
	Data      entities.ResponseConfig `json:"data"`
	Duration  int64                   `json:"duration"`
	Timestamp string                  `json:"timestamp"`
}

// EnvelopeDynamic is used for exec endpoints when param-resp is active.
// The schema of Data depends on the active param-resp configuration.
// See GET /param-resp/config for the full schema definition.
type EnvelopeDynamic struct {
	Data      interface{} `json:"data"`
	Duration  int64       `json:"duration"`
	Timestamp string      `json:"timestamp"`
}

// EnvelopeError wraps ErrorResponse.
type EnvelopeError struct {
	Data      ErrorResponse `json:"data"`
	Duration  int64         `json:"duration"`
	Timestamp string        `json:"timestamp"`
}
