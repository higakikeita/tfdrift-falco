package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// CloudTrailEvent represents a simulated CloudTrail event
type CloudTrailEvent struct {
	EventVersion string                 `json:"eventVersion"`
	EventTime    string                 `json:"eventTime"`
	EventSource  string                 `json:"eventSource"`
	EventName    string                 `json:"eventName"`
	AWSRegion    string                 `json:"awsRegion"`
	SourceIPAddr string                 `json:"sourceIPAddress"`
	UserAgent    string                 `json:"userAgent"`
	UserIdentity UserIdentity           `json:"userIdentity"`
	RequestParams map[string]interface{} `json:"requestParameters"`
	ResponseElements map[string]interface{} `json:"responseElements"`
	RequestID    string                 `json:"requestId"`
	EventID      string                 `json:"eventID"`
	EventType    string                 `json:"eventType"`
}

type UserIdentity struct {
	Type        string `json:"type"`
	PrincipalID string `json:"principalId"`
	ARN         string `json:"arn"`
	AccountID   string `json:"accountId"`
	UserName    string `json:"userName"`
}

var (
	// Supported CloudTrail events with weights (frequency)
	eventTemplates = []EventTemplate{
		// EC2 - 30%
		{Name: "ModifyInstanceAttribute", Service: "ec2.amazonaws.com", Weight: 15, Severity: "high"},
		{Name: "ModifyVolume", Service: "ec2.amazonaws.com", Weight: 10, Severity: "medium"},
		{Name: "AuthorizeSecurityGroupIngress", Service: "ec2.amazonaws.com", Weight: 5, Severity: "critical"},

		// IAM - 25%
		{Name: "PutRolePolicy", Service: "iam.amazonaws.com", Weight: 8, Severity: "critical"},
		{Name: "UpdateAssumeRolePolicy", Service: "iam.amazonaws.com", Weight: 7, Severity: "critical"},
		{Name: "CreateUser", Service: "iam.amazonaws.com", Weight: 5, Severity: "high"},
		{Name: "AttachRolePolicy", Service: "iam.amazonaws.com", Weight: 5, Severity: "high"},

		// S3 - 20%
		{Name: "PutBucketPolicy", Service: "s3.amazonaws.com", Weight: 8, Severity: "high"},
		{Name: "PutBucketEncryption", Service: "s3.amazonaws.com", Weight: 6, Severity: "critical"},
		{Name: "PutBucketVersioning", Service: "s3.amazonaws.com", Weight: 4, Severity: "medium"},
		{Name: "PutBucketPublicAccessBlock", Service: "s3.amazonaws.com", Weight: 2, Severity: "critical"},

		// RDS - 15%
		{Name: "ModifyDBInstance", Service: "rds.amazonaws.com", Weight: 10, Severity: "high"},
		{Name: "ModifyDBCluster", Service: "rds.amazonaws.com", Weight: 5, Severity: "high"},

		// Lambda - 10%
		{Name: "UpdateFunctionConfiguration", Service: "lambda.amazonaws.com", Weight: 6, Severity: "medium"},
		{Name: "UpdateFunctionCode", Service: "lambda.amazonaws.com", Weight: 4, Severity: "medium"},
	}

	awsRegions = []string{
		"us-east-1", "us-west-2", "eu-west-1",
		"ap-northeast-1", "ap-southeast-1",
	}

	users = []string{
		"developer-1", "developer-2", "ops-team",
		"automation-bot", "terraform-ci", "manual-admin",
	}
)

type EventTemplate struct {
	Name     string
	Service  string
	Weight   int
	Severity string
}

func main() {
	rate := flag.Int("rate", 100, "Events per minute")
	duration := flag.Duration("duration", 1*time.Hour, "Test duration")
	output := flag.String("output", "/tmp/simulated-cloudtrail-logs", "Output directory")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	rand.Seed(*seed)

	fmt.Printf("CloudTrail Simulator\n")
	fmt.Printf("  Rate: %d events/min\n", *rate)
	fmt.Printf("  Duration: %s\n", *duration)
	fmt.Printf("  Output: %s\n", *output)
	fmt.Println()

	// Create output directory
	if err := os.MkdirAll(*output, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Calculate interval
	interval := time.Minute / time.Duration(*rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	startTime := time.Now()
	endTime := startTime.Add(*duration)

	eventCount := 0
	fileCount := 0
	var currentFile *os.File
	var currentDate string

	fmt.Println("Starting event generation...")

	for now := range ticker.C {
		if now.After(endTime) {
			break
		}

		// Create new file every hour (like CloudTrail)
		dateStr := now.Format("20060102-15")
		if dateStr != currentDate {
			if currentFile != nil {
				currentFile.Close()
			}

			filename := filepath.Join(*output, fmt.Sprintf("cloudtrail-%s.json", dateStr))
			var err error
			currentFile, err = os.Create(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
				os.Exit(1)
			}
			currentDate = dateStr
			fileCount++
			fmt.Printf("Created new log file: %s\n", filename)
		}

		// Generate event
		event := generateEvent(now)

		// Write to file
		data, err := json.Marshal(event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal event: %v\n", err)
			continue
		}

		if _, err := currentFile.Write(append(data, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write event: %v\n", err)
			continue
		}

		eventCount++

		// Progress report every 1000 events
		if eventCount%1000 == 0 {
			elapsed := time.Since(startTime)
			remaining := endTime.Sub(now)
			fmt.Printf("Generated %d events (elapsed: %s, remaining: %s)\n",
				eventCount, elapsed.Round(time.Second), remaining.Round(time.Second))
		}
	}

	if currentFile != nil {
		currentFile.Close()
	}

	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total events: %d\n", eventCount)
	fmt.Printf("  Files created: %d\n", fileCount)
	fmt.Printf("  Actual rate: %.1f events/min\n",
		float64(eventCount)/time.Since(startTime).Minutes())
	fmt.Printf("  Duration: %s\n", time.Since(startTime).Round(time.Second))
}

func generateEvent(timestamp time.Time) CloudTrailEvent {
	template := selectEventTemplate()

	accountID := fmt.Sprintf("%012d", rand.Intn(1000000000000))
	userName := users[rand.Intn(len(users))]
	region := awsRegions[rand.Intn(len(awsRegions))]

	event := CloudTrailEvent{
		EventVersion: "1.08",
		EventTime:    timestamp.Format(time.RFC3339),
		EventSource:  template.Service,
		EventName:    template.Name,
		AWSRegion:    region,
		SourceIPAddr: generateIP(),
		UserAgent:    "aws-cli/2.0.0",
		UserIdentity: UserIdentity{
			Type:        "IAMUser",
			PrincipalID: fmt.Sprintf("AIDA%s", randomString(16)),
			ARN:         fmt.Sprintf("arn:aws:iam::%s:user/%s", accountID, userName),
			AccountID:   accountID,
			UserName:    userName,
		},
		RequestParams:    generateRequestParams(template),
		ResponseElements: generateResponseElements(template),
		RequestID:        fmt.Sprintf("%s-%s", randomString(8), randomString(12)),
		EventID:          randomString(32),
		EventType:        "AwsApiCall",
	}

	return event
}

func selectEventTemplate() EventTemplate {
	// Weighted random selection
	totalWeight := 0
	for _, t := range eventTemplates {
		totalWeight += t.Weight
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for _, t := range eventTemplates {
		cumulative += t.Weight
		if r < cumulative {
			return t
		}
	}

	return eventTemplates[0]
}

func generateRequestParams(template EventTemplate) map[string]interface{} {
	params := make(map[string]interface{})

	switch template.Name {
	case "ModifyInstanceAttribute":
		params["instanceId"] = fmt.Sprintf("i-%s", randomString(17))
		params["disableApiTermination"] = map[string]bool{"value": rand.Intn(2) == 1}

	case "ModifyVolume":
		params["volumeId"] = fmt.Sprintf("vol-%s", randomString(17))
		params["size"] = rand.Intn(1000) + 100

	case "AuthorizeSecurityGroupIngress":
		params["groupId"] = fmt.Sprintf("sg-%s", randomString(17))
		params["ipPermissions"] = []map[string]interface{}{
			{
				"ipProtocol": "tcp",
				"fromPort":   rand.Intn(65535),
				"toPort":     rand.Intn(65535),
				"ipRanges":   []string{"0.0.0.0/0"},
			},
		}

	case "PutRolePolicy":
		params["roleName"] = fmt.Sprintf("role-%s", randomString(8))
		params["policyName"] = fmt.Sprintf("policy-%s", randomString(8))
		params["policyDocument"] = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

	case "UpdateAssumeRolePolicy":
		params["roleName"] = fmt.Sprintf("role-%s", randomString(8))
		params["policyDocument"] = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

	case "CreateUser":
		params["userName"] = fmt.Sprintf("user-%s", randomString(8))

	case "AttachRolePolicy":
		params["roleName"] = fmt.Sprintf("role-%s", randomString(8))
		params["policyArn"] = fmt.Sprintf("arn:aws:iam::aws:policy/%s", randomString(10))

	case "PutBucketPolicy":
		params["bucket"] = fmt.Sprintf("bucket-%s", randomString(16))
		params["policy"] = "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

	case "PutBucketEncryption":
		params["bucket"] = fmt.Sprintf("bucket-%s", randomString(16))
		params["serverSideEncryptionConfiguration"] = map[string]interface{}{
			"rules": []map[string]interface{}{
				{"applyServerSideEncryptionByDefault": map[string]string{"sseAlgorithm": "AES256"}},
			},
		}

	case "PutBucketVersioning":
		params["bucket"] = fmt.Sprintf("bucket-%s", randomString(16))
		params["versioningConfiguration"] = map[string]string{"status": "Enabled"}

	case "ModifyDBInstance":
		params["dBInstanceIdentifier"] = fmt.Sprintf("db-%s", randomString(12))
		params["allocatedStorage"] = rand.Intn(1000) + 20

	case "ModifyDBCluster":
		params["dBClusterIdentifier"] = fmt.Sprintf("cluster-%s", randomString(12))
		params["backupRetentionPeriod"] = rand.Intn(30) + 1

	case "UpdateFunctionConfiguration":
		params["functionName"] = fmt.Sprintf("lambda-%s", randomString(12))
		params["timeout"] = rand.Intn(900) + 3
		params["memorySize"] = []int{128, 256, 512, 1024, 2048}[rand.Intn(5)]

	case "UpdateFunctionCode":
		params["functionName"] = fmt.Sprintf("lambda-%s", randomString(12))
		params["s3Bucket"] = fmt.Sprintf("lambda-code-%s", randomString(8))
		params["s3Key"] = fmt.Sprintf("code-%s.zip", randomString(8))
	}

	return params
}

func generateResponseElements(template EventTemplate) map[string]interface{} {
	// Simple response
	return map[string]interface{}{
		"requestId": randomString(32),
	}
}

func generateIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
