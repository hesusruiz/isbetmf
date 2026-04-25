package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// PagingConfig holds the configuration for the proxy
type PagingConfig struct {

	// Remote server configuration
	BaseURL string `json:"base_url" yaml:"base_url"`
	Timeout int    `json:"timeout" yaml:"timeout"` // in seconds

	// All possible object types to retrieve
	ObjectTypes []string `json:"object_types" yaml:"object_types"`

	// Pagination settings
	PageSize   int `json:"page_size" yaml:"page_size"`     // objects per page
	MaxObjects int `json:"max_objects" yaml:"max_objects"` // maximum objects to retrieve

	// Validation settings
	ValidateRequiredFields bool `json:"validate_required_fields" yaml:"validate_required_fields"`
	ValidateRelatedParty   bool `json:"validate_related_party" yaml:"validate_related_party"`
	FixValidationErrors    bool `json:"fix_validation_errors" yaml:"fix_validation_errors"`

	// Output settings
	OutputDir  string `json:"output_dir" yaml:"output_dir"`
	ReportFile string `json:"report_file" yaml:"report_file"`
}

// ClientWithPaging represents an HTTP client for connecting to TMForum servers supporting pagination
type ClientWithPaging struct {
	httpClient   *http.Client
	baseURL      string
	timeout      time.Duration
	pagingConfig *PagingConfig
}

// NewClientWithPaging creates a new TMForum client
func NewClientWithPaging(config *PagingConfig) *ClientWithPaging {
	// Strip a possible trailing slash from baseURL
	config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")

	return &ClientWithPaging{
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		baseURL:      config.BaseURL,
		timeout:      time.Duration(config.Timeout) * time.Second,
		pagingConfig: config,
	}
}

func (c *ClientWithPaging) GetObjectURL(objectType string, objectID string) (string, error) {

	pathPrefix, exists := config.GeneratedDefaultResourceToPathPrefix[objectType]
	if !exists {
		return "", fmt.Errorf("unknown object type: %s", objectType)
	}
	return fmt.Sprintf("%s%s/%s", c.baseURL, pathPrefix, objectID), nil
}

// GetAllObjectsOfType retrieves all objects of a specific type using pagination
func (c *ClientWithPaging) GetAllObjectsOfType(ctx context.Context, objectType string, processObject processObject) ([]repo.TMFObjectMap, error) {
	var allObjects []repo.TMFObjectMap
	limit := c.pagingConfig.PageSize
	offset := 0

	for {

		objects, err := c.GetPageOfObjects(ctx, objectType, offset, processObject)
		// In case of error, we stop retrieving pages and return the ones retrieved until now
		if err != nil {
			err = errl.Error(err)
			slog.Error("Error retrieving page of objects", "object_type", objectType, "offset", offset, "error", err)
			fmt.Printf("Total %s objects retrieved: %d\n", objectType, len(allObjects))
			return allObjects, nil
		}

		// Add objects from this page to the total collection
		allObjects = append(allObjects, objects...)

		// If we got fewer objects than the limit specified by the user, we do not need to fetch more objects.
		if len(objects) < limit {
			break
		}

		// Move to next page
		offset += limit

		// Safety check to prevent huge requests
		if offset >= c.pagingConfig.MaxObjects {
			fmt.Printf("Warning: Reached maximum objects limit (%d) for %s\n", c.pagingConfig.MaxObjects, objectType)
			break
		}
	}

	fmt.Printf("Total %s objects retrieved: %d\n", objectType, len(allObjects))
	return allObjects, nil
}

// GetPageOfObjects retrieves up to a single page of objects of a specific type
func (c *ClientWithPaging) GetPageOfObjects(ctx context.Context, resourceName string, offset int, processObject processObject) ([]repo.TMFObjectMap, error) {
	// Get the path prefix for this object type from the routes map
	pathPrefix, exists := config.GeneratedDefaultResourceToPathPrefix[resourceName]
	if !exists {
		return nil, fmt.Errorf("unknown object type: %s", resourceName)
	}

	limit := c.pagingConfig.PageSize

	// Build URL with pagination parameters
	url := fmt.Sprintf("%s%s?limit=%d&offset=%d", c.baseURL, pathPrefix, limit, offset)
	fmt.Printf("Retrieving %s objects: offset=%d, limit=%d\n", resourceName, offset, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as an array first
	var objects []repo.TMFObjectMap
	if err := json.Unmarshal(body, &objects); err != nil {
		// If it's not an array, try to parse as a single object
		var singleObject repo.TMFObjectMap
		if err := json.Unmarshal(body, &singleObject); err != nil {
			return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
		}
		objects = []repo.TMFObjectMap{singleObject}
	}

	// Process each object with the user-supplied logic
	var cont bool
	if processObject != nil {
		for i := range objects {
			objects[i], cont, err = processObject(objects[i])
			// In case of error we just log it and continue with the next object
			if err != nil {
				err = errl.Error(err)
				slog.Error("processing object", "object_id", objects[i].ID(), "error", err)
			}
			// If the user wants to stop processing, we return the objects retrieved so far
			if !cont {
				return objects, nil
			}
		}
	}

	return objects, nil
}

func (c *ClientWithPaging) GetSingleObject(ctx context.Context, objectType string, id string) (repo.TMFObjectMap, error) {
	// Get the path prefix for this object type from the routes map
	pathPrefix, exists := config.GeneratedDefaultResourceToPathPrefix[objectType]
	if !exists {
		return nil, fmt.Errorf("unknown object type: %s", objectType)
	}

	// Build URL with pagination parameters
	url := fmt.Sprintf("%s%s/%s", c.baseURL, pathPrefix, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If it's not an array, try to parse as a single object
	var singleObject repo.TMFObjectMap
	if err := json.Unmarshal(body, &singleObject); err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	return singleObject, nil
}

// DefaultPagingConfig returns a default configuration
func DefaultPagingConfig() *PagingConfig {
	return &PagingConfig{
		BaseURL:                "https://tmf.dome-marketplace-sbx.org",
		Timeout:                30,
		ObjectTypes:            DefaultObjectTypes(),
		PageSize:               100,
		MaxObjects:             10000,
		ValidateRequiredFields: true,
		ValidateRelatedParty:   true,
		FixValidationErrors:    false,
		OutputDir:              "./reports",
		ReportFile:             "tmf_validation_report.md",
	}
}

// DefaultObjectTypes returns the default list of object types to retrieve
func DefaultObjectTypes() []string {

	objectTypes := []string{}
	for objectType := range config.GeneratedDefaultResourceToPathPrefix {
		objectTypes = append(objectTypes, objectType)
	}

	return objectTypes
}
