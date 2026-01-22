package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	// Server config (from config file, can be overridden by env vars)
	Server struct {
		Timeout        int      `mapstructure:"timeout"` // request timeout in seconds
		Port           string   `mapstructure:"port"`
		Mode           string   `mapstructure:"mode"`            // debug, release
		AllowedOrigins []string `mapstructure:"allowed_origins"` // CORS allowed origins
	} `mapstructure:"server"`

	// Secrets (from .env only)
	APIKey  string `mapstructure:"-"`
	JWT     struct {
		Secret string `mapstructure:"-"` // from .env only
	} `mapstructure:"jwt"`

	// Database config (from config file + env vars, password from .env only)
	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"-"` // from .env only
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	// App config (from config file, can be overridden by env vars)
	App struct {
		LogLevel  string `mapstructure:"log_level"`  // debug, info, warn, error
		LogFormat string `mapstructure:"log_format"` // json, text
	} `mapstructure:"app"`
}

// Load loads configuration from config file and environment variables
// Priority: Env vars > Config file > Defaults
func Load(configPath string) (*Config, error) {
	// Step 1: Load .env file for secrets (optional, for local development)
	// nolint:errcheck // .env file is optional
	_ = godotenv.Load()

	// Step 2: Set default values
	setDefaults()

	// Step 3: Setup environment variable binding FIRST (before reading config)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Step 4: Read config file (but env vars will override)
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Step 5: Unmarshal config (viper.Get() checks env vars automatically)
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Step 6: Load secrets from environment variables (never from config file)
	cfg.APIKey = os.Getenv("API_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API_KEY is required (set in .env or environment)")
	}

	// Allow DATABASE_PASSWORD or DB_PASSWORD
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	// if dbPassword == "" {
	// 	dbPassword = os.Getenv("DB_PASSWORD")
	// }
	if dbPassword == "" {
		return nil, fmt.Errorf("DATABASE_PASSWORD is required (set in .env or environment)")
	}
	cfg.Database.Password = dbPassword

	// Load JWT secret from environment variable
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required (set in .env or environment)")
	}

	// Step 7: Explicitly check for env var overrides for common keys
	// This ensures Unmarshal didn't miss env vars
	checkEnvOverrides(cfg)

	return cfg, nil
}

// checkEnvOverrides explicitly checks environment variables for overrides
// This ensures values like SERVER_PORT override config file values
func checkEnvOverrides(cfg *Config) {
	// Server overrides
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
	if timeout := os.Getenv("SERVER_TIMEOUT"); timeout != "" {
		// nolint:errcheck // Partial parse is acceptable, default value if invalid
		fmt.Sscanf(timeout, "%d", &cfg.Server.Timeout)
	}

	// Database overrides (except password)
	if host := os.Getenv("DATABASE_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DATABASE_PORT"); port != "" {
		cfg.Database.Port = port
	}
	if user := os.Getenv("DATABASE_USER"); user != "" {
		cfg.Database.User = user
	}
	if dbname := os.Getenv("DATABASE_DBNAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}
	if sslmode := os.Getenv("DATABASE_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}

	// App overrides
	if logLevel := os.Getenv("APP_LOG_LEVEL"); logLevel != "" {
		cfg.App.LogLevel = logLevel
	}
	if logFormat := os.Getenv("APP_LOG_FORMAT"); logFormat != "" {
		cfg.App.LogFormat = logFormat
	}
}

// WatchConfig watches for config file changes and calls the callback
func WatchConfig(callback func(*Config)) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)

		// Get current env values (secrets)
		apiKey := os.Getenv("API_KEY")
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		jwtSecret := os.Getenv("JWT_SECRET")
		// if dbPassword == "" {
		// 	dbPassword = os.Getenv("DB_PASSWORD")
		// }

		// Unmarshal new config
		cfg := &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			log.Printf("Error reloading config: %v", err)
			return
		}

		// Restore secrets
		cfg.APIKey = apiKey
		cfg.Database.Password = dbPassword
		cfg.JWT.Secret = jwtSecret

		// Apply env overrides
		checkEnvOverrides(cfg)

		// Call user callback with new config
		callback(cfg)
	})
}

// Get dynamically gets any config value by key
// Checks: env vars > config file > defaults
func Get(key string) interface{} {
	return viper.Get(key)
}

// GetString dynamically gets a string config value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt dynamically gets an int config value
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool dynamically gets a bool config value
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// Set dynamically sets a config value at runtime
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

// Reload reloads the config from file and returns new Config
func Reload(configPath string) (*Config, error) {
	return Load(configPath)
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "finance")
	viper.SetDefault("database.dbname", "finance_tracker")
	viper.SetDefault("database.sslmode", "disable")

	// App defaults
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.log_format", "json")
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.DBName, c.Database.SSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}
