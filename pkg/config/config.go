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
	DryRun        bool                `yaml:"-"`
}

// ProvidersConfig contains cloud provider settings
type ProvidersConfig struct {
	AWS AWSConfig `yaml:"aws"`
	GCP GCPConfig `yaml:"gcp"`
}

// AWSConfig contains AWS-specific settings
type AWSConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Regions    []string          `yaml:"regions"`
	CloudTrail CloudTrailConfig  `yaml:"cloudtrail"`
	State      TerraformStateConfig `yaml:"state"`
}

// CloudTrailConfig contains CloudTrail settings
type CloudTrailConfig struct {
	S3Bucket string `yaml:"s3_bucket"`
	SQSQueue string `yaml:"sqs_queue"`
}

// TerraformStateConfig contains Terraform state settings
type TerraformStateConfig struct {
	Backend  string `yaml:"backend"`
	S3Bucket string `yaml:"s3_bucket"`
	S3Key    string `yaml:"s3_key"`
	LocalPath string `yaml:"local_path"`
}

// GCPConfig contains GCP-specific settings
type GCPConfig struct {
	Enabled bool     `yaml:"enabled"`
	Projects []string `yaml:"projects"`
}

// FalcoConfig contains Falco integration settings
type FalcoConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Socket       string `yaml:"socket"`
	GRPCEndpoint string `yaml:"grpc_endpoint"`
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
	Slack        SlackConfig   `yaml:"slack"`
	Discord      DiscordConfig `yaml:"discord"`
	FalcoOutput  FalcoOutputConfig `yaml:"falco_output"`
	Webhook      WebhookConfig `yaml:"webhook"`
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

	return nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
