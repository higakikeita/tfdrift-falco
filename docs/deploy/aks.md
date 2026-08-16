# Azure Kubernetes Service (AKS) Deployment Guide

This guide covers deploying driftwire on Azure Kubernetes Service (AKS) using Helm charts with Azure-specific integrations.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [AKS Cluster Setup](#aks-cluster-setup)
3. [Helm Chart Deployment](#helm-chart-deployment)
4. [Production Values Configuration](#production-values-configuration)
5. [Azure AD Integration](#azure-ad-integration)
6. [Key Vault Integration](#key-vault-integration)
7. [Application Gateway Ingress](#application-gateway-ingress)
8. [Azure Monitor Integration](#azure-monitor-integration)
9. [Network Configuration](#network-configuration)
10. [Deployment Verification](#deployment-verification)
11. [Scaling and Auto-Scaling](#scaling-and-auto-scaling)
12. [Troubleshooting](#troubleshooting)

## Prerequisites

### Azure Subscription and IAM

- Active Azure subscription with credits
- Required IAM roles:
  - Contributor (on resource group)
  - User Access Administrator (for RBAC)
  - Key Vault Administrator (for secrets)
  - Application Gateway Administrator

### Local Tools

```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Install kubectl
az aks install-cli

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verify installations
az --version
kubectl version --client
helm version
```

### Azure Resources

```bash
# Set defaults
az account set --subscription SUBSCRIPTION_ID
az configure --defaults group=rg-driftwire location=eastus

# Create resource group
az group create --name rg-driftwire --location eastus

# Create container registry
az acr create --resource-group rg-driftwire \
  --name driftwirefalcoacr \
  --sku Basic
```

## AKS Cluster Setup

### Create AKS Cluster

```bash
# Variables
CLUSTER_NAME=driftwire
RESOURCE_GROUP=rg-driftwire
REGISTRY_NAME=driftwirefalcoacr

# Create cluster with Azure AD and other features
az aks create \
  --name $CLUSTER_NAME \
  --resource-group $RESOURCE_GROUP \
  --vm-set-type VirtualMachineScaleSets \
  --node-count 3 \
  --node-vm-size Standard_D2s_v3 \
  --zones 1 2 3 \
  --network-plugin azure \
  --network-policy azure \
  --service-principal-id SP_ID \
  --client-secret SP_SECRET \
  --enable-managed-identity \
  --enable-aad \
  --aad-server-app-id SERVER_APP_ID \
  --aad-server-app-secret SERVER_APP_SECRET \
  --aad-client-app-id CLIENT_APP_ID \
  --enable-azure-keyvault-kms \
  --kms-key-vault-key-id KMS_KEY_ID \
  --kms-key-vault-network-access Public \
  --enable-cluster-autoscaling \
  --min-count 3 \
  --max-count 10 \
  --enable-pod-identity \
  --enable-managed-identity-custom \
  --enable-addons monitoring \
  --workspace-resource-id /subscriptions/SUBSCRIPTION_ID/resourcegroups/rg-driftwire/providers/microsoft.operationalinsights/workspaces/log-driftwire \
  --attach-acr $REGISTRY_NAME \
  --enable-http-application-routing \
  --generate-ssh-keys

# Get cluster credentials
az aks get-credentials --name $CLUSTER_NAME --resource-group $RESOURCE_GROUP
```

### Install Required Add-ons

```bash
# Enable Azure Monitor
az aks enable-addons --addons monitoring \
  --name $CLUSTER_NAME \
  --resource-group $RESOURCE_GROUP

# Enable Application Gateway Ingress Controller
az aks enable-addons --addons ingress-appgw \
  --name $CLUSTER_NAME \
  --resource-group $RESOURCE_GROUP \
  --appgw-id /subscriptions/SUBSCRIPTION_ID/resourceGroups/rg-driftwire/providers/Microsoft.Network/applicationGateways/appgw-driftwire

# Enable Azure Policy
az aks enable-addons --addons azure-policy \
  --name $CLUSTER_NAME \
  --resource-group $RESOURCE_GROUP

# Enable Azure Service Mesh
az aks enable-addons --addons service-mesh \
  --service-mesh-type osm \
  --name $CLUSTER_NAME \
  --resource-group $RESOURCE_GROUP
```

### Verify Cluster Access

```bash
# Check cluster info
kubectl cluster-info

# Check nodes
kubectl get nodes

# Check node pools
az aks nodepool list --cluster-name $CLUSTER_NAME --resource-group $RESOURCE_GROUP
```

### Create Namespace

```bash
# Create namespace
kubectl create namespace driftwire

# Set as default
kubectl config set-context --current --namespace=driftwire

# Verify
kubectl get namespace driftwire
```

## Helm Chart Deployment

### Add Helm Repository (if applicable)

```bash
# If hosting chart on a repository
helm repo add driftwire https://charts.example.com
helm repo update

# Or use local chart
cd /path/to/driftwire
```

### Deploy Using Helm

```bash
# Create values file
cp charts/driftwire/values.yaml values-azure.yaml

# Deploy with Helm
helm install driftwire charts/driftwire \
  --namespace driftwire \
  --values values-azure.yaml

# Check deployment
helm status driftwire --namespace driftwire

# Watch deployment
helm get values driftwire --namespace driftwire
```

## Production Values Configuration

### Azure-Specific values.yaml

```yaml
# Helm values for AKS production deployment

replicaCount: 3

image:
  repository: driftwirefalcoacr.azurecr.io/driftwire
  tag: "1.0.0"
  pullPolicy: IfNotPresent

# Azure Container Registry credentials
imagePullSecrets:
  - name: acr-credentials

serviceAccount:
  create: true
  annotations:
    azure.workload.identity/client-id: "CLIENT_ID"
  name: driftwire

# Pod security context
podSecurityContext:
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

# Pod disruption budget
podDisruptionBudget:
  enabled: true
  minAvailable: 2

# Service configuration
service:
  type: ClusterIP
  port: 8080

# Ingress with Application Gateway
ingress:
  enabled: true
  className: "azure-application-gateway"
  annotations:
    appgw.ingress.kubernetes.io/ssl-redirect: "true"
    appgw.ingress.kubernetes.io/use-private-ip: "false"
    appgw.ingress.kubernetes.io/cookie-based-affinity: "disabled"
    appgw.ingress.kubernetes.io/request-timeout: "30"
  hosts:
    - host: driftwire.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: driftwire-tls
      hosts:
        - driftwire.example.com

# Azure Key Vault integration
keyVault:
  enabled: true
  name: "kv-driftwire"
  tenantId: "TENANT_ID"
  secrets:
    - name: jwt-secret
      objectName: driftwire-jwt-secret
      objectType: secret
    - name: db-password
      objectName: driftwire-db-password
      objectType: secret

# Resource requests and limits
resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Autoscaling
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Node affinity for availability zones
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app
              operator: In
              values:
                - driftwire
        topologyKey: topology.kubernetes.io/zone
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 50
        preference:
          matchExpressions:
            - key: kubernetes.io/os
              operator: In
              values:
                - linux

# Health checks
healthCheck:
  enabled: true
  path: /health
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

# Azure Monitor integration
azureMonitor:
  enabled: true
  metricsEnabled: true
  logsEnabled: true

# Configuration
config:
  port: 8080
  logLevel: info
  logFormat: json

  auth:
    enabled: true
    jwtIssuer: "driftwire"
    jwtExpiry: "24h"

  rateLimit:
    enabled: true
    requestsPerMinute: 1000
    burstSize: 100

  falco:
    hostname: "falco.falco"
    port: 5060

  providers:
    azure:
      enabled: true
      subscriptions:
        - "SUBSCRIPTION_ID"
      state:
        backend: "azurerm"
        resourceGroup: "rg-terraform-state"
        storageAccount: "saterraformstate"
        containerName: "tfstate"
        key: "prod/terraform.tfstate"

# Pod monitoring
podMonitor:
  enabled: true
  interval: 30s
  namespace: driftwire
```

## Azure AD Integration

### Configure Azure AD for Application Authentication

```bash
# Create Azure AD application
az ad app create --display-name driftwire

# Get application ID
APP_ID=$(az ad app list --display-name driftwire --query "[0].appId" -o tsv)

# Create service principal
az ad sp create --id $APP_ID

# Create application password
az ad app credential reset --id $APP_ID

# Grant permissions
az ad app permission add --id $APP_ID \
  --api 00000002-0000-0000-c000-000000000000 \
  --api-permissions 311a71cc-e848-46a1-bdf8-97ff7156d8e6=Scope
```

### Configure OpenID Connect

```yaml
# Add to values.yaml
config:
  auth:
    enabled: true
    jwtIssuer: "https://login.microsoftonline.com/TENANT_ID/v2.0"
    azureAd:
      enabled: true
      clientId: "CLIENT_ID"
      tenantId: "TENANT_ID"
      authority: "https://login.microsoftonline.com/TENANT_ID"
      scopes:
        - "api://CLIENT_ID/.default"
```

## Key Vault Integration

### Set Up Azure Key Vault

```bash
# Create Key Vault
az keyvault create \
  --name kv-driftwire \
  --resource-group rg-driftwire \
  --location eastus \
  --enable-rbac-authorization

# Add secrets
az keyvault secret set \
  --vault-name kv-driftwire \
  --name driftwire-jwt-secret \
  --value "$(openssl rand -base64 32)"

az keyvault secret set \
  --vault-name kv-driftwire \
  --name driftwire-db-password \
  --value "secure-password-here"

# Grant access to managed identity
PRINCIPAL_ID=$(az aks show --name driftwire --resource-group rg-driftwire --query "identity.principalId" -o tsv)

az role assignment create \
  --role "Key Vault Secrets User" \
  --assignee-object-id $PRINCIPAL_ID \
  --scope /subscriptions/SUBSCRIPTION_ID/resourceGroups/rg-driftwire/providers/Microsoft.KeyVault/vaults/kv-driftwire
```

### Access Secrets from Pods

The Azure Workload Identity and Key Vault CSI driver automatically mount secrets:

```bash
# Install Azure Key Vault CSI driver
helm repo add csi-secrets-store-provider-azure https://raw.githubusercontent.com/Azure/secrets-store-csi-driver-provider-azure/master/charts

helm install csi-secrets-store-provider-azure/csi-secrets-store-provider-azure \
  --namespace kube-system
```

## Application Gateway Ingress

### Create Application Gateway

```bash
# Create public IP
az network public-ip create \
  --name pip-appgw-driftwire \
  --resource-group rg-driftwire \
  --sku Standard

# Create Application Gateway
az network application-gateway create \
  --name appgw-driftwire \
  --resource-group rg-driftwire \
  --capacity 2 \
  --sku WAF_v2 \
  --public-ip-address pip-appgw-driftwire \
  --subnet subnet-appgw \
  --cert-password "PASSWORD" \
  --cert-file certificate.pfx \
  --http-settings-cookie-based-affinity Disabled

# Link to AKS
az aks enable-addons \
  --addons ingress-appgw \
  --name driftwire \
  --resource-group rg-driftwire \
  --appgw-id /subscriptions/SUBSCRIPTION_ID/resourceGroups/rg-driftwire/providers/Microsoft.Network/applicationGateways/appgw-driftwire
```

### Configure WAF Rules

```bash
# Enable WAF
az network application-gateway waf-policy create \
  --name waf-driftwire \
  --resource-group rg-driftwire

# Create firewall rule
az network application-gateway waf-policy managed-rules add \
  --policy-name waf-driftwire \
  --resource-group rg-driftwire \
  --type OWASP \
  --version 3.1
```

## Azure Monitor Integration

### Enable Container Insights

Container Insights is automatically enabled when you enable monitoring add-on:

```bash
# Verify it's enabled
az aks show --name driftwire --resource-group rg-driftwire \
  --query addonProfiles.omsagent
```

### Create Alert Rules

```bash
# Create alert for pod CPU
az monitor metrics alert create \
  --name alert-driftwire-pod-cpu \
  --resource-group rg-driftwire \
  --scopes /subscriptions/SUBSCRIPTION_ID/resourcegroups/rg-driftwire/providers/Microsoft.ContainerService/managedClusters/driftwire \
  --condition "avg Percentage CPU > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --action email-action

# Create alert for pod memory
az monitor metrics alert create \
  --name alert-driftwire-pod-memory \
  --resource-group rg-driftwire \
  --scopes /subscriptions/SUBSCRIPTION_ID/resourcegroups/rg-driftwire/providers/Microsoft.ContainerService/managedClusters/driftwire \
  --condition "avg Memory Percentage > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --action email-action
```

### Create Log Analytics Queries

```kusto
// Pod CPU usage
ContainerMetricData
| where TimeGenerated > ago(30m)
| where ContainerName contains "driftwire"
| summarize AvgCpuPercent = avg(CpuPercent) by bin(TimeGenerated, 5m)

// Pod memory usage
ContainerMetricData
| where TimeGenerated > ago(30m)
| where ContainerName contains "driftwire"
| summarize AvgMemoryPercent = avg(MemoryPercent) by bin(TimeGenerated, 5m)

// Container logs with errors
ContainerLog
| where LogEntry contains "error" or LogEntry contains "failed"
| where ContainerName contains "driftwire"
| summarize Count = count() by LogLevel
```

## Network Configuration

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: driftwire-network-policy
  namespace: driftwire
spec:
  podSelector:
    matchLabels:
      app: driftwire
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: ingress-appgw
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: falco
      ports:
        - protocol: TCP
          port: 5060
    - to:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 53
        - protocol: UDP
          port: 53
```

### Network Security Groups

```bash
# Create NSG
az network nsg create \
  --resource-group rg-driftwire \
  --name nsg-driftwire

# Allow ingress traffic
az network nsg rule create \
  --nsg-name nsg-driftwire \
  --resource-group rg-driftwire \
  --name AllowHTTPS \
  --priority 100 \
  --source-address-prefixes '*' \
  --destination-address-prefixes '*' \
  --access Allow \
  --protocol Tcp \
  --direction Inbound \
  --destination-port-ranges 443 80

# Deny all inbound by default
az network nsg rule create \
  --nsg-name nsg-driftwire \
  --resource-group rg-driftwire \
  --name DenyAllInbound \
  --priority 1000 \
  --access Deny \
  --direction Inbound
```

## Deployment Verification

### Check Deployment Status

```bash
# Check all resources
kubectl get all -n driftwire

# Check pod status
kubectl get pods -n driftwire -o wide

# Check deployment
kubectl get deployment -n driftwire

# Check logs
kubectl logs -n driftwire -l app=driftwire --tail=50 -f

# Check events
kubectl get events -n driftwire --sort-by='.lastTimestamp'
```

### Test Connectivity

```bash
# Port forward for testing
kubectl port-forward -n driftwire svc/driftwire 8080:8080

# Test health
curl http://localhost:8080/health

# Test API
curl http://localhost:8080/api/v1/drifts
```

## Scaling and Auto-Scaling

### Horizontal Pod Autoscaling

```bash
# Check HPA
kubectl get hpa -n driftwire

# Describe HPA
kubectl describe hpa driftwire -n driftwire

# Watch HPA
kubectl get hpa -n driftwire --watch
```

### Node Pool Scaling

```bash
# Scale node pool
az aks nodepool scale \
  --cluster-name driftwire \
  --name nodepool1 \
  --resource-group rg-driftwire \
  --node-count 5
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name> -n driftwire

# Check logs
kubectl logs <pod-name> -n driftwire

# Check events
kubectl get events -n driftwire | grep <pod-name>
```

### Key Vault Access Issues

```bash
# Check identity binding
kubectl describe sa driftwire -n driftwire

# Verify Key Vault permissions
az keyvault show --name kv-driftwire

# Test access from pod
kubectl exec -it <pod-name> -n driftwire -- /bin/bash
az keyvault secret show --name driftwire-jwt-secret --vault-name kv-driftwire
```

### Azure Monitor Issues

```bash
# Check monitoring agent
kubectl get daemonset -n kube-system | grep omsagent

# Check logs
kubectl logs -n kube-system -l app=omsagent -c omsagent --tail=50
```

For more information, refer to the [AKS documentation](https://docs.microsoft.com/en-us/azure/aks/).
