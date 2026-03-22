package main

import (
	"log"
	"os"

	"github.com/psaraiva/time-trial/internal/entities"
	"github.com/psaraiva/time-trial/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func buildApp() *fiber.App {
	eState := entities.NewState()
	ePlan := entities.NewPlan()

	hSabotage := handlers.NewSabotageHandler(eState)
	hPlan := handlers.NewPlanHandler(ePlan)
	hExec := handlers.NewExecHandler(eState, ePlan)
	hConfig := handlers.NewConfigHandler(eState)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(envelope())

	app.Post("/sabotage", hSabotage.SetSabotage)
	app.Post("/plan/sabotage", hPlan.SetPlan)
	app.Get("/plan/config", hPlan.GetConfig)
	app.Get("/sabotage/exec", hExec.Exec)
	app.Get("/plan/exec", hExec.ExecPlan)
	app.Get("/sabotage/config", hConfig.GetConfig)

	return app
}

func main() {
	port := os.Getenv("TIME_TRIAL_API_PORT")
	if port == "" {
		port = "7777" // luck 7 ♠♥♦♣
	}

	log.Printf("Server listening on :%s", port)
	if err := buildApp().Listen(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
