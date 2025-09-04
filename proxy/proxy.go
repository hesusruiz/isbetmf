package proxy

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Proxy orchestrates the validation process
type Proxy struct {
	config    *Config
	client    *Client
	validator *Validator
	reporter  *Reporter
}

// NewProxy creates a new proxy instance
func NewProxy(config *Config) (*Proxy, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Proxy{
		config:    config,
		client:    NewClient(config),
		validator: NewValidator(config),
		reporter:  NewReporter(config),
	}, nil
}

// Run performs the complete validation process
func (p *Proxy) Run(ctx context.Context) error {
	log.Printf("Starting TMForum object validation process")
	log.Printf("Base URL: %s", p.config.BaseURL)
	log.Printf("Object types: %v", p.config.ObjectTypes)

	// // Test connection to remote server
	// log.Printf("Testing connection to remote server...")
	// if err := p.client.TestConnection(ctx); err != nil {
	// 	return fmt.Errorf("connection test failed: %w", err)
	// }
	// log.Printf("Connection successful")

	// // Get server information
	// log.Printf("Retrieving server information...")
	// if _, err := p.client.GetServerInfo(ctx); err != nil {
	// 	log.Printf("Warning: Failed to get server info: %v", err)
	// } else {
	// 	log.Printf("Server info retrieved successfully")
	// }

	// Process each object type
	var allResults []ValidationResult
	var allObjects []TMFObject

	for _, objectType := range p.config.ObjectTypes {
		log.Printf("Processing object type: %s", objectType)

		objects, err := p.client.GetObjects(ctx, objectType, p.config)
		if err != nil {
			log.Printf("Warning: Failed to retrieve %s objects: %v", objectType, err)
			continue
		}

		log.Printf("Retrieved %d %s objects", len(objects), objectType)
		allObjects = append(allObjects, objects...)

		// Validate objects
		results := p.validator.ValidateObjects(objects, objectType)
		allResults = append(allResults, results...)

		// Log validation summary
		validCount := 0
		errorCount := 0
		warningCount := 0

		for _, result := range results {
			if result.Valid {
				validCount++
			}
			errorCount += len(result.Errors)
			warningCount += len(result.Warnings)
		}

		log.Printf("Validation complete for %s: %d valid, %d errors, %d warnings",
			objectType, validCount, errorCount, warningCount)
	}

	// Generate report
	log.Printf("Generating validation report...")
	report, err := p.reporter.GenerateReport(allResults)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Log final summary
	log.Printf("Validation process complete")
	log.Printf("Total objects processed: %d", report.Statistics.TotalObjects)
	log.Printf("Valid objects: %d", report.Statistics.ValidObjects)
	log.Printf("Invalid objects: %d", report.Statistics.InvalidObjects)
	log.Printf("Total errors: %d", report.Statistics.TotalErrors)
	log.Printf("Total warnings: %d", report.Statistics.TotalWarnings)
	log.Printf("Processing time: %v", report.Statistics.Duration)
	log.Printf("Report saved to: %s", p.config.ReportFile)

	return nil
}

// RunWithProgress runs the validation process with progress reporting
func (p *Proxy) RunWithProgress(ctx context.Context, progressChan chan<- ProgressUpdate) error {
	defer close(progressChan)

	progressChan <- ProgressUpdate{
		Stage:     "Starting",
		Message:   "Initializing validation process",
		Progress:  0,
		Timestamp: time.Now(),
	}

	// Test connection
	progressChan <- ProgressUpdate{
		Stage:     "Connection",
		Message:   "Testing connection to remote server",
		Progress:  10,
		Timestamp: time.Now(),
	}

	if err := p.client.TestConnection(ctx); err != nil {
		progressChan <- ProgressUpdate{
			Stage:     "Error",
			Message:   fmt.Sprintf("Connection failed: %v", err),
			Progress:  0,
			Timestamp: time.Now(),
		}
		return fmt.Errorf("connection test failed: %w", err)
	}

	progressChan <- ProgressUpdate{
		Stage:     "Connection",
		Message:   "Connection successful",
		Progress:  20,
		Timestamp: time.Now(),
	}

	// Process object types
	var allResults []ValidationResult
	totalObjectTypes := len(p.config.ObjectTypes)

	for i, objectType := range p.config.ObjectTypes {
		progress := 20 + (i * 60 / totalObjectTypes)

		progressChan <- ProgressUpdate{
			Stage:     "Processing",
			Message:   fmt.Sprintf("Processing %s objects", objectType),
			Progress:  progress,
			Timestamp: time.Now(),
		}

		fmt.Printf("################# Processing %s objects\n", objectType)

		objects, err := p.client.GetObjects(ctx, objectType, p.config)
		if err != nil {
			log.Printf("Warning: Failed to retrieve %s objects: %v", objectType, err)
			continue
		}

		// Validate objects
		results := p.validator.ValidateObjects(objects, objectType)
		allResults = append(allResults, results...)

		progressChan <- ProgressUpdate{
			Stage:     "Processing",
			Message:   fmt.Sprintf("Completed %s: %d objects processed", objectType, len(objects)),
			Progress:  progress + (60 / totalObjectTypes),
			Timestamp: time.Now(),
		}
	}

	// Generate report
	progressChan <- ProgressUpdate{
		Stage:     "Reporting",
		Message:   "Generating validation report",
		Progress:  90,
		Timestamp: time.Now(),
	}

	report, err := p.reporter.GenerateReport(allResults)
	if err != nil {
		progressChan <- ProgressUpdate{
			Stage:     "Error",
			Message:   fmt.Sprintf("Report generation failed: %v", err),
			Progress:  0,
			Timestamp: time.Now(),
		}
		return fmt.Errorf("failed to generate report: %w", err)
	}

	progressChan <- ProgressUpdate{
		Stage: "Complete",
		Message: fmt.Sprintf("Validation complete: %d objects, %d errors, %d warnings",
			report.Statistics.TotalObjects, report.Statistics.TotalErrors, report.Statistics.TotalWarnings),
		Progress:  100,
		Timestamp: time.Now(),
	}

	return nil
}

// ProgressUpdate represents a progress update during the validation process
type ProgressUpdate struct {
	Stage     string    `json:"stage"`
	Message   string    `json:"message"`
	Progress  int       `json:"progress"` // 0-100
	Timestamp time.Time `json:"timestamp"`
}

// GetStatistics returns the current validation statistics
func (p *Proxy) GetStatistics() *Statistics {
	return &Statistics{
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
}
