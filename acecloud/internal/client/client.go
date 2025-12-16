package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/AceCloudAI/terraform-provider-acecloud/acecloud/internal/client/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AceCloudClient struct {
	BaseURL    string
	APIKey     string
	Region     string
	ProjectID  string
	HTTPClient *http.Client
}

func NewAceCloudClient(baseURL, apiKey, region, projectID string) *AceCloudClient {
	return &AceCloudClient{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		Region:    region,
		ProjectID: projectID,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AceCloudClient) CreateVM(ctx context.Context, vmReq *types.VMCreateRequest) (*types.VMCreateResponse, error) {
	endpoint := fmt.Sprintf("%s/cloud/instances", c.BaseURL)
	tflog.Debug(ctx, fmt.Sprintf("Creating VM with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	req, err := c.newRequest(ctx, "POST", fullURL, vmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var createResp types.VMCreateResponse
	if err := c.doRequest(req, &createResp); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	if createResp.Error {
		return nil, fmt.Errorf("API returned error: %s", createResp.Message)
	}

	return &createResp, nil
}

func (c *AceCloudClient) GetVM(ctx context.Context, id string) (*types.VMGetResponse, error) {
	endpoint := fmt.Sprintf("%s/cloud/instances/%s", c.BaseURL, id)
	tflog.Debug(ctx, fmt.Sprintf("Getting VM with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	req, err := c.newRequest(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var getResp types.VMGetResponse
	if err := c.doRequest(req, &getResp); err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	if getResp.Error {
		return nil, fmt.Errorf("API returned error: %s", getResp.Message)
	}

	return &getResp, nil
}

// DeleteVMs deletes one or more VMs by IDs using the bulk-delete endpoint.
// The API expects a JSON body like: {"key":"id","values":["id1","id2"]}
func (c *AceCloudClient) DeleteVMs(ctx context.Context, ids []string) (*types.DeleteResponse, error) {
	endpoint := fmt.Sprintf("%s/cloud/instances", c.BaseURL)
	tflog.Debug(ctx, fmt.Sprintf("Deleting VMs with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	body := map[string]interface{}{
		"key":    "id",
		"values": ids,
	}

	req, err := c.newRequest(ctx, "DELETE", fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create delete request: %w", err)
	}

	var delResp types.DeleteResponse
	if err := c.doRequest(req, &delResp); err != nil {
		return nil, fmt.Errorf("failed to delete VMs: %w", err)
	}

	if delResp.Error {
		return nil, fmt.Errorf("API returned error: %s", delResp.Message)
	}

	return &delResp, nil
}

// UpdateVM updates a VM's attributes (currently supports updating the name)
func (c *AceCloudClient) UpdateVM(ctx context.Context, id string, body interface{}) (*types.VMUpdateResponse, error) {
	endpoint := fmt.Sprintf("%s/cloud/instances/%s", c.BaseURL, id)
	tflog.Debug(ctx, fmt.Sprintf("Updating VM with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	req, err := c.newRequest(ctx, "PUT", fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}

	var updResp types.VMUpdateResponse
	if err := c.doRequest(req, &updResp); err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}

	if updResp.Error {
		return nil, fmt.Errorf("API returned error: %s", updResp.Message)
	}

	return &updResp, nil
}

func (c *AceCloudClient) newRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ace-api-key", c.APIKey)
	req.Header.Set("x-api-key-service-name", "acecloud_keypair")
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Failed to dump request: %v", err))
	} else {
		tflog.Debug(ctx, "HTTP Request", map[string]interface{}{
			"request": string(dump),
		})
	}
	return req, nil
}

func (c *AceCloudClient) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	// bodyBytes, err := io.ReadAll(resp.Body)

	if req.URL.Path == "/cloud/key-pairs" || req.Method == "POST" {
		tflog.Debug(req.Context(), fmt.Sprintf("Received response  %v", resp))
	}

	if err != nil {
		tflog.Debug(req.Context(), fmt.Sprintf("HTTP request failed: %v", err))
		return fmt.Errorf("request failed: %w", err)

	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP error status codes
	if resp.StatusCode >= 400 {
		var apiError struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.Message != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, apiError.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if v == nil {
		return nil
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////
//! key-pair related methods
//////////////////////////////////////////////////////////////////////////

// *Create key pair request
func (c *AceCloudClient) CreateKeyPair(ctx context.Context, reqBody *types.KeyPairCreateRequest) (*types.KeyPairData, error) {
	endpoint := fmt.Sprintf("%s/cloud/key-pairs", c.BaseURL)
	tflog.Debug(ctx, fmt.Sprintf("Creating KeyPair with endpoint: %s", endpoint))

	// Build query params from client defaults (Region/ProjectID)
	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	req, err := c.newRequest(ctx, "POST", fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create keypair request: %w", err)
	}

	//*The API response shape is expected to be KeyPairResponse { data: KeyPairData }
	var kpResp types.KeyPairResponse
	if err := c.doRequest(req, &kpResp); err != nil {
		return nil, fmt.Errorf("failed to create keypair: %w", err)
	}

	if kpResp.Error {
		return nil, fmt.Errorf("API returned error: %s", kpResp.Message)
	}
	// tflog.Debug(ctx, fmt.Sprintf("Created KeyPair with ID: %v", kpResp.Data))
	return &kpResp.Data, nil
}

// *GetKeyPair to retrieve a key-pair by id
func (c *AceCloudClient) GetKeyPair(ctx context.Context, id string) (*types.KeyPairData, error) {
	endpoint := fmt.Sprintf("%s/cloud/key-pairs/%s", c.BaseURL, url.PathEscape(id))
	tflog.Debug(ctx, fmt.Sprintf("Getting KeyPair with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	req, err := c.newRequest(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get keypair request: %w", err)
	}

	var kpResp types.KeyPairResponse
	if err := c.doRequest(req, &kpResp); err != nil {

		return nil, fmt.Errorf("failed to get keypair: %w", err)
	}

	if kpResp.Error {
		// treat as not found if the API uses error flag and message; otherwise return error
		return nil, fmt.Errorf("API returned error: %s", kpResp.Message)
	}

	return &kpResp.Data, nil
}

// DeleteKeyPair deletes a key-pair by id.
func (c *AceCloudClient) DeleteKeyPair(ctx context.Context, reqBody *types.KeyPairDeleteRequest, id string) error {
	endpoint := fmt.Sprintf("%s/cloud/key-pairs", c.BaseURL)
	tflog.Debug(ctx, fmt.Sprintf("Deleting KeyPair with endpoint: %s", endpoint))

	params := url.Values{}
	params.Add("region", c.Region)
	params.Add("project_id", c.ProjectID)

	fullURL := endpoint + "?" + params.Encode()

	tflog.Debug(ctx, fmt.Sprintf("Full URL for deleting KeyPair: %s", fullURL))

	req, err := c.newRequest(ctx, "DELETE", fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create delete keypair request: %w", err)
	}

	tflog.Debug(req.Context(), fmt.Sprintf("Deleting KeyPair with ID: %s", reqBody.Values))

	//*in case of we receive any deletion response
	var delResp types.DeleteResponse
	if err := c.doRequest(req, &delResp); err != nil {
		return fmt.Errorf("failed to delete keypair: %w", err)
	}

	// if delResp.Error {
	// 	return fmt.Errorf("API returned error: %s", delResp.Message)
	// }

	return nil
}
