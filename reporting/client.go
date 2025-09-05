package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an HTTP client for connecting to TMForum servers
type Client struct {
	httpClient *http.Client
	baseURL    string
	timeout    time.Duration
}

// NewClient creates a new TMForum client
func NewClient(config *Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		baseURL: config.BaseURL,
		timeout: time.Duration(config.Timeout) * time.Second,
	}
}

// GetObjectsWithPagination retrieves all objects of a specific type using pagination
func (c *Client) GetObjectsWithPagination(ctx context.Context, objectType string, config *Config) ([]TMFObject, error) {
	// Get the path prefix for this object type from the routes map
	pathPrefix, exists := GeneratedDefaultResourceToPathPrefixV4[objectType]
	if !exists {
		return nil, fmt.Errorf("unknown object type: %s", objectType)
	}

	var allObjects []TMFObject
	limit := config.PageSize
	offset := 0

	for {
		// Build URL with pagination parameters
		url := fmt.Sprintf("%s%s?limit=%d&offset=%d", c.baseURL, pathPrefix, limit, offset)
		fmt.Printf("Retrieving %s objects: offset=%d, limit=%d\n", objectType, offset, limit)

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
		var objects []TMFObject
		if err := json.Unmarshal(body, &objects); err != nil {
			// If it's not an array, try to parse as a single object
			var singleObject TMFObject
			if err := json.Unmarshal(body, &singleObject); err != nil {
				return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
			}
			objects = []TMFObject{singleObject}
		}

		// Process each object to extract additional fields
		for i := range objects {
			objects[i] = c.processObject(objects[i], body)
		}

		// Add objects from this page to the total collection
		allObjects = append(allObjects, objects...)

		// If we got fewer objects than the limit, we've reached the end
		if len(objects) < limit {
			break
		}

		// Move to next page
		offset += limit

		// Safety check to prevent infinite loops
		if offset >= config.MaxObjects {
			fmt.Printf("Warning: Reached maximum objects limit (%d) for %s\n", config.MaxObjects, objectType)
			break
		}
	}

	fmt.Printf("Total %s objects retrieved: %d\n", objectType, len(allObjects))
	return allObjects, nil
}

// GetObjects retrieves all objects of a specific type, using pagination if enabled
func (c *Client) GetObjects(ctx context.Context, objectType string, config *Config) ([]TMFObject, error) {
	if config.PaginationEnabled {
		return c.GetObjectsWithPagination(ctx, objectType, config)
	}
	return c.GetObjectsWithoutPagination(ctx, objectType)
}

// GetObjectsWithoutPagination retrieves objects without pagination (legacy method)
func (c *Client) GetObjectsWithoutPagination(ctx context.Context, objectType string) ([]TMFObject, error) {
	// Get the path prefix for this object type from the routes map
	pathPrefix, exists := GeneratedDefaultResourceToPathPrefixV4[objectType]
	if !exists {
		return nil, fmt.Errorf("unknown object type: %s", objectType)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, pathPrefix)
	fmt.Printf("URL: %s\n", url)

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
	var objects []TMFObject
	if err := json.Unmarshal(body, &objects); err != nil {
		// If it's not an array, try to parse as a single object
		var singleObject TMFObject
		if err := json.Unmarshal(body, &singleObject); err != nil {
			return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
		}
		objects = []TMFObject{singleObject}
	}

	// Process each object to extract additional fields
	for i := range objects {
		objects[i] = c.processObject(objects[i], body)
	}

	return objects, nil
}

// processObject processes a TMF object and extracts additional fields
func (c *Client) processObject(obj TMFObject, rawBody []byte) TMFObject {
	// Parse the raw JSON to get all fields
	var rawObj map[string]any
	if err := json.Unmarshal(rawBody, &rawObj); err != nil {
		return obj
	}

	// Extract additional fields that are not in our struct
	obj.AdditionalFields = make(map[string]any)
	for key, value := range rawObj {
		switch key {
		case "id", "href", "lastUpdate", "version", "@type", "relatedParty":
			// Skip fields already handled by our struct
			continue
		default:
			obj.AdditionalFields[key] = value
		}
	}

	return obj
}

// TestConnection tests the connection to the remote server
func (c *Client) TestConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// GetServerInfo retrieves basic information about the server
func (c *Client) GetServerInfo(ctx context.Context) (map[string]any, error) {
	url := fmt.Sprintf("%s/", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var info map[string]any
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	return info, nil
}
