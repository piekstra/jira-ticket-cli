package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig creates a temporary config directory for testing
// On macOS, os.UserConfigDir() returns $HOME/Library/Application Support
// So we manipulate HOME to point to a temp directory
func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Save original HOME
	origHome := os.Getenv("HOME")

	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "jira-cli-test-*")
	require.NoError(t, err)

	// Create Library/Application Support structure for macOS
	libDir := filepath.Join(tempDir, "Library", "Application Support")
	err = os.MkdirAll(libDir, 0700)
	require.NoError(t, err)

	// Set HOME to temp directory
	os.Setenv("HOME", tempDir)

	// Return cleanup function
	return libDir, func() {
		os.RemoveAll(tempDir)
		os.Setenv("HOME", origHome)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := &Config{
		Domain:   "testdomain",
		Email:    "test@example.com",
		APIToken: "secret-token",
	}

	// Save config
	err := Save(cfg)
	require.NoError(t, err)

	// Load config
	loaded, err := Load()
	require.NoError(t, err)

	assert.Equal(t, cfg.Domain, loaded.Domain)
	assert.Equal(t, cfg.Email, loaded.Email)
	assert.Equal(t, cfg.APIToken, loaded.APIToken)
}

func TestConfig_Load_NotExists(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Load when file doesn't exist should return empty config
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Domain)
	assert.Empty(t, cfg.Email)
	assert.Empty(t, cfg.APIToken)
}

func TestConfig_Clear(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Save config first
	cfg := &Config{
		Domain:   "testdomain",
		Email:    "test@example.com",
		APIToken: "secret-token",
	}
	err := Save(cfg)
	require.NoError(t, err)

	// Clear config
	err = Clear()
	require.NoError(t, err)

	// Load should return empty config
	loaded, err := Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.Domain)
}

func TestConfig_Clear_NotExists(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Clear when file doesn't exist should not error
	err := Clear()
	assert.NoError(t, err)
}

func TestConfig_FilePermissions(t *testing.T) {
	configBaseDir, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := &Config{
		Domain:   "testdomain",
		Email:    "test@example.com",
		APIToken: "secret-token",
	}
	err := Save(cfg)
	require.NoError(t, err)

	// Check file permissions
	// configBaseDir is Library/Application Support, config goes under jira-ticket-cli/
	configFile := filepath.Join(configBaseDir, configDirName, configFileName)
	info, err := os.Stat(configFile)
	require.NoError(t, err)

	// File should be 0600 (user read/write only)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestGetDomain_EnvOverride(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Save config
	cfg := &Config{Domain: "config-domain"}
	err := Save(cfg)
	require.NoError(t, err)

	// Without env, should return config value
	os.Unsetenv("JIRA_DOMAIN")
	assert.Equal(t, "config-domain", GetDomain())

	// With env, should return env value
	os.Setenv("JIRA_DOMAIN", "env-domain")
	defer os.Unsetenv("JIRA_DOMAIN")
	assert.Equal(t, "env-domain", GetDomain())
}

func TestGetEmail_EnvOverride(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Save config
	cfg := &Config{Email: "config@example.com"}
	err := Save(cfg)
	require.NoError(t, err)

	// Without env, should return config value
	os.Unsetenv("JIRA_EMAIL")
	assert.Equal(t, "config@example.com", GetEmail())

	// With env, should return env value
	os.Setenv("JIRA_EMAIL", "env@example.com")
	defer os.Unsetenv("JIRA_EMAIL")
	assert.Equal(t, "env@example.com", GetEmail())
}

func TestGetAPIToken_EnvOverride(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Save config
	cfg := &Config{APIToken: "config-token"}
	err := Save(cfg)
	require.NoError(t, err)

	// Without env, should return config value
	os.Unsetenv("JIRA_API_TOKEN")
	assert.Equal(t, "config-token", GetAPIToken())

	// With env, should return env value
	os.Setenv("JIRA_API_TOKEN", "env-token")
	defer os.Unsetenv("JIRA_API_TOKEN")
	assert.Equal(t, "env-token", GetAPIToken())
}

func TestIsConfigured(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Not configured initially
	assert.False(t, IsConfigured())

	// Partially configured
	cfg := &Config{Domain: "test"}
	Save(cfg)
	assert.False(t, IsConfigured())

	// Fully configured
	cfg = &Config{
		Domain:   "test",
		Email:    "test@example.com",
		APIToken: "token",
	}
	Save(cfg)
	assert.True(t, IsConfigured())
}

func TestIsConfigured_EnvOnly(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Set all env vars
	os.Setenv("JIRA_DOMAIN", "env-domain")
	os.Setenv("JIRA_EMAIL", "env@example.com")
	os.Setenv("JIRA_API_TOKEN", "env-token")
	defer func() {
		os.Unsetenv("JIRA_DOMAIN")
		os.Unsetenv("JIRA_EMAIL")
		os.Unsetenv("JIRA_API_TOKEN")
	}()

	// Should be configured via env vars only
	assert.True(t, IsConfigured())
}

func TestPath(t *testing.T) {
	path := Path()
	assert.Contains(t, path, configDirName)
	assert.Contains(t, path, configFileName)
}
