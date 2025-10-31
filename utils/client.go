
package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.skilzy.ai"
	UserAgent      = "skilzy-cli/1.0.0"
)

// SkilzyClient handles API requests to the Skilzy registry
type SkilzyClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewSkilzyClient creates a new API client
func NewSkilzyClient(apiKey string) *SkilzyClient {
	return &SkilzyClient{
		BaseURL: DefaultBaseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
		 },
	}
}

// PublishSkillResponse represents the API response from publishing a skill
type PublishSkillResponse struct {
	Skill   string `json:"skill"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// SearchResult represents a skill in search results
type SearchResult struct {
	Name          string `json:"name"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	LatestVersion string `json:"latest_version"`
}

// SearchResponse represents the API response from searching skills
type SearchResponse struct {
	Data  []SearchResult `json:"data"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// MySkillLatestVersion represents version info for published skills
type MySkillLatestVersion struct {
	Version     string `json:"version"`
	Status      string `json:"status"`
	ReviewNotes string `json:"reviewNotes,omitempty"`
}

// MySkill represents a skill owned by the authenticated user
type MySkill struct {
	ID                    int                   `json:"id"`
	Name                  string                `json:"name"`
	Description           string                `json:"description"`
	License               string                `json:"license"`
	LatestVersion         *MySkillLatestVersion `json:"latestVersion"`
	PublishedVersionCount int                   `json:"publishedVersionCount"`
	TotalVersions         int                   `json:"totalVersions"`
}

// SearchSkills searches for skills in the registry
func (c *SkilzyClient) SearchSkills(query string, author string, keywords []string) (*SearchResponse, error) {
	url := c.BaseURL + "/skills/search"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if query != "" {
		q.Add("q", query)
	}
	if author != "" {
		q.Add("author", author)
	}
	if len(keywords) > 0 {
		q.Add("keywords", strings.Join(keywords, ","))
	}
	q.Add("page", "1")
	q.Add("limit", "20")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", UserAgent)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var searchResp SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &searchResp, nil
}

// GetMySkills retrieves all skills published by the authenticated user
func (c *SkilzyClient) GetMySkills() ([]MySkill, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	url := c.BaseURL + "/users/me/skills"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", UserAgent)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var skills []MySkill
	if err := json.Unmarshal(respBody, &skills); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return skills, nil
}

// PublishSkill uploads a skill package to the registry
func (c *SkilzyClient) PublishSkill(packagePath string) (*PublishSkillResponse, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("API key is required for publishing")
	}

	// Validate package file exists
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill package not found at '%s'", packagePath)
	}

	// Extract manifest from the zip file
	manifestContent, err := extractManifestFromZip(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract manifest: %w", err)
	}

	// Prepare multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	fileField, err := writer.CreateFormFile("file", filepath.Base(packagePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	fileData, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package file: %w", err)
	}

	if _, err := fileField.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Add the manifest
	if err := writer.WriteField("manifest", manifestContent); err != nil {
		return nil, fmt.Errorf("failed to add manifest field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create the request
	url := c.BaseURL + "/skills/publish"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", UserAgent)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var publishResp PublishSkillResponse
	if err := json.Unmarshal(respBody, &publishResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &publishResp, nil
}

// extractManifestFromZip extracts the skill.json content from a zip file
func extractManifestFromZip(zipPath string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Find skill.json in the zip
	for _, file := range reader.File {
		// Check if this is skill.json (could be at root or in a subdirectory)
		if filepath.Base(file.Name) == "skill.json" {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open skill.json: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read skill.json: %w", err)
			}

			return string(content), nil
		}
	}

	return "", fmt.Errorf("skill.json not found in package")
}
