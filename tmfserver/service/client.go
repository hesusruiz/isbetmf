package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// TMFClientConfig holds the configuration for the tmfclient package
type TMFClientConfig struct {
	// BaseURL of the remote TMForum server
	BaseURL string `json:"base_url" yaml:"base_url"`

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
	return &TMFClient{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

func ForwardTMFPost(req *Request, remoteServer string, objMap repository.TMFObjectMap) (repository.TMFObjectMap, []error) {

	requestBody, err := json.Marshal(objMap)
	if err != nil {
		return nil, []error{errl.Errorf("failed to marshall object: %w", err)}

	}

	path := fmt.Sprintf("/tmf-api/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName)
	fullURL := remoteServer + path

	agent := fiber.Post(fullURL)
	agent.Body(requestBody)

	agent.Set("Authorization", "Bearer "+req.AccessToken)
	agent.Set("Content-Type", "application/json")
	agent.Set("Content-Length", strconv.Itoa(len(requestBody)))

	statusCode, responseBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs
	}

	if statusCode != fiber.StatusCreated {
		// Log the response body
		slog.Error("unexpected status code", slog.Int("status_code", statusCode), slog.String("response_body", string(responseBody)))
		return nil, []error{errl.Errorf("unexpected status code: %d, response body: %s", statusCode, string(responseBody))}
	}

	obj, err := repository.NewTMFObjectMapFromBytes(req.ResourceName, responseBody)
	if err != nil {
		return nil, []error{errl.Errorf("failed to bind request body: %w", err)}
	}

	validation := obj.Validate(req.ResourceName)
	if len(validation.Errors) > 0 {
		return nil, []error{errl.Errorf("validation errors: %v", validation.Errors)}
	}

	return obj, nil

}

// Get sends a GET request to the remote server.
func (c *TMFClient) Get(path string, headers map[string]string) (*http.Response, error) {
	return c.do("GET", path, nil, headers)
}

func (c *TMFClient) ForwardTMFGetList(path string, headers map[string]string) ([]repository.TMFObjectMap, error) {

	method := "GET"
	url := fmt.Sprintf("%s%s", c.config.BaseURL, path)
	slog.Debug("sending", slog.String("method", method), "url", url)

	var req *http.Request
	var err error

	req, err = http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errl.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errl.Errorf("remote server %s returned error: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errl.Errorf("remote server %s returned error: %w", url, err)
	}

	// Check the content type of the response and return an error if it is not JSON
	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, errl.Errorf("remote server %s returned non-JSON content type: %s", url, resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode >= 300 {
		return nil, errl.Errorf("remote server %s returned status %d: %s", url, resp.StatusCode, string(body))
	}

	var objects []repository.TMFObjectMap
	if err := json.Unmarshal(body, &objects); err != nil {
		return nil, errl.Errorf("remote server %s returned invalid JSON: %w", url, err)
	}

	return objects, nil
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

func (c *TMFClient) do(method, path string, body []byte, headers map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.BaseURL, path)
	slog.Debug("sending", slog.String("method", method), "url", url)

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}
