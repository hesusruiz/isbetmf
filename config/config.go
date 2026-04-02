package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/sqlogger"
	"github.com/hesusruiz/isbetmf/types"
)

// Indicates the environment (SBX, DEV2, PRO, LCL) where the server is running.
// It is used to determine the default configuration profile if the user does not specify anything else.
type Environment string

const (
	DOME_PRO Environment = "domepro"
	DOME_PRE Environment = "domepre"
	DOME_DEV Environment = "domedev"
	DOME_LCL Environment = "domelcl"
	ISBE_PRE Environment = "isbepre"
	ISBE_DEV Environment = "isbedev"
)

const DefaultClonePeriod = 10 * time.Minute

type Config struct {

	// The environment for the default configuration profile
	Environment Environment

	// Server operator information
	ServerOperatorOrganizationIdentifier string
	ServerOperatorDid                    string
	ServerOperatorName                   string
	ServerOperatorCountry                string
	ServerEmailAddress                   string

	// VerifierServer is the URL of the verifier server, which is used to verify the access tokens.
	VerifierServer string

	// The domain of the remote TMForum API server when we act as proxy
	RemoteTMFServer string

	// ProxyEnabled enables the TMF caching proxy functionality.
	ProxyEnabled bool

	// Dbname is the name of the database file where the local TMForum data is stored
	// It is used to store the data in a local SQLite database, the best SQL database for this purpose.
	Dbname         string
	BackupDisabled bool

	// The power required by a caller to be considered LEAR
	LEARPower types.OnePower

	// The powers required by a caller to be able to create, update and delete a product
	ProductCreatePower types.OnePower
	ProductUpdatePower types.OnePower
	ProductDeletePower types.OnePower

	// PolicyFileName is the name of the file where the policies are stored.
	// It can specify a local file or a remote URL.
	PolicyFileName string

	// PDPAddress is the address of the PDP server.
	PDPAddress string

	// Debug mode, more logs and less caching
	Debug bool

	// TODO: this is temporary for testing
	FakeClaims bool

	// The admin token used to authenticate the superadmin
	AdminToken string

	// Enable synchronization with the remote server in background
	BackgroudSync bool

	// ClonePeriod is the period in which the reporting tool will clone the TMForum objects from the DOME instance,
	// to keep the local cache up to date.
	ClonePeriod time.Duration

	// LogHandler is the handler used to log messages.
	// It is a custom handler that uses the slog package to log messages both to the console and to a SQLite database.
	LogHandler *sqlogger.SQLogHandler

	// LogLevel is a slog.LevelVar that can be set to different log levels (e.g. Debug, Info, Warn, Error).
	LogLevel *slog.LevelVar

	// Hour and minute of the day when the server will automatically restart (each day). Hour=-1 disables restart.
	RestartHour, RestartMinute int

	// The features of the environment
	Features Features
}

// Features defines a set of feature flags which may depend on the environment at a given time
type Features struct {
	OfferingLaunchOnlyByAdmin bool
	GenerateIDOnCreate        bool
	AllowIDInBody             bool
}

// LoadConfig initializes and returns a Config struct based on the provided parameters.
// It sets up logging, selects the appropriate environment, and applies configuration options.
//
// Parameters:
//   - envir:        The environment to use ("pro", "dev2", "sbx", "lcl").
//   - debug:        Enables debug logging if true.
//
// Returns:
//   - *Config: The initialized configuration struct.
//   - error:   An error if configuration or logger setup fails.
func LoadConfig(
	envir string,
	debug bool,
) (*Config, error) {
	var conf *Config

	// Normalize to lowercase for comparisons
	envir = strings.ToLower(envir)

	// The environment has precedence over the parameter
	if en := os.Getenv("ISBETMF_RUN_ENVIRONMENT"); en != "" {
		envir = strings.ToLower(en)
	}
	environment := Environment(envir)

	// Get the admin token from the environment variable ISBETMF_ADMIN_TOKEN
	adminToken := os.Getenv("ISBETMF_ADMIN_TOKEN")
	if adminToken == "" {
		// For testing, use the testing token
		adminToken = "eyJhdWQiOiJodHRwczovL2NhdGFsb2cuaX"
	}

	// Configure the slog logger
	var logLevel slog.Level
	if debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	// Initialize the custom SQLogHandler
	logOptions := &sqlogger.Options{
		Level:  &logLevel,
		LogDir: "data/logs",
	}

	// Check if the logs should be colored:
	// - If the process is running in a container (pid=1) then do not color the logs
	// - If the environment variable ISBETMF_LOGS_NOCOLOR is set to "true" then do not color the logs
	ourpid := os.Getpid()
	if ourpid == 1 || os.Getenv("ISBETMF_LOGS_NOCOLOR") == "true" {
		logOptions.NoColor = true
	}

	// Initialize the logging system
	sqlog, err := sqlogger.NewSQLogHandler(logOptions)
	if err != nil {
		slog.Error("failed to initialize SQLogHandler, exiting", slog.Any("error", err))
		os.Exit(1)
	}

	// And set the default logging system for all components
	slog.SetDefault(slog.New(sqlog))

	// Choose the profile from the environment passed
	switch environment {
	case DOME_PRO:
		conf = domeproConfig
		slog.Info("Using the DOME PRO environment")
	case DOME_PRE:
		conf = domepreConfig
		slog.Info("Using the DOME PRE environment")
	case DOME_DEV:
		conf = domedevConfig
		slog.Info("Using the DOME SBX environment")
	case DOME_LCL:
		conf = lclConfig
		slog.Info("Using the LCL environment")
	case ISBE_PRE:
		conf = isbepreConfig
		slog.Info("Using the ISBE PRE environment")
	case ISBE_DEV:
		conf = isbedevConfig
		slog.Info("Using the ISBE DEV environment")
	default:
		conf = isbedevConfig
		slog.Info("Using the default (ISBE DEV) environment")
	}

	conf.Debug = debug
	conf.LogHandler = sqlog

	// Check for overrides with environment variables

	verifierServer := os.Getenv("ISBETMF_VERIFIER")
	if verifierServer != "" {
		conf.VerifierServer = verifierServer
	}
	slog.Info("Verifier", slog.String("url", conf.VerifierServer))

	remoteTMFServer := os.Getenv("ISBETMF_REMOTE_SERVER")
	if remoteTMFServer != "" {
		conf.RemoteTMFServer = remoteTMFServer
	}
	slog.Info("RemoteTMFServer", slog.String("url", conf.RemoteTMFServer))

	proxyEnabled := os.Getenv("ISBETMF_PROXY_ENABLED")
	switch proxyEnabled {
	case "true":
		conf.ProxyEnabled = true
	case "false":
		conf.ProxyEnabled = false
	}
	slog.Info("Proxy", slog.Bool("enabled", conf.ProxyEnabled))

	return conf, nil

}

func (c *Config) IsDOME() bool {
	return c.Environment == DOME_PRO || c.Environment == DOME_PRE || c.Environment == DOME_DEV || c.Environment == DOME_LCL
}

func (c *Config) IsISBE() bool {
	return c.Environment == ISBE_PRE || c.Environment == ISBE_DEV
}

func (c *Config) Close() {
	if c.LogHandler != nil {
		c.LogHandler.Close()
	}
}

// The names of some special objects in the DOME ecosystem
const ProductOffering = "productOffering"
const ProductSpecification = "productSpecification"
const ProductOfferingPrice = "productOfferingPrice"
const ServiceSpecification = "serviceSpecification"
const ResourceSpecification = "resourceSpecification"
const Category = "category"
const Catalog = "catalog"
const Organization = "organization"
const Individual = "individual"
