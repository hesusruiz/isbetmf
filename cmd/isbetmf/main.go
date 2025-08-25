package main

import (
	"flag" // Added
	"os"

	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/hesusruiz/isbetmf/notification"
	"github.com/hesusruiz/isbetmf/pdp"
	fiberhandler "github.com/hesusruiz/isbetmf/tmfserver/handler/fiber"
	repository "github.com/hesusruiz/isbetmf/tmfserver/repository"
	service "github.com/hesusruiz/isbetmf/tmfserver/service"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/greyxor/slogor"
)

func main() {
	// Configure slog logger
	var debugFlag bool
	var verifierServer string
	flag.BoolVar(&debugFlag, "d", false, "Enable debug logging")
	flag.StringVar(&verifierServer, "verifier", "https://verifier.dome-marketplace.eu", "Full URL of the verifier which signs access tokens")
	flag.Parse()

	var logLevel slog.Level
	if debugFlag {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	handler := slogor.NewHandler(os.Stdout, slogor.ShowSource(), slogor.SetLevel(logLevel))
	slog.SetDefault(slog.New(handler))

	// Connect to the database
	db, err := sqlx.Connect("sqlite3", "isbetmf.db")
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec(repository.CreateTMFTableSQL)
	if err != nil {
		slog.Error("failed to create table", slog.Any("error", err))
		os.Exit(1)
	}

	// Create the PDP (aka rules engine)
	rulesEngine, err := pdp.NewPDP(&pdp.Config{
		PolicyFileName: "auth_policies.star",
		VerifierServer: verifierServer,
		Debug:          debugFlag,
	})
	if err != nil {
		slog.Error("failed to create rules engine", slog.Any("error", err))
		os.Exit(1)
	}

	// Create the service
	s := service.NewService(db, rulesEngine, verifierServer)

	app := fiber.New()

	// Serve the OpenAPI UI
	app.Static("/oapi", "./www/oapiui")

	// Create handler and set the routes for the APIs
	h := fiberhandler.NewHandler(s)
	h.RegisterRoutes(app)

	// Create and register the hub handler
	// Create and register the hub handler
	hubHandler := notification.NewHubHandler(s.HubManager)
	app.Post("/hub", hubHandler.Subscribe)
	app.Delete("/hub/:id", hubHandler.Unsubscribe)

	// And start the server
	slog.Info("TMF API server starting", slog.String("port", ":9991"))
	app.Listen("0.0.0.0:9991")

}
