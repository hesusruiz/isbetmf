package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hesusruiz/isbetmf/proxy"
)

func main() {
	// Parse command line flags
	var (
		baseURL              = flag.String("base-url", "", "Base URL of the TMForum server")
		timeout              = flag.Int("timeout", 30, "Timeout in seconds for HTTP requests")
		objectTypes          = flag.String("object-types", "", "Comma-separated list of object types to validate")
		outputDir            = flag.String("output-dir", "./reports", "Output directory for reports")
		reportFile           = flag.String("report-file", "tmf_validation_report.md", "Name of the report file")
		paginationEnabled    = flag.Bool("pagination", true, "Enable pagination for object retrieval")
		pageSize             = flag.Int("page-size", 100, "Number of objects per page")
		maxObjects           = flag.Int("max-objects", 10000, "Maximum objects to retrieve per type")
		validateRequired     = flag.Bool("validate-required", true, "Validate required fields")
		validateRelatedParty = flag.Bool("validate-related-party", true, "Validate related party requirements")
		progress             = flag.Bool("progress", false, "Show progress updates")
		configFile           = flag.String("config", "", "Configuration file (JSON or YAML)")
		help                 = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	config := proxy.DefaultConfig()

	// Load from config file if specified
	if *configFile != "" {
		if err := loadConfigFromFile(config, *configFile); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
	}

	// Override with command line flags
	if *baseURL != "" {
		config.BaseURL = *baseURL
	}
	if *timeout > 0 {
		config.Timeout = *timeout
	}
	if *objectTypes != "" {
		config.ObjectTypes = parseObjectTypes(*objectTypes)
	}
	if *outputDir != "" {
		config.OutputDir = *outputDir
	}
	if *reportFile != "" {
		config.ReportFile = *reportFile
	}
	config.ValidateRequiredFields = *validateRequired
	config.ValidateRelatedParty = *validateRelatedParty
	config.PaginationEnabled = *paginationEnabled
	config.PageSize = *pageSize
	config.MaxObjects = *maxObjects

	// Load from environment variables
	config.LoadConfigFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create proxy instance
	proxyInstance, err := proxy.NewProxy(config)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	// Run validation process
	if *progress {
		if err := runWithProgress(ctx, proxyInstance); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
	} else {
		if err := proxyInstance.Run(ctx); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
	}

	log.Printf("Validation completed successfully")
}

// runWithProgress runs the validation process with progress reporting
func runWithProgress(ctx context.Context, proxyInstance *proxy.Proxy) error {
	progressChan := make(chan proxy.ProgressUpdate)

	// Start validation in goroutine
	go func() {
		if err := proxyInstance.RunWithProgress(ctx, progressChan); err != nil {
			log.Printf("Error: %v", err)
		}
	}()

	// Monitor progress
	for update := range progressChan {
		fmt.Printf("[%s] %s - %d%%\n",
			update.Stage, update.Message, update.Progress)

		if update.Stage == "Error" {
			return fmt.Errorf("validation failed: %s", update.Message)
		}

		if update.Stage == "Complete" {
			break
		}
	}

	return nil
}

// parseObjectTypes parses a comma-separated string of object types
func parseObjectTypes(typesStr string) []string {
	if typesStr == "" {
		return nil
	}

	types := strings.Split(typesStr, ",")
	for i, t := range types {
		types[i] = strings.TrimSpace(t)
	}
	return types
}

// loadConfigFromFile loads configuration from a file
func loadConfigFromFile(config *proxy.Config, filename string) error {
	// This is a placeholder - in a real implementation, you would
	// parse JSON or YAML files here
	return fmt.Errorf("config file loading not implemented yet")
}

// showHelp displays help information
func showHelp() {
	fmt.Printf(`TMForum Proxy Validator

Usage: tmfproxy [options]

Options:
  -base-url string
        Base URL of the TMForum server
  -timeout int
        Timeout in seconds for HTTP requests (default 30)
  -object-types string
        Comma-separated list of object types to validate
  -output-dir string
        Output directory for reports (default "./reports")
  -report-file string
        Name of the report file (default "tmf_validation_report.md")
  -pagination
        Enable pagination for object retrieval (default true)
  -page-size int
        Number of objects per page (default 100)
  -max-objects int
        Maximum objects to retrieve per type (default 10000)
  -validate-required
        Validate required fields (default true)
  -validate-related-party
        Validate related party requirements (default true)
  -progress
        Show progress updates
  -config string
        Configuration file (JSON or YAML)
  -help
        Show this help information

Environment Variables:
  TMF_BASE_URL          Base URL of the TMForum server
  TMF_TIMEOUT           Timeout in seconds
  TMF_OBJECT_TYPES      Comma-separated list of object types
  TMF_OUTPUT_DIR        Output directory
  TMF_REPORT_FILE       Report file name
  TMF_PAGINATION_ENABLED Enable pagination (true/false)
  TMF_PAGE_SIZE         Number of objects per page
  TMF_MAX_OBJECTS       Maximum objects to retrieve per type

Examples:
  tmfproxy -base-url "https://tmf.example.com" -object-types "productOffering,productSpecification"
  tmfproxy -config config.yaml
  TMF_BASE_URL="https://tmf.example.com" tmfproxy

`)
}
