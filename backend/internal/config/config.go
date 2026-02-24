package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	App       AppConfig       `yaml:"app"`
	Admin     AdminConfig     `yaml:"admin"`
	Binance   BinanceConfig   `yaml:"binance"`
	Symbols   []string        `yaml:"symbols"`
	Snapshot  SnapshotConfig  `yaml:"snapshot"`
	Trading   TradingConfig   `yaml:"trading"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// DSN returns the PostgreSQL data source name
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.DBName,
	)
}

// AppConfig holds application server configuration
type AppConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Addr returns the server address
func (a AppConfig) Addr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// AdminConfig holds admin authentication configuration
type AdminConfig struct {
	Password string `yaml:"password"`
}

// BinanceConfig holds Binance WebSocket configuration
type BinanceConfig struct {
	WSURL string `yaml:"ws_url"`
}

// SnapshotConfig holds snapshot scheduler configuration
type SnapshotConfig struct {
	Interval string `yaml:"interval"`
}

// IntervalDuration returns the snapshot interval as time.Duration
func (s SnapshotConfig) IntervalDuration() (time.Duration, error) {
	return time.ParseDuration(s.Interval)
}

// TradingConfig holds trading parameters configuration
type TradingConfig struct {
	MakerFeeRate    string `yaml:"maker_fee_rate"`   // Maker 手续费率
	FundingRate    string `yaml:"funding_rate"`     // 资金费率
	FundingInterval string `yaml:"funding_interval"` // 资金结算间隔
}

// FundingIntervalDuration returns the funding interval as time.Duration
func (t TradingConfig) FundingIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(t.FundingInterval)
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse environment variables in the config
	parsedData := os.ExpandEnv(string(data))

	// Unmarshal YAML
	var cfg Config
	if err := yaml.Unmarshal([]byte(parsedData), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Port == 0 {
		return fmt.Errorf("database.port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database.dbname is required")
	}
	if c.App.Port == 0 {
		return fmt.Errorf("app.port is required")
	}
	if c.Admin.Password == "" {
		return fmt.Errorf("admin.password is required")
	}
	if c.Binance.WSURL == "" {
		return fmt.Errorf("binance.ws_url is required")
	}
	if len(c.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}
	if c.Snapshot.Interval == "" {
		return fmt.Errorf("snapshot.interval is required")
	}
	if _, err := c.Snapshot.IntervalDuration(); err != nil {
		return fmt.Errorf("invalid snapshot.interval: %w", err)
	}
	if c.Trading.MakerFeeRate == "" {
		return fmt.Errorf("trading.maker_fee_rate is required")
	}
	if c.Trading.FundingInterval == "" {
		return fmt.Errorf("trading.funding_interval is required")
	}
	if _, err := c.Trading.FundingIntervalDuration(); err != nil {
		return fmt.Errorf("invalid trading.funding_interval: %w", err)
	}
	return nil
}
