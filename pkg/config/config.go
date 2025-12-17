// Package config provides configuration management for TFDrift-Falco.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Providers     ProvidersConfig     `yaml:"providers"`
	Falco         FalcoConfig         `yaml:"falco"`
	DriftRules    []DriftRule         `yaml:"drift_rules"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Logging       LoggingConfig       `yaml:"logging"`
	AutoImport    AutoImportConfig    `yaml:"auto_import"`
	DryRun        bool                `yaml:"-"`
}

// ProvidersConfig contains cloud provider settings
type ProvidersConfig struct {
	AWS AWSConfig `yaml:"aws"`
	GCP GCPConfig `yaml:"gcp"`
}

// AWSConfig contains AWS-specific settings
type AWSConfig struct {
	Enabled bool                 `yaml:"enabled"`
	Regions []string             `yaml:"regions"`
	State   TerraformStateConfig `yaml:"state"`
}

// TerraformStateConfig contains Terraform state settings
type TerraformStateConfig struct {
	Backend   string `yaml:"backend" mapstructure:"backend"`
	LocalPath string `yaml:"local_path" mapstructure:"local_path"`

	// S3 backend settings
	S3Bucket string `yaml:"s3_bucket" mapstructure:"s3_bucket"`
	S3Key    string `yaml:"s3_key" mapstructure:"s3_key"`
	S3Region string `yaml:"s3_region" mapstructure:"s3_region"`

	// GCS backend settings
	GCSBucket string `yaml:"gcs_bucket" mapstructure:"gcs_bucket"`
	GCSPrefix string `yaml:"gcs_prefix" mapstructure:"gcs_prefix"`
}

// GCPConfig contains GCP-specific settings
type GCPConfig struct {
	Enabled  bool                 `yaml:"enabled"`
	Projects []string             `yaml:"projects"`
	State    TerraformStateConfig `yaml:"state"`
}

// FalcoConfig contains Falco integration settings
type FalcoConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	Hostname   string `yaml:"hostname" mapstructure:"hostname"`
	Port       uint16 `yaml:"port" mapstructure:"port"`
	CertFile   string `yaml:"cert_file" mapstructure:"cert_file"`
	KeyFile    string `yaml:"key_file" mapstructure:"key_file"`
	CARootFile string `yaml:"ca_root_file" mapstructure:"ca_root_file"`
}

// DriftRule defines a drift detection rule
type DriftRule struct {
	Name              string   `yaml:"name"`
	ResourceTypes     []string `yaml:"resource_types"`
	WatchedAttributes []string `yaml:"watched_attributes"`
	Severity          string   `yaml:"severity"`
}

// NotificationsConfig contains notification channel settings
type NotificationsConfig struct {
	Slack       SlackConfig       `yaml:"slack"`
	Discord     DiscordConfig     `yaml:"discord"`
	FalcoOutput FalcoOutputConfig `yaml:"falco_output"`
	Webhook     WebhookConfig     `yaml:"webhook"`
}

// SlackConfig contains Slack notification settings
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// DiscordConfig contains Discord notification settings
type DiscordConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

// FalcoOutputConfig contains Falco output settings
type FalcoOutputConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Priority string `yaml:"priority"`
}

// WebhookConfig contains generic webhook settings
type WebhookConfig struct {
	Enabled bool              `yaml:"enabled"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// AutoImportConfig contains automatic import settings
type AutoImportConfig struct {
	Enabled          bool     `yaml:"enabled"`
	TerraformDir     string   `yaml:"terraform_dir"`
	OutputDir        string   `yaml:"output_dir"` // Where to save generated .tf files
	AllowedResources []string `yaml:"allowed_resources"`
	RequireApproval  bool     `yaml:"require_approval"`
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("TFDRIFT")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Providers.AWS.Enabled && !c.Providers.GCP.Enabled {
		return fmt.Errorf("at least one provider must be enabled")
	}

	if c.Providers.AWS.Enabled {
		if len(c.Providers.AWS.Regions) == 0 {
			return fmt.Errorf("AWS regions must be specified")
		}
	}

	// Validate Falco configuration
	if !c.Falco.Enabled {
		return fmt.Errorf("falco must be enabled - TFDrift-Falco requires Falco gRPC connection")
	}
	if c.Falco.Hostname == "" {
		return fmt.Errorf("falco hostname must be specified")
	}
	if c.Falco.Port == 0 {
		return fmt.Errorf("falco port must be specified")
	}

	return nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
