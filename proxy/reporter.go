package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Reporter generates validation reports and statistics
type Reporter struct {
	config *Config
}

// NewReporter creates a new reporter
func NewReporter(config *Config) *Reporter {
	return &Reporter{
		config: config,
	}
}

// GenerateReport generates a complete validation report
func (r *Reporter) GenerateReport(results []ValidationResult) (*ValidationReport, error) {
	startTime := time.Now()

	// Calculate statistics
	stats := r.calculateStatistics(results)
	stats.StartTime = startTime
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	report := &ValidationReport{
		Config:      r.config,
		Statistics:  stats,
		Results:     results,
		GeneratedAt: time.Now(),
	}

	// Generate Markdown report
	if err := r.generateMarkdownReport(report); err != nil {
		return nil, fmt.Errorf("failed to generate markdown report: %w", err)
	}

	return report, nil
}

// calculateStatistics calculates statistics from validation results
func (r *Reporter) calculateStatistics(results []ValidationResult) *Statistics {
	stats := &Statistics{
		ObjectsByType:  make(map[string]TypeStats),
		ErrorsByType:   make(map[string]int),
		WarningsByType: make(map[string]int),
	}

	for _, result := range results {
		stats.TotalObjects++

		if result.Valid {
			stats.ValidObjects++
		} else {
			stats.InvalidObjects++
		}

		// Count by type
		typeStats := stats.ObjectsByType[result.ObjectType]
		typeStats.Count++
		if result.Valid {
			typeStats.Valid++
		} else {
			typeStats.Invalid++
		}
		typeStats.Errors += len(result.Errors)
		typeStats.Warnings += len(result.Warnings)
		stats.ObjectsByType[result.ObjectType] = typeStats

		// Count errors and warnings
		stats.TotalErrors += len(result.Errors)
		stats.TotalWarnings += len(result.Warnings)

		// Count errors by type
		for _, err := range result.Errors {
			stats.ErrorsByType[err.Code]++
		}

		// Count warnings by type
		for _, warning := range result.Warnings {
			stats.WarningsByType[warning.Code]++
		}
	}

	return stats
}

// generateMarkdownReport generates a Markdown report file
func (r *Reporter) generateMarkdownReport(report *ValidationReport) error {
	// Ensure output directory exists
	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filePath := filepath.Join(r.config.OutputDir, r.config.ReportFile)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write report header
	r.writeReportHeader(file, report)

	// Write summary statistics
	r.writeSummaryStatistics(file, report.Statistics)

	// Write detailed statistics by object type
	r.writeDetailedStatistics(file, report.Statistics)

	// Write error and warning summary
	r.writeErrorWarningSummary(file, report.Statistics)

	// Write detailed validation results
	r.writeDetailedResults(file, report.Results)

	// Write report footer
	r.writeReportFooter(file, report)

	return nil
}

// writeReportHeader writes the report header
func (r *Reporter) writeReportHeader(file *os.File, report *ValidationReport) {
	fmt.Fprintf(file, "# TMForum Object Validation Report\n\n")
	fmt.Fprintf(file, "**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(file, "**Configuration:**\n")
	fmt.Fprintf(file, "- Base URL: `%s`\n", report.Config.BaseURL)
	fmt.Fprintf(file, "- Object Types: %s\n", strings.Join(report.Config.ObjectTypes, ", "))
	fmt.Fprintf(file, "- Timeout: %d seconds\n", report.Config.Timeout)
	fmt.Fprintf(file, "- Validate Required Fields: %t\n", report.Config.ValidateRequiredFields)
	fmt.Fprintf(file, "- Validate Related Party: %t\n\n", report.Config.ValidateRelatedParty)
}

// writeSummaryStatistics writes the summary statistics
func (r *Reporter) writeSummaryStatistics(file *os.File, stats *Statistics) {
	fmt.Fprintf(file, "## Summary Statistics\n\n")
	fmt.Fprintf(file, "| Metric | Value |\n")
	fmt.Fprintf(file, "|--------|-------|\n")
	fmt.Fprintf(file, "| Total Objects | %d |\n", stats.TotalObjects)
	fmt.Fprintf(file, "| Valid Objects | %d |\n", stats.ValidObjects)
	fmt.Fprintf(file, "| Invalid Objects | %d |\n", stats.InvalidObjects)
	fmt.Fprintf(file, "| Total Errors | %d |\n", stats.TotalErrors)
	fmt.Fprintf(file, "| Total Warnings | %d |\n", stats.TotalWarnings)
	fmt.Fprintf(file, "| Processing Time | %v |\n\n", stats.Duration)
}

// writeDetailedStatistics writes detailed statistics by object type
func (r *Reporter) writeDetailedStatistics(file *os.File, stats *Statistics) {
	fmt.Fprintf(file, "## Statistics by Object Type\n\n")
	fmt.Fprintf(file, "| Object Type | Count | Valid | Invalid | Errors | Warnings |\n")
	fmt.Fprintf(file, "|-------------|-------|-------|---------|--------|----------|\n")

	// Sort object types for consistent output
	var objectTypes []string
	for objType := range stats.ObjectsByType {
		objectTypes = append(objectTypes, objType)
	}
	sort.Strings(objectTypes)

	for _, objType := range objectTypes {
		typeStats := stats.ObjectsByType[objType]
		fmt.Fprintf(file, "| %s | %d | %d | %d | %d | %d |\n",
			objType, typeStats.Count, typeStats.Valid, typeStats.Invalid, typeStats.Errors, typeStats.Warnings)
	}
	fmt.Fprintf(file, "\n")
}

// writeErrorWarningSummary writes error and warning summary
func (r *Reporter) writeErrorWarningSummary(file *os.File, stats *Statistics) {
	if stats.TotalErrors > 0 {
		fmt.Fprintf(file, "## Error Summary\n\n")
		fmt.Fprintf(file, "| Error Code | Count |\n")
		fmt.Fprintf(file, "|-------------|-------|\n")

		var errorCodes []string
		for code := range stats.ErrorsByType {
			errorCodes = append(errorCodes, code)
		}
		sort.Strings(errorCodes)

		for _, code := range errorCodes {
			fmt.Fprintf(file, "| %s | %d |\n", code, stats.ErrorsByType[code])
		}
		fmt.Fprintf(file, "\n")
	}

	if stats.TotalWarnings > 0 {
		fmt.Fprintf(file, "## Warning Summary\n\n")
		fmt.Fprintf(file, "| Warning Code | Count |\n")
		fmt.Fprintf(file, "|---------------|-------|\n")

		var warningCodes []string
		for code := range stats.WarningsByType {
			warningCodes = append(warningCodes, code)
		}
		sort.Strings(warningCodes)

		for _, code := range warningCodes {
			fmt.Fprintf(file, "| %s | %d |\n", code, stats.WarningsByType[code])
		}
		fmt.Fprintf(file, "\n")
	}
}

// writeDetailedResults writes detailed validation results
func (r *Reporter) writeDetailedResults(file *os.File, results []ValidationResult) {
	fmt.Fprintf(file, "## Detailed Validation Results\n\n")

	// Group results by object type
	resultsByType := make(map[string][]ValidationResult)
	for _, result := range results {
		resultsByType[result.ObjectType] = append(resultsByType[result.ObjectType], result)
	}

	// Sort object types for consistent output
	var objectTypes []string
	for objType := range resultsByType {
		objectTypes = append(objectTypes, objType)
	}
	sort.Strings(objectTypes)

	for _, objType := range objectTypes {
		typeResults := resultsByType[objType]
		fmt.Fprintf(file, "### %s Objects\n\n", objType)

		for _, result := range typeResults {
			fmt.Fprintf(file, "#### Object: %s\n\n", result.ObjectID)
			fmt.Fprintf(file, "- **Type:** %s\n", result.ObjectType)
			fmt.Fprintf(file, "- **Valid:** %t\n", result.Valid)
			fmt.Fprintf(file, "- **Timestamp:** %s\n", result.Timestamp.Format("2006-01-02 15:04:05 UTC"))

			if len(result.Errors) > 0 {
				fmt.Fprintf(file, "- **Errors:**\n")
				for _, err := range result.Errors {
					fmt.Fprintf(file, "  - %s: %s (Code: %s)\n", err.Field, err.Message, err.Code)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Fprintf(file, "- **Warnings:**\n")
				for _, warning := range result.Warnings {
					fmt.Fprintf(file, "  - %s: %s (Code: %s)\n", warning.Field, warning.Message, warning.Code)
				}
			}

			fmt.Fprintf(file, "\n")
		}
	}
}

// writeReportFooter writes the report footer
func (r *Reporter) writeReportFooter(file *os.File, report *ValidationReport) {
	fmt.Fprintf(file, "---\n\n")
	fmt.Fprintf(file, "*Report generated by TMForum Proxy Validator*\n")
	fmt.Fprintf(file, "*Generated at: %s*\n", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
}
