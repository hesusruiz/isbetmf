package main

import (
	"context"
	"flag" // Added
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/cloudflare/tableflip"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/sqlogger"
	_ "github.com/hesusruiz/isbetmf/migrations"
	"github.com/hesusruiz/isbetmf/pdp"
	fiberhandler "github.com/hesusruiz/isbetmf/tmfserver/handler/fiber"
	repository "github.com/hesusruiz/isbetmf/tmfserver/repository"
	service "github.com/hesusruiz/isbetmf/tmfserver/service"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var debugFlag bool
	var environment string
	var restartHour, restartMinute int

	// Parse command-line flags
	flag.BoolVar(&debugFlag, "d", true, "Enable debug logging")
	flag.StringVar(&environment, "run", "isbedev", "Environment where run: isbedev, isbepre, domesbx, domedev2, domepro, lcl")
	flag.IntVar(&restartHour, "rh", 3, "Restart program every day at this hour")
	flag.IntVar(&restartMinute, "rm", 0, "Restart program every day at this minute")
	flag.Parse()

	// Generate a default configuration suitable for the environment
	// The approach is that instead of many configurable parameters, we have a set of profiles, with "hardcoded"
	// parameters for each environment, but that can be easity extended for other purposes.
	configuration, err := config.LoadConfig(environment, debugFlag)
	if err != nil {
		panic(err)
	}

	// Set restart schedule
	configuration.RestartHour = restartHour
	configuration.RestartMinute = restartMinute

	// Get the PID and name of our executable
	ourPid := os.Getpid()
	ourExecPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	// Exclude the name of the program from the list of arguments
	args := os.Args[1:]

	// ******************************************************
	// ******************************************************
	// The initial section is for when we are the init process in a container
	// Detect if we are running as PID=1 (most probably as init process in a container),
	// and act accordingly.
	runAsInit := false
	if ourPid == 1 {
		runAsInit = true
	} else {
		// For testing, 'init' must be passed as the first argument in the command line
		if len(args) > 0 && args[0] == "init" {
			runAsInit = true
			// Remove 'init' from the arguments, to prepare for executing our child processes
			args = args[1:]
		}
	}

	if runAsInit {
		runAsInitProcess(ourPid, ourExecPath, args)
		return
	}

	slog.Info("Process started", "PID", os.Getpid(), "executable", ourExecPath, "args", args)

	runNormalProcess(configuration)

}

func runNormalProcess(configuration *config.Config) {
	// We are now executing the normal process

	// Connect to the database
	db := sqlx.MustConnect("sqlite3", configuration.Dbname)
	defer db.Close()

	slog.Info("About to create tables if they do not exist")
	err := repository.CreateTables(db)
	if err != nil {
		slog.Error("failed to create tables", slog.Any("error", err))
		os.Exit(1)
	}

	// Create the PDP (aka Policy Decision Point or rules engine)
	rulesEngine, err := pdp.NewPDP(&pdp.Config{
		PolicyFileName: configuration.PolicyFileName,
		Debug:          configuration.Debug,
	})
	if err != nil {
		slog.Error("failed to create rules engine", slog.Any("error", err))
		os.Exit(1)
	}

	// Create the service
	s, err := service.NewTMFService(configuration, db, rulesEngine)
	if err != nil {
		slog.Error("failed to create service", slog.Any("error", err))
		os.Exit(1)
	}

	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		AppName:      "TMForum API Server",
		ServerHeader: "TMForum",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			// slog.Error("Fiber error", slog.Any("error", err), slog.Int("status", code))

			meth := fmt.Sprintf("<= %s %s", c.Method(), c.Path())
			slog.Error(meth, slog.Any("error", err), slog.Int("status", code), slog.String("ip", c.IP()))

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add middleware in order (order matters!)

	// 1. Recovery middleware - should be first to catch panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: configuration.Debug,
	}))

	// 2. Request ID middleware - for tracing requests
	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-Id",
		Generator: func() string {
			return "req_" + time.Now().Format("20060102150405") + "_" + os.Getenv("HOSTNAME")
		},
	}))

	// 3. CORS middleware - enable cross-origin requests
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Allow all origins as requested
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-Id",
		AllowCredentials: false, // Set to false when AllowOrigins is "*"
		ExposeHeaders:    "X-Request-Id",
		MaxAge:           86400, // 24 hours
	}))

	// 4. Compression middleware - compress responses
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// 5. Logger middleware - log requests and replies
	app.Use(sqlogger.FiberRequestLogger)

	// TODO: Implement request timeout

	// Serve the OpenAPI UI. We support V4 and V5
	app.Static("/oapiv5", "./www/oapiv5")
	app.Static("/oapiv4", "./www/oapiv4")

	// Create handler and set the routes for the APIs
	h := fiberhandler.NewHandler(s)
	h.RegisterRoutes(app)

	// TABLEFLIP for seamless restarts and upgrades
	upg, err := tableflip.New(tableflip.Options{
		PIDFile: "isbetmf.pid",
	})
	if err != nil {
		slog.Error("failed to create tableflip upgrader, exiting", slog.Any("error", err))
		os.Exit(1)
	}
	defer upg.Stop()

	// Listen for the process signal to trigger the tableflip upgrade.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			fmt.Println("Received SIGHUP, upgrading...")
			upg.Upgrade()
		}
	}()

	// Schedule restarts if enabled
	scheduleRestart(configuration, upg)

	// Listen must be called before signaling we are ready
	ln, err := upg.Listen("tcp", "0.0.0.0:9991")
	if err != nil {
		panic(errl.Error(err))
	}
	defer ln.Close()

	// Start the server in a separate goroutine
	go func() {
		slog.Info("TMF API server starting", slog.String("port", ":9991"))
		err := app.Listener(ln)
		if err != nil {
			slog.Error("Error starting TMF API server", "error", errl.Error(err))
			panic(err)
		}
	}()

	// tableflip ready
	slog.Info("Server is ready")
	if err := upg.Ready(); err != nil {
		panic(errl.Error(err))
	}

	<-upg.Exit()

	// Make sure to set a deadline on exiting the process
	// after upg.Exit() is closed. No new upgrades can be
	// performed if the parent doesn't exit.
	time.AfterFunc(30*time.Second, func() {
		log.Println("Graceful shutdown timed out")
		os.Exit(1)
	})

	// Wait for connections to drain.
	err = app.ShutdownWithContext(context.Background())
	if err != nil {
		fmt.Println("Exiting with error:", errl.Error(err).Error())
	} else {
		fmt.Println("Exiting without error")
	}

}

// scheduleRestart starts a goroutine to initiate an upgrade at the scheduled time every day.
// The restart uses the https://github.com/cloudflare/tableflip graceful restart to keeo client connections
func scheduleRestart(configuration *config.Config, upg *tableflip.Upgrader) {

	if configuration.RestartHour < 0 {
		return
	}

	// Schedule restarts/upgrades every night, the exact time does not matter
	targetHour := configuration.RestartHour
	targetMinute := configuration.RestartMinute
	targetSecond := 0

	go func() {
		for {
			now := time.Now()

			// Calculate the next scheduled time
			nextRun := time.Date(
				now.Year(), now.Month(), now.Day(),
				targetHour, targetMinute, targetSecond, 0, now.Location(),
			)

			// If the next run time is in the past, schedule it for the next day
			if nextRun.Before(now) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			fmt.Printf("Time now: %v\n", now)

			// Calculate the duration until the next run
			duration := time.Until(nextRun)
			fmt.Printf("Next restart/upgrade scheduled at: %v\n", nextRun)

			// Wait until the next run time
			time.Sleep(duration)

			// Execute the function
			fmt.Println("Executing scheduled restart...")
			upg.Upgrade()
		}
	}()

}

func runAsInitProcess(ourPid int, ourExecPath string, args []string) {
	// We are the init process in a container, or testing the functionality
	fmt.Println("We are the init process! PID:", ourPid)

	// Pass to child all arguments following the "init" entry (first argument after program name)
	cmd := exec.Command(ourExecPath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set ProcessGroupID for child process as init process. Both will be under same process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// We need notification of all relevant signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	done := make(chan bool, 1)

	// Forward all signals to the child in a goroutine
	go func() {
		for sig := range sigs {

			if sig == syscall.SIGHUP {
				fmt.Println("INIT: process received SIGHUP")
			} else {
				fmt.Printf("INIT: process received %v signal\n", sig)
			}

			if cmd.Process != nil {
				fmt.Printf("INIT: received %v signal for PID %v\n", sig, cmd.Process.Pid)
				_ = cmd.Process.Signal(sig)
			}

			if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				// We are done, and the init process will terminate
				fmt.Println("INIT: using DONE channel to terminate init process")
				done <- true
			}
		}
	}()

	// Enter in a goroutine an infinite loop reaping periodically the zombie children
	go func() {
		for {
			var ws syscall.WaitStatus
			pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
			if pid <= 0 || err != nil {
				time.Sleep(1 * time.Second)
			} else {
				fmt.Printf("INIT: reaped zombie child with PID: %v\n", pid)
			}
		}

	}()

	// Start the child (a fork of ourselves) without waiting for termination
	fmt.Println("INIT: starting child process")
	if err := cmd.Start(); err != nil {
		log.Fatalf("INIT: failed to start child process: %v", err)
	}

	fmt.Println("INIT: awaiting signal to terminate init process")
	<-done
	fmt.Println("INIT: exiting init process")
}
