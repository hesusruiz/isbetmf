package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// TMFClientConfig holds the configuration for the tmfclient package
type TMFClientConfig struct {
	// BaseURL of the remote TMForum server
	BaseURL string `json:"base_url" yaml:"base_url"`

	// The path prefix to use in all requests to the remote server
	PathPrefix string `json:"path_prefix" yaml:"path_prefix"`

	// Timeout in seconds for HTTP requests
	Timeout int `json:"timeout" yaml:"timeout"`
}

// TMFClient is a client for the TMForum API.
type TMFClient struct {
	config *TMFClientConfig
	client *http.Client
}

// NewClient creates a new client.
func NewClient(config *TMFClientConfig) *TMFClient {
	if config.Timeout == 0 {
		config.Timeout = 10
	}
	if config.PathPrefix == "" {
		config.PathPrefix = "/tmf-api"
	}
	return &TMFClient{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

func (c *TMFClient) TMFPost(req *Request, objMap repository.TMFObjectMap) (repository.TMFObjectMap, []error) {

	requestBody, err := json.Marshal(objMap)
	if err != nil {
		return nil, []error{errl.Errorf("failed to marshall object: %w", err)}
	}

	path := fmt.Sprintf("/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName)

	headers := map[string]string{
		"Authorization": "Bearer " + req.AccessToken,
		"Content-Type":  "application/json",
	}

	resp, err := c.Post(path, requestBody, headers)
	if err != nil {
		return nil, []error{errl.Errorf("remote server returned error: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, []error{errl.Errorf("failed to read response body: %w", err)}
	}

	if resp.StatusCode != http.StatusCreated {
		slog.Error("unexpected status code from remote server", slog.Int("status_code", resp.StatusCode), slog.String("response_body", string(responseBody)), slog.String("path", path))

		var errs []error
		// Try to parse the response body as a TMF error format or generic map
		var errorResponse map[string]any
		if jsonErr := json.Unmarshal(responseBody, &errorResponse); jsonErr == nil {
			if reason, ok := errorResponse["reason"].(string); ok {
				errs = append(errs, errl.Errorf("remote server returned status %d: %s", resp.StatusCode, reason))
			}
			if message, ok := errorResponse["message"].(string); ok {
				errs = append(errs, errl.Errorf("remote server message: %s", message))
			}
		}

		if len(errs) == 0 {
			errs = append(errs, errl.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody)))
		}
		return nil, errs
	}

	obj, err := repository.NewTMFObjectMapFromBytes(req.ResourceName, responseBody)
	if err != nil {
		return nil, []error{errl.Errorf("failed to bind request body: %w", err)}
	}

	validation := obj.Validate(req.ResourceName)
	if len(validation.Errors) > 0 {
		var errs []error
		for _, vErr := range validation.Errors {
			errs = append(errs, errl.Errorf("validation error on field '%s': %s (code: %s)", vErr.Field, vErr.Message, vErr.Code))
		}
		return nil, errs
	}

	return obj, nil

}

func (c *TMFClient) TMFPatch(req *Request, patchMap repository.TMFObjectMap) (repository.TMFObjectMap, []error) {

	requestBody, err := json.Marshal(patchMap)
	if err != nil {
		return nil, []error{errl.Errorf("failed to marshall object: %w", err)}
	}

	path := fmt.Sprintf("/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID)

	headers := map[string]string{
		"Authorization": "Bearer " + req.AccessToken,
		"Content-Type":  "application/json",
	}

	resp, err := c.Patch(path, requestBody, headers)
	if err != nil {
		return nil, []error{errl.Errorf("remote server returned error: %w", err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, []error{errl.Errorf("failed to read response body: %w", err)}
	}

	if resp.StatusCode >= 300 {
		slog.Error("unexpected status code from remote server", slog.Int("status_code", resp.StatusCode), slog.String("response_body", string(responseBody)), slog.String("path", path))

		var errs []error
		// Try to parse the response body as a TMF error format or generic map
		var errorResponse map[string]any
		if jsonErr := json.Unmarshal(responseBody, &errorResponse); jsonErr == nil {
			if reason, ok := errorResponse["reason"].(string); ok {
				errs = append(errs, errl.Errorf("remote server returned status %d: %s", resp.StatusCode, reason))
			}
			if message, ok := errorResponse["message"].(string); ok {
				errs = append(errs, errl.Errorf("remote server message: %s", message))
			}
		}

		if len(errs) == 0 {
			errs = append(errs, errl.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(responseBody)))
		}
		return nil, errs
	}

	obj, err := repository.NewTMFObjectMapFromBytes(req.ResourceName, responseBody)
	if err != nil {
		return nil, []error{errl.Errorf("failed to bind request body: %w", err)}
	}

	validation := obj.Validate(req.ResourceName)
	if len(validation.Errors) > 0 {
		var errs []error
		for _, vErr := range validation.Errors {
			errs = append(errs, errl.Errorf("validation error on field '%s': %s (code: %s)", vErr.Field, vErr.Message, vErr.Code))
		}
		return nil, errs
	}

	return obj, nil

}

func (c *TMFClient) TMFGetList(path string, headers map[string]string) ([]repository.TMFObjectMap, error) {

	resp, err := c.Get(path, headers)
	if err != nil {
		return nil, errl.Errorf("remote server returned error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errl.Errorf("failed to read response body: %w", err)
	}

	// Check the content type of the response and return an error if it is not JSON
	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, errl.Errorf("remote server returned non-JSON content type: %s", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode >= 300 {
		slog.Error("unexpected status code from remote server", slog.Int("status_code", resp.StatusCode), slog.String("response_body", string(body)), slog.String("path", path))

		var errMsgs []string
		var errorResponse map[string]any
		if jsonErr := json.Unmarshal(body, &errorResponse); jsonErr == nil {
			if reason, ok := errorResponse["reason"].(string); ok {
				errMsgs = append(errMsgs, reason)
			}
			if message, ok := errorResponse["message"].(string); ok {
				errMsgs = append(errMsgs, message)
			}
		}

		if len(errMsgs) > 0 {
			return nil, errl.Errorf("remote server returned status %d: %s", resp.StatusCode, strings.Join(errMsgs, ", "))
		}

		return nil, errl.Errorf("remote server returned status %d: %s", resp.StatusCode, string(body))
	}

	var objects []repository.TMFObjectMap
	if err := json.Unmarshal(body, &objects); err != nil {
		return nil, errl.Errorf("remote server returned invalid JSON: %w", err)
	}

	return objects, nil
}

// Get sends a GET request to the remote server.
func (c *TMFClient) Get(path string, headers map[string]string) (*http.Response, error) {
	return c.do("GET", path, nil, headers)
}

// Post sends a POST request to the remote server.
func (c *TMFClient) Post(path string, body []byte, headers map[string]string) (*http.Response, error) {
	return c.do("POST", path, body, headers)
}

// Patch sends a PATCH request to the remote server.
func (c *TMFClient) Patch(path string, body []byte, headers map[string]string) (*http.Response, error) {
	return c.do("PATCH", path, body, headers)
}

// Delete sends a DELETE request to the remote server.
func (c *TMFClient) Delete(path string, headers map[string]string) (*http.Response, error) {
	return c.do("DELETE", path, nil, headers)
}

// do sends an HTTP request to the remote server.
// It uses the BaseURL and PathPrefix for the server from the configuration.
func (c *TMFClient) do(method, path string, body []byte, headers map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s%s", c.config.BaseURL, c.config.PathPrefix, path)
	slog.Debug("sending", slog.String("method", method), "url", url)

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, errl.Errorf("failed to create request for %s: %w", url, err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errl.Errorf("error sending %s request to %s: %w", method, url, err)
	}

	return resp, nil
}
