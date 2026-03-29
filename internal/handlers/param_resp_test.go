package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/psaraiva/time-trial/internal/entities"
)

const validParamRespBody = `{
	"statusCode": 200,
	"item": {
		"isColection": true,
		"quantity": 3,
		"properties": [{
			"name": "productName",
			"type": "string",
			"isRequired": true,
			"maxLength": 10,
			"minLength": 3,
			"propertyString": {"chars": "abcABC"}
		},{
			"name": "value",
			"type": "float",
			"isRequired": true,
			"maxLength": 10000,
			"minLength": 0,
			"propertyFloat": {"floatPrecision": 2, "isAcceptNegativeValue": false}
		},{
			"name": "version",
			"type": "int",
			"isRequired": true,
			"maxLength": 99,
			"minLength": 0,
			"propertyInt": {"isAcceptNegativeValue": false}
		}]
	}
}`

func newParamRespApp(t *testing.T) (*fiber.App, *entities.ParamResp) {
	t.Helper()
	pr := entities.NewParamResp()
	h := NewParamRespHandler(pr)
	app := fiber.New()
	app.Post("/param-resp", h.SetParamResp)
	app.Get("/param-resp/config", h.GetConfig)
	return app, pr
}

func TestParamRespHandler_SetParamResp_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		expectStatus   int
		expectActive   bool
		expectErrKey   bool
		expectPropCount int
	}{
		{
			name:         "empty body clears config",
			body:         "",
			expectStatus: http.StatusOK,
			expectActive: false,
		},
		{
			name:         "malformed json returns 400",
			body:         `{invalid}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "invalid statusCode returns 400",
			body:         `{"statusCode":500,"item":{"isColection":false,"quantity":1,"properties":[{"name":"f","type":"string","isRequired":true,"maxLength":5,"minLength":1,"propertyString":{"chars":"abc"}}]}}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "empty properties returns 400",
			body:         `{"statusCode":200,"item":{"isColection":false,"quantity":1,"properties":[]}}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "invalid property name returns 400",
			body:         `{"statusCode":200,"item":{"isColection":false,"quantity":1,"properties":[{"name":"bad name","type":"string","isRequired":true,"maxLength":5,"minLength":1,"propertyString":{"chars":"abc"}}]}}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:         "chars with digit returns 400",
			body:         `{"statusCode":200,"item":{"isColection":false,"quantity":1,"properties":[{"name":"f","type":"string","isRequired":true,"maxLength":5,"minLength":1,"propertyString":{"chars":"abc1"}}]}}`,
			expectStatus: http.StatusBadRequest,
			expectErrKey: true,
		},
		{
			name:            "valid config activates",
			body:            validParamRespBody,
			expectStatus:    http.StatusOK,
			expectActive:    true,
			expectPropCount: 3,
		},
		{
			name:         "valid single item (no collection)",
			body:         `{"statusCode":200,"item":{"isColection":false,"quantity":0,"properties":[{"name":"name","type":"string","isRequired":true,"maxLength":8,"minLength":2,"propertyString":{"chars":"abcABC"}}]}}`,
			expectStatus: http.StatusOK,
			expectActive: true,
			expectPropCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, _ := newParamRespApp(t)

			req := httptest.NewRequest(http.MethodPost, "/param-resp", strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}

			if tc.expectErrKey {
				if _, ok := body["error"]; !ok {
					t.Fatal("expected error key in response body")
				}
				return
			}

			if got := body["active"].(bool); got != tc.expectActive {
				t.Fatalf("expected active=%v, got %v", tc.expectActive, got)
			}

			if tc.expectActive {
				if got := int(body["properties"].(float64)); got != tc.expectPropCount {
					t.Fatalf("expected properties=%d, got %d", tc.expectPropCount, got)
				}
				if got := int(body["statusCode"].(float64)); got != 200 {
					t.Fatalf("expected statusCode=200, got %d", got)
				}
			}
		})
	}
}

func TestParamRespHandler_GetConfig_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func(*entities.ParamResp)
		expectStatus int
		expectErrKey bool
	}{
		{
			name:         "no active config returns 404",
			setup:        func(pr *entities.ParamResp) {},
			expectStatus: http.StatusNotFound,
			expectErrKey: true,
		},
		{
			name: "active config returns 200 with body",
			setup: func(pr *entities.ParamResp) {
				pr.Set(&entities.ResponseConfig{
					StatusCode: 200,
					Item: entities.ItemConfig{
						IsCollection: true,
						Quantity:     2,
						Properties: []entities.Property{
							{
								PropertyBase: entities.PropertyBase{
									Name: "f", Type: entities.PropertyTypeInt,
									IsRequired: true, MinLength: 0, MaxLength: 10,
								},
								PropertyInt: &entities.PropertyIntConfig{},
							},
						},
					},
				})
			},
			expectStatus: http.StatusOK,
			expectErrKey: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, pr := newParamRespApp(t)
			tc.setup(pr)

			req := httptest.NewRequest(http.MethodGet, "/param-resp/config", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tc.expectStatus {
				t.Fatalf("expected status %d, got %d", tc.expectStatus, resp.StatusCode)
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}

			if tc.expectErrKey {
				if _, ok := body["error"]; !ok {
					t.Fatal("expected error key in response body")
				}
				return
			}

			if _, ok := body["statusCode"]; !ok {
				t.Fatal("expected statusCode key in response body")
			}
			if _, ok := body["item"]; !ok {
				t.Fatal("expected item key in response body")
			}
		})
	}
}
