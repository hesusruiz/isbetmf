package main

import (
	"flag" // Added
	"fmt"
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
		slog.Error("Failed to load configuration", slog.Any("error", err))
		panic(err)
	}
	defer configuration.Close()
	slog.Info("Configuration loaded", slog.Any("config", configuration))

	for i := range 100000 {
		slog.Info("Looping" + fmt.Sprintf("%d", i))
	}

	// Set restart schedule
	configuration.RestartHour = restartHour
	configuration.RestartMinute = restartMinute

	// Get the PID and name of our executable
	ourPid := os.Getpid()
	ourExecPath, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path", slog.Any("error", err))
		panic(err)
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
		slog.Info("We are the INIT process!", "PID", ourPid, "executable", ourExecPath, "args", args)
		runAsInitProcess(ourExecPath, args)
	} else {
		slog.Info("TMF API server starting", "PID", ourPid, "executable", ourExecPath, "args", args)
		runNormalProcess(configuration)
	}

}

func cleanup(db *sqlx.DB) {
	// This deferred function will run!
	fmt.Println("Running deferred cleanup functions...")

	// Close database connection (triggers WAL cleanup)
	fmt.Println("Closing database connection and exiting...")
	db.Close()

	fmt.Println("Database connections closed.")
}

// runNormalProcess starts the TMF API server and handles its lifecycle,
// including database connection, rules engine initialization, and graceful shutdown.
func runNormalProcess(configuration *config.Config) {

	// TABLEFLIP for seamless restarts and upgrades
	upg, err := tableflip.New(tableflip.Options{
		PIDFile: "isbetmf.pid",
	})
	if err != nil {
		slog.Error("failed to create tableflip upgrader, exiting", slog.Any("error", err))
		panic(err)
	}

	// Connect to the database and create tables if they do not exist
	db, err := repository.NewDBService(configuration)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		return
	}
	defer cleanup(db)

	// Schedule maintenance tasks every 2 hours
	repository.ScheduleMaintenance(db, configuration.Dbname, 2*time.Hour)

	// Create the PDP (aka Policy Decision Point or rules engine)
	rulesEngine, err := pdp.NewPDPService(&pdp.Config{
		PolicyFileName: configuration.PolicyFileName,
		Debug:          configuration.Debug,
	})
	if err != nil {
		slog.Error("failed to create rules engine", slog.Any("error", err))
		return
	}

	// Create the service, which will use the database and the rules engine
	s, err := service.NewTMFService(configuration, db, rulesEngine)
	if err != nil {
		slog.Error("failed to create service", slog.Any("error", err))
		return
	}

	// Create Fiber web server with custom configuration
	webServer := fiber.New(fiber.Config{
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
	webServer.Use(recover.New(recover.Config{
		EnableStackTrace: configuration.Debug,
	}))

	// 2. Request ID middleware - for tracing requests
	webServer.Use(requestid.New(requestid.Config{
		Header: "X-Request-Id",
		Generator: func() string {
			return "req_" + time.Now().Format("20060102150405") + "_" + os.Getenv("HOSTNAME")
		},
	}))

	// 3. CORS middleware - enable cross-origin requests
	webServer.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Allow all origins as requested
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-Id",
		AllowCredentials: false, // Set to false when AllowOrigins is "*"
		ExposeHeaders:    "X-Request-Id",
		MaxAge:           86400, // 24 hours
	}))

	// 4. Compression middleware - compress responses
	webServer.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// 5. Logger middleware - log requests and replies
	webServer.Use(sqlogger.FiberRequestLogger)

	// TODO: Implement request timeout

	// Serve the OpenAPI UI. We support V4 and V5
	webServer.Static("/oapiv5", "./www/oapiv5")
	webServer.Static("/oapiv4", "./www/oapiv4")

	// Create handler and set the routes for the APIs
	h := fiberhandler.NewHandler(s)
	h.RegisterRoutes(webServer)

	// Schedule restarts if enabled
	scheduleRestart(configuration, db, upg)

	// Listen must be called before signaling we are ready
	ln, err := upg.Listen("tcp", "0.0.0.0:9991")
	if err != nil {
		slog.Error("failed to listen on port 9991, exiting", slog.Any("error", err))
		panic(err)
	}
	defer ln.Close()

	// Start the server in a separate goroutine
	go func() {
		slog.Info("TMF API server starting", slog.String("port", ":9991"))
		err := webServer.Listener(ln)
		if err != nil {
			slog.Error("Error starting TMF API server", "error", errl.Error(err))
			panic(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sigs {
			// Process each type of signal
			switch sig {

			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("CHILD: Received SIGINT or SIGTERM, exiting...")

				// Close listeners and inherited FDs
				// Call upg.Stop() to shut down listeners immediately
				upg.Stop()

			case syscall.SIGHUP:
				// Perform a Tableflip upgrade
				fmt.Println("CHILD: Received SIGHUP, upgrading...")
				upg.Upgrade()
			}
		}
	}()

	// Signal that we are ready so Tableflip can stop the parent process
	slog.Info("CHILD: Server is ready")
	if err := upg.Ready(); err != nil {
		panic(errl.Error(err))
	}

	// Wait until we are told to exit by the Tableflip mechanism.
	// This happens when the child process has signalled that it is ready.
	fmt.Println("CHILD: Waiting for Tableflip to exit...")
	<-upg.Exit()
	fmt.Println("CHILD: Tableflip exit received")

	// Wait for connections to drain for a maximum of 30 seconds
	fmt.Println("CHILD: Waiting for connections to drain...")
	err = webServer.ShutdownWithTimeout(30 * time.Second)
	if err != nil {
		fmt.Println("CHILD: Exiting with error", errl.Error(err))
		os.Exit(1)
	} else {
		fmt.Println("CHILD: Exiting without error")
	}

}

// scheduleRestart starts a goroutine to initiate an upgrade at the scheduled time every day.
// The restart uses the https://github.com/cloudflare/tableflip graceful restart to keep client connections
func scheduleRestart(configuration *config.Config, db *sqlx.DB, upg *tableflip.Upgrader) {

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

			// Perform the maintenance tasks before restarting (VACUUM and Backup)
			repository.PerformMaintenance(db, configuration.Dbname)

			// Execute the function
			fmt.Println("Executing scheduled restart...")
			upg.Upgrade()
		}
	}()

}

// runAsInitProcess acts as an init process, launching a child process and forwarding
// signals to it. This is typically used in container environments where
// `main` might be PID 1.
//
// It sets up the child process to share stdout/stderr, places it in the same
// process group, and captures system signals (SIGINT, SIGTERM, SIGHUP) to
// gracefully relay them to the child.
//
//   - ourExecPath: The path to the executable to run as the child process.
//   - args: Command-line arguments to pass to the child process.
func runAsInitProcess(ourExecPath string, args []string) {

	// Pass to child all arguments following the "init" entry (first argument after program name)
	cmd := exec.Command(ourExecPath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set ProcessGroupID for child process as init process. Both will be under same process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Start the child (a fork of ourselves) without waiting for termination
	slog.Info("INIT: starting child process")
	if err := cmd.Start(); err != nil {
		slog.Error("INIT: failed to start child process", "error", err)
		return
	}

	// We need notification of all relevant signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	done := make(chan bool, 1)

	// Forward all signals to the child in a goroutine
	go func() {
		for sig := range sigs {

			// Forward the signal to the child process
			slog.Info("INIT: forwarding signal to child process", "signal", sig, "PID", cmd.Process.Pid)
			err := cmd.Process.Signal(sig)
			if err != nil {
				slog.Error("INIT: failed to forward signal to child process", "signal", sig, "PID", cmd.Process.Pid, "error", err)
			}

			// If we receive a SIGTERM or SIGINT, wait 5 seconds for the child to terminate and send a KILL signal
			if sig == syscall.SIGTERM || sig == syscall.SIGINT {

				go func() {
					// Wait 10 seconds for the child process to finish
					time.Sleep(10 * time.Second)
					// Kill the child immediately
					cmd.Process.Kill()
				}()

				slog.Info("INIT: using DONE channel to terminate init process")
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
				time.Sleep(5 * time.Second)
			} else {
				slog.Info("INIT: reaped zombie child with PID", "PID", pid)
			}
		}

	}()

	slog.Info("INIT: awaiting signal to terminate init process")
	<-done

	// Wait for the child process to finish and release its resources
	slog.Info("INIT: waiting for child process to finish")
	cmd.Process.Wait()

	slog.Info("INIT: exiting init process")
}
