//	@title			time-trial API
//	@version		1.0
//	@description	Minimal service for injecting controlled HTTP failures (status codes, delays, dynamic response bodies) into dependent services during testing.
//	@contact.name	time-trial contributors
//	@host			localhost:7777
//	@BasePath		/

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/psaraiva/time-trial/docs/swagger"
	"github.com/psaraiva/time-trial/internal/entities"
	"github.com/psaraiva/time-trial/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func buildApp() *fiber.App {
	eState := entities.NewState()
	ePlan := entities.NewPlan()
	eParamResp := entities.NewParamResp()

	hSabotage := handlers.NewSabotageHandler(eState)
	hPlan := handlers.NewPlanHandler(ePlan)
	hExec := handlers.NewExecHandler(eState, ePlan, eParamResp)
	hConfig := handlers.NewConfigHandler(eState)
	hParamResp := handlers.NewParamRespHandler(eParamResp)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(envelope())

	app.Get("/swag/*", fiberSwagger.WrapHandler)

	app.Post("/time-trial", hSabotage.SetSabotage)
	app.Get("/time-trial/config", hConfig.GetConfig)
	app.Post("/plan", hPlan.SetPlan)
	app.Get("/plan/config", hPlan.GetConfig)
	app.Get("/sabotage", hExec.Exec)
	app.Get("/plan/sabotage", hExec.ExecPlan)
	app.Post("/param-resp", hParamResp.SetParamResp)
	app.Get("/param-resp/config", hParamResp.GetConfig)

	return app
}

func main() {
	host := os.Getenv("TIME_TRIAL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TIME_TRIAL_API_PORT")
	if port == "" {
		port = "7777" // luck 7 ♠♥♦♣
	}

	swagger.SwaggerInfo.Host = host + ":" + port

	pprofPort := os.Getenv("TIME_TRIAL_PPROF_PORT")
	if pprofPort == "" {
		pprofPort = "6060"
	}

	go func() {
		log.Printf("pprof listening on :%s", pprofPort)
		if err := http.ListenAndServe(":"+pprofPort, nil); err != nil {
			log.Printf("pprof error: %v", err)
		}
	}()

	log.Printf("Server listening on :%s", port)
	if err := buildApp().Listen(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
