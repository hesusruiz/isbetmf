package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hesusruiz/isbetmf/reporting"
)

func main() {
	fmt.Println("TMForum Reporting Example")
	fmt.Println("=====================")

	// Create configuration
	config := reporting.DefaultConfig()
	config.BaseURL = "https://dome-marketplace-sbx.org/"
	config.ObjectTypes = []string{"productOffering", "productSpecification"}
	config.OutputDir = "./example_reports"
	config.ReportFile = "example_report.md"

	// Load configuration from environment variables
	config.LoadConfigFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Base URL: %s\n", config.BaseURL)
	fmt.Printf("  Object Types: %v\n", config.ObjectTypes)
	fmt.Printf("  Timeout: %d seconds\n", config.Timeout)
	fmt.Printf("  Output Directory: %s\n", config.OutputDir)
	fmt.Printf("  Report File: %s\n", config.ReportFile)

	// Create proxy instance
	fmt.Println("\nCreating proxy instance...")
	proxyInstance, err := reporting.NewProxy(config)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run validation with progress tracking
	fmt.Println("\nStarting validation process...")
	progressChan := make(chan reporting.ProgressUpdate)

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
			log.Fatalf("Validation failed: %s", update.Message)
		}

		if update.Stage == "Complete" {
			break
		}
	}

	fmt.Println("\nValidation completed successfully!")
	fmt.Printf("Check the report at: %s/%s\n", config.OutputDir, config.ReportFile)
}
