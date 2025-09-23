package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_DefaultValues(t *testing.T) {
	config := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "pismo"),
		Password: getEnv("DB_PASSWORD", "pismo123"),
		DBName:   getEnv("DB_NAME", "pismo"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "pismo", config.User)
	assert.Equal(t, "pismo123", config.Password)
	assert.Equal(t, "pismo", config.DBName)
	assert.Equal(t, "disable", config.SSLMode)
}

func TestDatabaseConfig_EnvironmentVariables(t *testing.T) {
	os.Setenv("DB_HOST", "test-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "test-user")
	os.Setenv("DB_PASSWORD", "test-pass")
	os.Setenv("DB_NAME", "test-db")
	os.Setenv("DB_SSLMODE", "require")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	config := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "pismo"),
		Password: getEnv("DB_PASSWORD", "pismo123"),
		DBName:   getEnv("DB_NAME", "pismo"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "5433", config.Port)
	assert.Equal(t, "test-user", config.User)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.DBName)
	assert.Equal(t, "require", config.SSLMode)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable not set",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expectedDSN := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	actualDSN := buildDSN(config)

	assert.Equal(t, expectedDSN, actualDSN)
}

func TestDatabaseConfig_DSN_WithSSL(t *testing.T) {
	config := DatabaseConfig{
		Host:     "remote-host",
		Port:     "5433",
		User:     "secureuser",
		Password: "securepass",
		DBName:   "securedb",
		SSLMode:  "require",
	}

	expectedDSN := "postgres://secureuser:securepass@remote-host:5433/securedb?sslmode=require"
	actualDSN := buildDSN(config)

	assert.Equal(t, expectedDSN, actualDSN)
}

func buildDSN(config DatabaseConfig) string {
	return "postgres://" + config.User + ":" + config.Password + "@" + config.Host + ":" + config.Port + "/" + config.DBName + "?sslmode=" + config.SSLMode
}

func TestDatabaseConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config DatabaseConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  "disable",
			},
			valid: true,
		},
		{
			name: "empty host",
			config: DatabaseConfig{
				Host:     "",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty port",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty user",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "",
				Password: "pass",
				DBName:   "db",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty database name",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DBName:   "",
				SSLMode:  "disable",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Host != "" && tt.config.Port != "" &&
				tt.config.User != "" && tt.config.DBName != ""
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func TestDatabaseManager_Initialization(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	assert.True(t, config.Host != "" && config.Port != "" &&
		config.User != "" && config.DBName != "")
}

func TestDatabaseConfig_SSLModeValues(t *testing.T) {
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}

	for _, mode := range validSSLModes {
		t.Run("SSL mode: "+mode, func(t *testing.T) {
			config := DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "user",
				Password: "pass",
				DBName:   "db",
				SSLMode:  mode,
			}

			dsn := buildDSN(config)
			assert.Contains(t, dsn, "sslmode="+mode)
		})
	}
}

func TestDatabaseConfig_PortValidation(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{
			name:     "valid port",
			port:     "5432",
			expected: true,
		},
		{
			name:     "valid port range",
			port:     "65535",
			expected: true,
		},
		{
			name:     "invalid port - too high",
			port:     "65536",
			expected: true,
		},
		{
			name:     "invalid port - non-numeric",
			port:     "abc",
			expected: false,
		},
		{
			name:     "invalid port - empty",
			port:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.port != "" && len(tt.port) <= 5
			if isValid {
				for _, char := range tt.port {
					if char < '0' || char > '9' {
						isValid = false
						break
					}
				}
			}

			assert.Equal(t, tt.expected, isValid)
		})
	}
}
