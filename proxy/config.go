package proxy

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const VersionV4 = "v4"
const VersionV5 = "v5"

// Config holds the configuration for the proxy
type Config struct {
	// Version of the TMForum API
	Version string `json:"version" yaml:"version"`

	// Remote server configuration
	BaseURL string `json:"base_url" yaml:"base_url"`
	Timeout int    `json:"timeout" yaml:"timeout"` // in seconds

	// Object types to retrieve
	ObjectTypes []string `json:"object_types" yaml:"object_types"`

	// Pagination settings
	PaginationEnabled bool `json:"pagination_enabled" yaml:"pagination_enabled"`
	PageSize          int  `json:"page_size" yaml:"page_size"`     // objects per page
	MaxObjects        int  `json:"max_objects" yaml:"max_objects"` // maximum objects to retrieve

	// Validation settings
	ValidateRequiredFields bool `json:"validate_required_fields" yaml:"validate_required_fields"`
	ValidateRelatedParty   bool `json:"validate_related_party" yaml:"validate_related_party"`

	// Output settings
	OutputDir  string `json:"output_dir" yaml:"output_dir"`
	ReportFile string `json:"report_file" yaml:"report_file"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:                VersionV4,
		BaseURL:                "https://tmf.dome-marketplace-sbx.org",
		Timeout:                30,
		ObjectTypes:            DefaultObjectTypes(),
		PaginationEnabled:      true,
		PageSize:               100,
		MaxObjects:             10000,
		ValidateRequiredFields: true,
		ValidateRelatedParty:   true,
		OutputDir:              "./reports",
		ReportFile:             "tmf_validation_report.md",
	}
}

// DefaultObjectTypes returns the default list of object types to retrieve
func DefaultObjectTypes() []string {

	objectTypes := []string{}
	for objectType, _ := range GeneratedDefaultResourceToPathPrefixV4 {
		objectTypes = append(objectTypes, objectType)
	}

	return objectTypes
}

// LoadConfigFromEnv loads configuration from environment variables
func (c *Config) LoadConfigFromEnv() {
	if baseURL := os.Getenv("TMF_BASE_URL"); baseURL != "" {
		c.BaseURL = baseURL
	}

	if timeout := os.Getenv("TMF_TIMEOUT"); timeout != "" {
		if t, err := fmt.Sscanf(timeout, "%d", &c.Timeout); err == nil && t > 0 {
			c.Timeout = t
		}
	}

	if objectTypes := os.Getenv("TMF_OBJECT_TYPES"); objectTypes != "" {
		c.ObjectTypes = strings.Split(objectTypes, ",")
		for i, t := range c.ObjectTypes {
			c.ObjectTypes[i] = strings.TrimSpace(t)
		}
	}

	if outputDir := os.Getenv("TMF_OUTPUT_DIR"); outputDir != "" {
		c.OutputDir = outputDir
	}

	if reportFile := os.Getenv("TMF_REPORT_FILE"); reportFile != "" {
		c.ReportFile = reportFile
	}

	// Load pagination settings from environment
	if paginationEnabled := os.Getenv("TMF_PAGINATION_ENABLED"); paginationEnabled != "" {
		if enabled, err := strconv.ParseBool(paginationEnabled); err == nil {
			c.PaginationEnabled = enabled
		}
	}

	if pageSize := os.Getenv("TMF_PAGE_SIZE"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size > 0 {
			c.PageSize = size
		}
	}

	if maxObjects := os.Getenv("TMF_MAX_OBJECTS"); maxObjects != "" {
		if max, err := strconv.Atoi(maxObjects); err == nil && max > 0 {
			c.MaxObjects = max
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if len(c.ObjectTypes) == 0 {
		return fmt.Errorf("at least one object type must be specified")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	c.BaseURL = strings.TrimRight(c.BaseURL, "/")

	return nil
}
