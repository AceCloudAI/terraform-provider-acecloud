package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// cliConfig mirrors the ace CLI config.json structure.
// See: the local credentials file
type cliConfig struct {
	APIBaseURL        string `json:"api_base_url"`
	Token             string `json:"token,omitempty"`
	APIKeyID          string `json:"api_key_id,omitempty"`
	APIKeySecret      string `json:"api_key_secret,omitempty"`
	APIKeyServiceName string `json:"api_key_service_name,omitempty"`
	Region            string `json:"region"`
	ProjectID         string `json:"project_id"`
}

// defaultCLIConfigPath returns ~/.ace/config.json.
func defaultCLIConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ace", "config.json")
}

// readCLIConfig reads the ace CLI config file and returns an AuthResult.
func readCLIConfig(path string) (*AuthResult, error) {
	if path == "" {
		path = defaultCLIConfigPath()
	}

	// Check file permissions — warn if config is readable by others
	info, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("ace CLI config file not found at %s (run 'ace auth login' to create it)", path)
		}
		return nil, fmt.Errorf("failed to stat ace CLI config at %s: %w", path, statErr)
	}
	if info.Mode().Perm()&0077 != 0 {
		fmt.Fprintf(os.Stderr, "[WARN] ace CLI config %s has overly permissive permissions (%o). Consider running: chmod 600 %s\n", path, info.Mode().Perm(), path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ace CLI config at %s: %w", path, err)
	}

	var cfg cliConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse ace CLI config at %s: %w", path, err)
	}

	// Prefer API key over JWT token if both are present (matches ace CLI precedence).
	if cfg.APIKeyID != "" && cfg.APIKeySecret != "" {
		if cfg.APIKeyServiceName == "" {
			return nil, fmt.Errorf("ace CLI config at %s has API key credentials but no api_key_service_name (re-authenticate with 'ace auth login-api-key --service-name <name>')", path)
		}
		return &AuthResult{
			APIKeyID:          cfg.APIKeyID,
			APIKeySecret:      cfg.APIKeySecret,
			APIKeyServiceName: cfg.APIKeyServiceName,
			Region:            cfg.Region,
			ProjectID:         cfg.ProjectID,
			Method:            "cli_config",
		}, nil
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("ace CLI config at %s has no credentials (run 'ace auth login' or 'ace auth login-api-key')", path)
	}

	return &AuthResult{
		Token:     cfg.Token,
		Region:    cfg.Region,
		ProjectID: cfg.ProjectID,
		Method:    "cli_config",
	}, nil
}

// ReadCLIConfigAPIURL reads just the api_base_url from the ace CLI config.
func ReadCLIConfigAPIURL(path string) string {
	if path == "" {
		path = defaultCLIConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg cliConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.APIBaseURL
}
