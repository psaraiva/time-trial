package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Envelope wraps every JSON response in a standard envelope:
//
//	{
//	  "data":      { ...original response... },
//	  "duration":  123,           // ms from request received to response sent
//	  "timestamp": "2006-01-02T15:04:05Z07:00"  // RFC3339
//	}
func envelope() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		if err := c.Next(); err != nil {
			return err
		}

		if strings.HasPrefix(c.Path(), "/swag") {
			return nil
		}

		body := c.Response().Body()

		var data any
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}

		c.Response().ResetBody()

		return c.JSON(fiber.Map{
			"data":      data,
			"duration":  time.Since(start).Milliseconds(),
			"timestamp": start.UTC().Format(time.RFC3339),
		})
	}
}
