package reporting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/hesusruiz/isbetmf/tmfserver/service"
)

// RemoteOrchestrator orchestrates the validation process
type RemoteOrchestrator struct {
	config    *Config
	client    *ClientPaging
	validator *Validator
	reporter  *Reporter
	pager     *service.ClientWithPaging
}

// NewRemoteOrchestrator creates a new proxy instance
func NewRemoteOrchestrator(config *Config) (*RemoteOrchestrator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	pagingConfig := service.DefaultPagingConfig()
	paging := service.NewClientWithPaging(pagingConfig)

	return &RemoteOrchestrator{
		config:    config,
		client:    NewClientPaging(config),
		validator: NewValidator(config),
		reporter:  NewReporter(config),
		pager:     paging,
	}, nil
}

// Run performs the complete validation process
func (p *RemoteOrchestrator) Run(ctx context.Context) error {
	log.Printf("Starting TMForum object validation process")
	log.Printf("Base URL: %s", p.config.BaseURL)
	log.Printf("Object types: %v", p.config.ObjectTypes)

	// Process each object type
	var allResults []repository.ValidationResult
	// var allObjects []TMFObject
	var allObjects []repository.TMFObjectMap

	for _, objectType := range p.config.ObjectTypes {
		log.Printf("Processing object type: %s", objectType)

		// Process objects page by page to avoid loading all into memory
		offset := 0
		totalObjectsForType := 0
		totalValidForType := 0
		totalErrorsForType := 0
		totalWarningsForType := 0

		for {

			// Get a single page of objects
			objects, err := p.pager.GetPageOfObjects(ctx, objectType, offset, nil)

			// objects, err := p.client.GetObjectsPage(ctx, objectType, p.config, offset)
			if err != nil {
				log.Printf("Warning: Failed to retrieve %s objects at offset %d: %v", objectType, offset, err)
				break
			}

			// If no objects returned, we've reached the end
			if len(objects) == 0 {
				break
			}

			log.Printf("Retrieved %d %s objects (offset %d)", len(objects), objectType, offset)
			allObjects = append(allObjects, objects...)
			totalObjectsForType += len(objects)

			// Validate objects from this page
			results := p.validator.ValidateObjects(objects, objectType)
			allResults = append(allResults, results...)

			// Count validation results for this page
			for _, result := range results {

				if result.Valid {
					totalValidForType++
				}
				totalErrorsForType += len(result.Errors)
				totalWarningsForType += len(result.Warnings)
			}

			// If we got fewer objects than the page size, we've reached the end
			if len(objects) < p.config.PageSize {
				break
			}

			// Move to next page
			offset += p.config.PageSize

			// Safety check to prevent infinite loops
			if offset >= p.config.MaxObjects {
				log.Printf("Warning: Reached maximum objects limit (%d) for %s", p.config.MaxObjects, objectType)
				break
			}
		}

		for i, result := range allResults {
			// Set the URL of the object
			url, err := p.pager.GetObjectURL(result.ObjectType, result.ObjectID)
			if err != nil {
				log.Printf("Warning: Failed to get URL for object %s: %v", result.ObjectID, err)
			}
			allResults[i].Href = url
		}

		log.Printf("Validation complete for %s: %d total objects, %d valid, %d errors, %d warnings",
			objectType, totalObjectsForType, totalValidForType, totalErrorsForType, totalWarningsForType)
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
func (p *RemoteOrchestrator) RunWithProgress(ctx context.Context, progressChan chan<- ProgressUpdate) error {
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
	var allResults []repository.ValidationResult
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

		// Process objects page by page to avoid loading all into memory
		offset := 0
		totalObjectsForType := 0

		for {
			// Get a single page of objects
			objects, err := p.pager.GetPageOfObjects(ctx, objectType, offset, nil)

			// objects, err := p.client.GetObjectsPage(ctx, objectType, p.config, offset)
			if err != nil {
				log.Printf("Warning: Failed to retrieve %s objects at offset %d: %v", objectType, offset, err)
				break
			}

			// If no objects returned, we've reached the end
			if len(objects) == 0 {
				break
			}

			totalObjectsForType += len(objects)

			// Validate objects from this page
			results := p.validator.ValidateObjects(objects, objectType)
			allResults = append(allResults, results...)

			// If we got fewer objects than the page size, we've reached the end
			if len(objects) < p.config.PageSize {
				break
			}

			// Move to next page
			offset += p.config.PageSize

			// Safety check to prevent infinite loops
			if offset >= p.config.MaxObjects {
				log.Printf("Warning: Reached maximum objects limit (%d) for %s", p.config.MaxObjects, objectType)
				break
			}
		}

		progressChan <- ProgressUpdate{
			Stage:     "Processing",
			Message:   fmt.Sprintf("Completed %s: %d objects processed", objectType, totalObjectsForType),
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
func (p *RemoteOrchestrator) GetStatistics() *Statistics {
	return &Statistics{
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
}
