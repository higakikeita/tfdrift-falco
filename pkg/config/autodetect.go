// Package config provides configuration management for TFDrift-Falco.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AutoDetectResult contains the result of Terraform state auto-detection
type AutoDetectResult struct {
	Found         bool
	Backend       string
	LocalPath     string
	S3Bucket      string
	S3Key         string
	S3Region      string
	DetectionPath string
	Message       string
}

// AutoDetectTerraformState automatically detects Terraform state configuration
// from the current directory or specified path.
//
// Detection order:
// 1. Check for .terraform directory (indicates terraform init was run)
// 2. Check for terraform.tfstate file (local backend)
// 3. Check for backend "s3" configuration in .tf files (S3 backend)
func AutoDetectTerraformState(searchPath string) (*AutoDetectResult, error) {
	if searchPath == "" {
		searchPath, _ = os.Getwd()
	}

	result := &AutoDetectResult{
		DetectionPath: searchPath,
	}

	// Check for .terraform directory
	terraformDir := filepath.Join(searchPath, ".terraform")
	if _, err := os.Stat(terraformDir); os.IsNotExist(err) {
		result.Message = fmt.Sprintf("No .terraform directory found in %s\n"+
			"Run 'terraform init' first, or specify state explicitly with --config", searchPath)
		return result, nil
	}

	// Check for local state file
	localStatePath := filepath.Join(searchPath, "terraform.tfstate")
	if _, err := os.Stat(localStatePath); err == nil {
		result.Found = true
		result.Backend = "local"
		result.LocalPath = localStatePath
		result.Message = fmt.Sprintf("✓ Detected local Terraform state: %s", localStatePath)
		return result, nil
	}

	// Check for S3 backend configuration in .tf files
	s3Config, err := detectS3Backend(searchPath)
	if err == nil && s3Config != nil {
		result.Found = true
		result.Backend = "s3"
		result.S3Bucket = s3Config.Bucket
		result.S3Key = s3Config.Key
		result.S3Region = s3Config.Region
		result.Message = fmt.Sprintf("✓ Detected S3 backend: s3://%s/%s (region: %s)",
			s3Config.Bucket, s3Config.Key, s3Config.Region)
		return result, nil
	}

	// No state found
	result.Message = "Terraform initialized but no state found.\n" +
		"This might be a fresh initialization or remote backend.\n" +
		"Specify state explicitly with --config or check your backend configuration."
	return result, nil
}

// S3BackendConfig represents S3 backend configuration extracted from .tf files
type S3BackendConfig struct {
	Bucket string
	Key    string
	Region string
}

// detectS3Backend searches for backend "s3" configuration in .tf files
func detectS3Backend(searchPath string) (*S3BackendConfig, error) {
	// Find all .tf files
	tfFiles, err := filepath.Glob(filepath.Join(searchPath, "*.tf"))
	if err != nil {
		return nil, err
	}

	// Regex patterns to match backend "s3" configuration
	backendPattern := regexp.MustCompile(`backend\s+"s3"\s*\{([^}]+)\}`)
	bucketPattern := regexp.MustCompile(`bucket\s*=\s*"([^"]+)"`)
	keyPattern := regexp.MustCompile(`key\s*=\s*"([^"]+)"`)
	regionPattern := regexp.MustCompile(`region\s*=\s*"([^"]+)"`)

	// Search through all .tf files
	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			continue
		}

		// Find backend "s3" block
		matches := backendPattern.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			backendBlock := matches[1]

			config := &S3BackendConfig{}

			// Extract bucket
			if bucketMatch := bucketPattern.FindStringSubmatch(backendBlock); len(bucketMatch) > 1 {
				config.Bucket = bucketMatch[1]
			}

			// Extract key
			if keyMatch := keyPattern.FindStringSubmatch(backendBlock); len(keyMatch) > 1 {
				config.Key = keyMatch[1]
			}

			// Extract region
			if regionMatch := regionPattern.FindStringSubmatch(backendBlock); len(regionMatch) > 1 {
				config.Region = regionMatch[1]
			}

			// If bucket and key are found, return the configuration
			if config.Bucket != "" && config.Key != "" {
				// Default region if not specified
				if config.Region == "" {
					config.Region = "us-east-1"
				}
				return config, nil
			}
		}
	}

	return nil, fmt.Errorf("no S3 backend configuration found")
}

// CreateAutoConfig creates a Config object from auto-detection results
func CreateAutoConfig(result *AutoDetectResult) (*Config, error) {
	if !result.Found {
		return nil, fmt.Errorf("no Terraform state detected: %s", result.Message)
	}

	cfg := &Config{
		Providers: ProvidersConfig{
			AWS: AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"}, // Default region, can be overridden
				State: TerraformStateConfig{
					Backend: result.Backend,
				},
			},
		},
		Falco: FalcoConfig{
			Enabled:  true,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: NotificationsConfig{
			FalcoOutput: FalcoOutputConfig{
				Enabled:  true,
				Priority: "warning",
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// Configure state based on backend type
	switch result.Backend {
	case "local":
		cfg.Providers.AWS.State.LocalPath = result.LocalPath

	case "s3":
		cfg.Providers.AWS.State.S3Bucket = result.S3Bucket
		cfg.Providers.AWS.State.S3Key = result.S3Key
		cfg.Providers.AWS.State.S3Region = result.S3Region

		// Extract region from S3 configuration if available
		if result.S3Region != "" {
			cfg.Providers.AWS.Regions = []string{result.S3Region}
		}
	}

	return cfg, nil
}

// PrintAutoDetectHelp prints helpful information when auto-detection fails
func PrintAutoDetectHelp(result *AutoDetectResult) {
	fmt.Println("\n" + strings.Repeat("━", 60))
	fmt.Println("🔍 Terraform State Auto-Detection")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	if result.Found {
		fmt.Println("✓ " + result.Message)
		fmt.Println()
		fmt.Println("Auto-detected configuration:")
		fmt.Printf("  Backend: %s\n", result.Backend)
		if result.Backend == "local" {
			fmt.Printf("  Path:    %s\n", result.LocalPath)
		} else if result.Backend == "s3" {
			fmt.Printf("  Bucket:  %s\n", result.S3Bucket)
			fmt.Printf("  Key:     %s\n", result.S3Key)
			fmt.Printf("  Region:  %s\n", result.S3Region)
		}
	} else {
		fmt.Println("❌ " + result.Message)
		fmt.Println()
		fmt.Println("💡 Quick Start Guide:")
		fmt.Println()
		fmt.Println("  1. If you have Terraform initialized in this directory:")
		fmt.Println("     cd path/to/terraform/directory")
		fmt.Println("     tfdrift --auto")
		fmt.Println()
		fmt.Println("  2. If you want to specify state explicitly:")
		fmt.Println("     tfdrift --config config.yaml")
		fmt.Println()
		fmt.Println("  3. Example config.yaml:")
		fmt.Println()
		fmt.Println("     providers:")
		fmt.Println("       aws:")
		fmt.Println("         enabled: true")
		fmt.Println("         regions: [us-east-1]")
		fmt.Println("         state:")
		fmt.Println("           backend: local")
		fmt.Println("           local_path: ./terraform.tfstate")
		fmt.Println()
		fmt.Println("     falco:")
		fmt.Println("       enabled: true")
		fmt.Println("       hostname: localhost")
		fmt.Println("       port: 5060")
		fmt.Println()
	}

	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
}
