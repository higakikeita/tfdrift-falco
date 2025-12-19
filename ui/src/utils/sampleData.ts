/**
 * Sample Data Generator
 *
 * TFDrift-Falco因果関係グラフのデモ用サンプルデータ
 */

import type { CytoscapeElements } from '../types/graph';
import { NodeType, EdgeType } from '../types/graph';

/**
 * 完全な因果関係チェーンのサンプルデータを生成
 *
 * Drift → IAM → ServiceAccount → Pod → Container → Falco Event
 */
export function generateSampleCausalChain(): CytoscapeElements {
  return {
    nodes: [
      // 1. Terraform Change (起点)
      {
        data: {
          id: 'drift-001',
          label: 'IAM Policy\nDrift',
          type: NodeType.TERRAFORM_CHANGE,
          severity: 'critical',
          resource_type: 'aws_iam_policy',
          resource_name: 'eks-admin-policy',
          metadata: {
            change_type: 'modification',
            timestamp: '2025-12-19T17:45:00Z',
            changed_attributes: ['Statement.Action']
          }
        }
      },

      // 2. IAM Policy
      {
        data: {
          id: 'iam-policy-001',
          label: 'eks-admin\nPolicy',
          type: NodeType.IAM_POLICY,
          severity: 'critical',
          resource_type: 'aws_iam_policy',
          resource_name: 'eks-admin-policy',
          metadata: {
            policy_arn: 'arn:aws:iam::123456789012:policy/eks-admin-policy',
            permissions: ['eks:*', 's3:*', 'dynamodb:*']
          }
        }
      },

      // 3. IAM Role
      {
        data: {
          id: 'iam-role-001',
          label: 'app-pod\nRole',
          type: NodeType.IAM_ROLE,
          severity: 'high',
          resource_type: 'aws_iam_role',
          resource_name: 'app-pod-role',
          metadata: {
            role_arn: 'arn:aws:iam::123456789012:role/app-pod-role'
          }
        }
      },

      // 4. Service Account
      {
        data: {
          id: 'sa-001',
          label: 'app-service\nAccount',
          type: NodeType.SERVICE_ACCOUNT,
          severity: 'high',
          resource_type: 'kubernetes_service_account',
          resource_name: 'app-service-account',
          metadata: {
            namespace: 'production',
            annotations: {
              'eks.amazonaws.com/role-arn': 'arn:aws:iam::123456789012:role/app-pod-role'
            }
          }
        }
      },

      // 5. Pod
      {
        data: {
          id: 'pod-001',
          label: 'payment-api\nPod',
          type: NodeType.POD,
          severity: 'medium',
          resource_type: 'kubernetes_pod',
          resource_name: 'payment-api-7d9f8b6c5-xk9tz',
          metadata: {
            namespace: 'production',
            service_account: 'app-service-account',
            labels: {
              app: 'payment-api',
              version: 'v1.2.3'
            }
          }
        }
      },

      // 6. Container
      {
        data: {
          id: 'container-001',
          label: 'payment\nContainer',
          type: NodeType.CONTAINER,
          severity: 'medium',
          resource_type: 'container',
          resource_name: 'payment-api',
          metadata: {
            image: 'payment-api:v1.2.3',
            container_id: 'docker://abc123def456'
          }
        }
      },

      // 7. Falco Event (結果)
      {
        data: {
          id: 'falco-001',
          label: 'Suspicious\nS3 Access',
          type: NodeType.FALCO_EVENT,
          severity: 'critical',
          resource_type: 'falco_alert',
          resource_name: 'Suspicious AWS S3 Access',
          metadata: {
            rule: 'Unauthorized S3 Bucket Access',
            priority: 'Critical',
            output: 'Suspicious S3 GetObject call to sensitive-data-bucket',
            timestamp: '2025-12-19T17:50:00Z'
          }
        }
      }
    ],

    edges: [
      // Drift → IAM Policy
      {
        data: {
          id: 'e1',
          source: 'drift-001',
          target: 'iam-policy-001',
          label: 'caused',
          type: EdgeType.CAUSED_BY,
          relationship: 'configuration_drift'
        }
      },

      // IAM Policy → IAM Role
      {
        data: {
          id: 'e2',
          source: 'iam-policy-001',
          target: 'iam-role-001',
          label: 'attached to',
          type: EdgeType.GRANTS_ACCESS,
          relationship: 'policy_attachment'
        }
      },

      // IAM Role → Service Account
      {
        data: {
          id: 'e3',
          source: 'iam-role-001',
          target: 'sa-001',
          label: 'grants',
          type: EdgeType.GRANTS_ACCESS,
          relationship: 'irsa_binding'
        }
      },

      // Service Account → Pod
      {
        data: {
          id: 'e4',
          source: 'sa-001',
          target: 'pod-001',
          label: 'used by',
          type: EdgeType.USED_BY,
          relationship: 'service_account_mount'
        }
      },

      // Pod → Container
      {
        data: {
          id: 'e5',
          source: 'pod-001',
          target: 'container-001',
          label: 'contains',
          type: EdgeType.CONTAINS,
          relationship: 'pod_container'
        }
      },

      // Container → Falco Event
      {
        data: {
          id: 'e6',
          source: 'container-001',
          target: 'falco-001',
          label: 'triggered',
          type: EdgeType.TRIGGERED,
          relationship: 'runtime_event'
        }
      }
    ]
  };
}

/**
 * より複雑な因果関係グラフ（複数の分岐）
 */
export function generateComplexSampleGraph(): CytoscapeElements {
  const simple = generateSampleCausalChain();

  // 追加のノードとエッジを追加
  const additionalNodes = [
    {
      data: {
        id: 'drift-002',
        label: 'Security Group\nDrift',
        type: NodeType.TERRAFORM_CHANGE,
        severity: 'high' as const,
        resource_type: 'aws_security_group',
        resource_name: 'eks-node-sg',
        metadata: {}
      }
    },
    {
      data: {
        id: 'sg-001',
        label: 'eks-node\nSecurity Group',
        type: NodeType.SECURITY_GROUP,
        severity: 'high' as const,
        resource_type: 'aws_security_group',
        resource_name: 'eks-node-sg',
        metadata: {
          ingress_rules: [
            { from_port: 0, to_port: 65535, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] }
          ]
        }
      }
    },
    {
      data: {
        id: 'pod-002',
        label: 'database\nPod',
        type: NodeType.POD,
        severity: 'medium' as const,
        resource_type: 'kubernetes_pod',
        resource_name: 'postgres-0',
        metadata: {}
      }
    },
    {
      data: {
        id: 'falco-002',
        label: 'Unexpected\nNetwork Access',
        type: NodeType.FALCO_EVENT,
        severity: 'high' as const,
        resource_type: 'falco_alert',
        resource_name: 'Unexpected Network Connection',
        metadata: {}
      }
    }
  ];

  const additionalEdges = [
    {
      data: {
        id: 'e7',
        source: 'drift-002',
        target: 'sg-001',
        label: 'caused',
        type: EdgeType.CAUSED_BY,
        relationship: 'configuration_drift'
      }
    },
    {
      data: {
        id: 'e8',
        source: 'sg-001',
        target: 'pod-002',
        label: 'allows access to',
        type: EdgeType.GRANTS_ACCESS,
        relationship: 'network_policy'
      }
    },
    {
      data: {
        id: 'e9',
        source: 'pod-002',
        target: 'falco-002',
        label: 'triggered',
        type: EdgeType.TRIGGERED,
        relationship: 'runtime_event'
      }
    }
  ];

  return {
    nodes: [...simple.nodes, ...additionalNodes],
    edges: [...simple.edges, ...additionalEdges]
  };
}

/**
 * Blast Radiusデモ用のデータ
 */
export function generateBlastRadiusGraph(): CytoscapeElements {
  const center = {
    data: {
      id: 'iam-policy-blast',
      label: 'Critical\nIAM Policy',
      type: NodeType.IAM_POLICY,
      severity: 'critical' as const,
      resource_type: 'aws_iam_policy',
      resource_name: 'overly-permissive-policy',
      metadata: {}
    }
  };

  const affectedNodes = [
    { id: 'sa-1', label: 'SA 1', type: NodeType.SERVICE_ACCOUNT, severity: undefined },
    { id: 'sa-2', label: 'SA 2', type: NodeType.SERVICE_ACCOUNT, severity: undefined },
    { id: 'sa-3', label: 'SA 3', type: NodeType.SERVICE_ACCOUNT, severity: undefined },
    { id: 'pod-1', label: 'Pod 1', type: NodeType.POD, severity: undefined },
    { id: 'pod-2', label: 'Pod 2', type: NodeType.POD, severity: undefined },
    { id: 'pod-3', label: 'Pod 3', type: NodeType.POD, severity: undefined },
    { id: 'pod-4', label: 'Pod 4', type: NodeType.POD, severity: undefined },
    { id: 'pod-5', label: 'Pod 5', type: NodeType.POD, severity: undefined },
    { id: 'falco-1', label: 'Alert 1', type: NodeType.FALCO_EVENT, severity: 'critical' as const },
    { id: 'falco-2', label: 'Alert 2', type: NodeType.FALCO_EVENT, severity: 'high' as const },
    { id: 'falco-3', label: 'Alert 3', type: NodeType.FALCO_EVENT, severity: 'critical' as const }
  ];

  const edges = [
    // IAM Policy → ServiceAccounts
    { id: 'eb1', source: 'iam-policy-blast', target: 'sa-1', type: EdgeType.GRANTS_ACCESS },
    { id: 'eb2', source: 'iam-policy-blast', target: 'sa-2', type: EdgeType.GRANTS_ACCESS },
    { id: 'eb3', source: 'iam-policy-blast', target: 'sa-3', type: EdgeType.GRANTS_ACCESS },

    // ServiceAccounts → Pods
    { id: 'eb4', source: 'sa-1', target: 'pod-1', type: EdgeType.USED_BY },
    { id: 'eb5', source: 'sa-1', target: 'pod-2', type: EdgeType.USED_BY },
    { id: 'eb6', source: 'sa-2', target: 'pod-3', type: EdgeType.USED_BY },
    { id: 'eb7', source: 'sa-2', target: 'pod-4', type: EdgeType.USED_BY },
    { id: 'eb8', source: 'sa-3', target: 'pod-5', type: EdgeType.USED_BY },

    // Pods → Falco Events
    { id: 'eb9', source: 'pod-1', target: 'falco-1', type: EdgeType.TRIGGERED },
    { id: 'eb10', source: 'pod-3', target: 'falco-2', type: EdgeType.TRIGGERED },
    { id: 'eb11', source: 'pod-4', target: 'falco-3', type: EdgeType.TRIGGERED }
  ];

  return {
    nodes: [
      center,
      ...affectedNodes.map(n => ({
        data: {
          ...n,
          resource_type: n.type,
          resource_name: n.label,
          metadata: {}
        }
      }))
    ],
    edges: edges.map(e => ({
      data: {
        ...e,
        label: '',
        relationship: e.type
      }
    }))
  };
}
