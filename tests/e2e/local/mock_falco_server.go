//go:build ignore
// +build ignore

// Mock Falco gRPC server for local E2E testing
// This binary runs as a service and sends CloudTrail-like events via gRPC
// Events can be triggered via HTTP endpoints for controlled testing

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockFalcoServer implements the Falco outputs.Service gRPC server
type MockFalcoServer struct {
	outputs.UnimplementedServiceServer
	mu      sync.Mutex
	clients []outputs.Service_SubServer
	events  chan *outputs.Response
}

// NewMockFalcoServer creates a new mock Falco server
func NewMockFalcoServer() *MockFalcoServer {
	return &MockFalcoServer{
		clients: make([]outputs.Service_SubServer, 0),
		events:  make(chan *outputs.Response, 100),
	}
}

// Sub implements the subscription RPC
func (s *MockFalcoServer) Sub(server outputs.Service_SubServer) error {
	s.mu.Lock()
	s.clients = append(s.clients, server)
	s.mu.Unlock()

	log.Infof("Client subscribed. Total clients: %d", len(s.clients))

	// Send events to this client until context is done
	for {
		select {
		case event := <-s.events:
			if err := server.Send(event); err != nil {
				log.Errorf("Failed to send event: %v", err)
				return err
			}
		case <-server.Context().Done():
			log.Info("Client unsubscribed")
			return nil
		}
	}
}

// PublishEvent sends an event to all connected clients
func (s *MockFalcoServer) PublishEvent(event *outputs.Response) {
	select {
	case s.events <- event:
		log.Debugf("Event queued: %s", event.Rule)
	default:
		log.Warn("Event queue full, dropping event")
	}
}

// EventRequest represents an event request body for HTTP API
type EventRequest struct {
	Rule       string                 `json:"rule"`
	Priority   string                 `json:"priority"`
	SourceName string                 `json:"source_name"`
	OutputFields map[string]interface{} `json:"output_fields"`
}

// HTTPServer provides REST endpoints to trigger events
type HTTPServer struct {
	falcoServer *MockFalcoServer
	port        int
}

// NewHTTPServer creates a new HTTP server for event control
func NewHTTPServer(falcoServer *MockFalcoServer, port int) *HTTPServer {
	return &HTTPServer{
		falcoServer: falcoServer,
		port:        port,
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/trigger-event", h.handleTriggerEvent)
	mux.HandleFunc("/trigger-ec2-change", h.handleEC2Change)
	mux.HandleFunc("/trigger-sg-change", h.handleSecurityGroupChange)
	mux.HandleFunc("/trigger-s3-change", h.handleS3Change)
	mux.HandleFunc("/trigger-azure-vm-change", h.handleAzureVMChange)
	mux.HandleFunc("/trigger-azure-nsg-change", h.handleAzureNSGChange)
	mux.HandleFunc("/trigger-azure-storage-change", h.handleAzureStorageChange)

	addr := fmt.Sprintf(":%d", h.port)
	log.Infof("Starting HTTP server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// handleHealth returns health status
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleTriggerEvent handles generic event trigger requests
func (h *HTTPServer) handleTriggerEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	fields, _ := json.Marshal(req.OutputFields)
	event := &outputs.Response{
		Rule:         req.Rule,
		Priority:     req.Priority,
		SourceName:   req.SourceName,
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "event sent"})
}

// handleEC2Change triggers an EC2 instance modification event
func (h *HTTPServer) handleEC2Change(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"source.ip":      "10.0.0.1",
		"user.name":      "admin",
		"evt.action":     "RunInstances",
		"ec2.instance_id": "i-0123456789abcdef0",
		"ec2.region":     "us-east-1",
		"aws.service":    "ec2",
		"evt.type":       "api_call",
		"cloudtrail.eventname": "ModifyInstanceAttribute",
		"cloudtrail.eventID":   "12345-67890-abcde",
		"cloudtrail.sourceIPAddress": "203.0.113.42",
		"cloudtrail.userAgent": "aws-cli/2.0",
		"cloudtrail.requestParameters": map[string]interface{}{
			"instanceId":              "i-0123456789abcdef0",
			"disableApiTermination":   map[string]bool{"value": true},
		},
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "EC2 Instance Modified",
		Priority:     "warning",
		SourceName:   "aws_cloudtrail",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "EC2 event sent"})
}

// handleSecurityGroupChange triggers a security group modification event
func (h *HTTPServer) handleSecurityGroupChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"source.ip":      "10.0.0.1",
		"user.name":      "admin",
		"evt.action":     "AuthorizeSecurityGroupIngress",
		"ec2.sg_id":      "sg-0123456789abcdef0",
		"ec2.region":     "us-east-1",
		"aws.service":    "ec2",
		"evt.type":       "api_call",
		"cloudtrail.eventname": "AuthorizeSecurityGroupIngress",
		"cloudtrail.eventID":   "12345-67890-fghij",
		"cloudtrail.sourceIPAddress": "203.0.113.42",
		"cloudtrail.requestParameters": map[string]interface{}{
			"groupId": "sg-0123456789abcdef0",
			"ipPermissions": []map[string]interface{}{
				{
					"ipProtocol": "tcp",
					"fromPort":   443,
					"toPort":     443,
					"ipRanges": []map[string]interface{}{
						{"cidrIp": "0.0.0.0/0"},
					},
				},
			},
		},
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "Security Group Modified",
		Priority:     "warning",
		SourceName:   "aws_cloudtrail",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Security Group event sent"})
}

// handleS3Change triggers an S3 bucket modification event
func (h *HTTPServer) handleS3Change(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"source.ip":      "10.0.0.1",
		"user.name":      "admin",
		"evt.action":     "PutBucketEncryption",
		"s3.bucket_name": "my-test-bucket",
		"aws.service":    "s3",
		"evt.type":       "api_call",
		"cloudtrail.eventname": "PutBucketEncryption",
		"cloudtrail.eventID":   "12345-67890-klmno",
		"cloudtrail.sourceIPAddress": "203.0.113.42",
		"cloudtrail.requestParameters": map[string]interface{}{
			"bucketName": "my-test-bucket",
			"ServerSideEncryptionConfiguration": map[string]interface{}{
				"Rules": []map[string]interface{}{
					{
						"ApplyServerSideEncryptionByDefault": map[string]interface{}{
							"SSEAlgorithm": "AES256",
						},
					},
				},
			},
		},
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "S3 Bucket Encryption Modified",
		Priority:     "warning",
		SourceName:   "aws_cloudtrail",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "S3 event sent"})
}

// handleAzureVMChange triggers an Azure Virtual Machine modification event
func (h *HTTPServer) handleAzureVMChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"azure.operationName":    "Microsoft.Compute/virtualMachines/write",
		"azure.resourceId":       "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/tfdrift-e2e-rg/providers/Microsoft.Compute/virtualMachines/tfdrift-test-vm",
		"azure.subscriptionId":   "00000000-0000-0000-0000-000000000000",
		"azure.resourceGroup":    "tfdrift-e2e-rg",
		"azure.caller":           "testuser@example.com",
		"azure.resourceLocation": "eastus",
		"azure.status":           "Succeeded",
		"azure.correlationId":    "e2e-test-correlation-001",
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "Azure VM Modified",
		Priority:     "warning",
		SourceName:   "azure_activity",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Azure VM event sent"})
}

// handleAzureNSGChange triggers an Azure Network Security Group modification event
func (h *HTTPServer) handleAzureNSGChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"azure.operationName":    "Microsoft.Network/networkSecurityGroups/write",
		"azure.resourceId":       "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/tfdrift-e2e-rg/providers/Microsoft.Network/networkSecurityGroups/tfdrift-test-nsg",
		"azure.subscriptionId":   "00000000-0000-0000-0000-000000000000",
		"azure.resourceGroup":    "tfdrift-e2e-rg",
		"azure.caller":           "testuser@example.com",
		"azure.resourceLocation": "eastus",
		"azure.status":           "Succeeded",
		"azure.correlationId":    "e2e-test-correlation-002",
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "Azure NSG Modified",
		Priority:     "warning",
		SourceName:   "azure_activity",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Azure NSG event sent"})
}

// handleAzureStorageChange triggers an Azure Storage Account modification event
func (h *HTTPServer) handleAzureStorageChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	outputFields := map[string]interface{}{
		"azure.operationName":    "Microsoft.Storage/storageAccounts/write",
		"azure.resourceId":       "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/tfdrift-e2e-rg/providers/Microsoft.Storage/storageAccounts/tfdrifte2estorage",
		"azure.subscriptionId":   "00000000-0000-0000-0000-000000000000",
		"azure.resourceGroup":    "tfdrift-e2e-rg",
		"azure.caller":           "testuser@example.com",
		"azure.resourceLocation": "eastus",
		"azure.status":           "Succeeded",
		"azure.correlationId":    "e2e-test-correlation-003",
	}

	fields, _ := json.Marshal(outputFields)
	event := &outputs.Response{
		Rule:         "Azure Storage Modified",
		Priority:     "warning",
		SourceName:   "azure_activity",
		OutputFields: string(fields),
		Time:         timestamppb.Now(),
	}

	h.falcoServer.PublishEvent(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Azure Storage event sent"})
}

func main() {
	grpcPort := flag.Int("grpc-port", 5060, "gRPC server port")
	httpPort := flag.Int("http-port", 8081, "HTTP control server port")
	flag.Parse()

	// Setup logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Create Falco mock server
	falcoServer := NewMockFalcoServer()

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			log.Fatalf("Failed to listen on port %d: %v", *grpcPort, err)
		}
		defer lis.Close()

		grpcSrv := grpc.NewServer()
		outputs.RegisterServiceServer(grpcSrv, falcoServer)

		log.Infof("Starting gRPC server on port %d", *grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Start HTTP control server
	httpServer := NewHTTPServer(falcoServer, *httpPort)
	log.Infof("Starting HTTP control server on port %d", *httpPort)
	httpServer.Start()
}
