import { useState, useMemo } from 'react';
import { DriftDashboard } from './components/DriftDashboard';
import { CytoscapeGraph } from './components/CytoscapeGraph';
import { useGraph } from './api/hooks/useGraph';
import { useDriftSummary, useDriftDetection } from './api/hooks/useDiscovery';
import type { CytoscapeElements } from './types/graph';
import { mockDriftSummaryWithDrift, mockDriftDetectionWithDrift } from './mocks/driftData';
import { mockMediumGraph } from './mocks/graphData';

type ViewMode = 'drift' | 'graph';

function App() {
  const [viewMode, setViewMode] = useState<ViewMode>('graph');
  const [region] = useState('us-east-1');

  // Toggle between mock data and real API data
  // Mock data provides optimal visualization with 30 nodes
  // Real API data (119 nodes) can be enabled when needed
  const USE_MOCK_GRAPH_DATA = true;
  const USE_MOCK_DRIFT_DATA = true;

  const { data: graphDataFromAPI, isLoading: isGraphLoading } = useGraph();
  const { data: driftSummaryFromAPI } = useDriftSummary(region, { enabled: !USE_MOCK_DRIFT_DATA });
  const { data: driftDetectionFromAPI } = useDriftDetection(region, { enabled: !USE_MOCK_DRIFT_DATA });

  // Use mock data when backend is not available
  const graphData = USE_MOCK_GRAPH_DATA ? mockMediumGraph : graphDataFromAPI;
  const driftSummary = USE_MOCK_DRIFT_DATA ? mockDriftSummaryWithDrift : driftSummaryFromAPI;
  const driftDetection = USE_MOCK_DRIFT_DATA ? mockDriftDetectionWithDrift : driftDetectionFromAPI;

  // Use API graph data directly - it's already in Cytoscape format
  const cytoscapeElements = useMemo(() => {
    if (!graphData) return null;

    // The API already returns data in Cytoscape format
    const elements = graphData as CytoscapeElements;

    // Build a set of resource IDs that have drift
    const driftedResourceIds = new Set<string>();
    const modifiedResourceIds = new Set<string>();
    const missingResourceIds = new Set<string>();

    if (driftDetection?.drift) {
      // Add unmanaged resources (resources in AWS but not in Terraform)
      driftDetection.drift.unmanaged_resources?.forEach(resource => {
        driftedResourceIds.add(resource.id);
        driftedResourceIds.add(resource.name);
        if (resource.arn) driftedResourceIds.add(resource.arn);
      });

      // Add modified resources (resources with configuration differences)
      driftDetection.drift.modified_resources?.forEach(resource => {
        modifiedResourceIds.add(resource.resource_id);
      });

      // Add missing resources (resources in Terraform but not in AWS)
      driftDetection.drift.missing_resources?.forEach(resource => {
        missingResourceIds.add(`${resource.type}.${resource.name}`);
      });
    }

    // Map resource_type to display type for proper styling
    const typeMapping: Record<string, string> = {
      'aws_security_group': 'security_group',
      'aws_iam_role': 'iam_role',
      'aws_iam_policy': 'iam_policy',
      'aws_instance': 'terraform_change',
      'aws_subnet': 'network',
      'aws_vpc': 'network',
      'aws_eks_cluster': 'terraform_change',
      'aws_db_instance': 'terraform_change',
      'aws_lb': 'network',
      'aws_ecs_cluster': 'terraform_change',
      'aws_elasticache_replication_group': 'terraform_change',
      'aws_internet_gateway': 'network',
      'aws_nat_gateway': 'network',
      'aws_route_table': 'network',
    };

    // Build parent-child relationships for VPC/Subnet hierarchy
    // Strategy: Use node metadata (vpc_id, subnet_id attributes) instead of edges
    const vpcNodes = new Map<string, string>(); // vpc_id -> vpc_id (for lookup)
    const subnetNodes = new Map<string, string>(); // subnet_id -> vpc_id
    const resourceToSubnet = new Map<string, string>(); // resource_id -> subnet_id

    // Step 1: Collect ALL VPCs first
    elements.nodes.forEach(node => {
      if (node.data.resource_type === 'aws_vpc') {
        vpcNodes.set(node.data.id, node.data.id);
      }
    });

    // Step 2: Process Subnets and link to VPCs
    elements.nodes.forEach(node => {
      if (node.data.resource_type === 'aws_subnet') {
        // Look for vpc_id in metadata
        const vpcId = node.data.metadata?.attributes?.vpc_id;

        // Debug: Log subnet processing
        if (subnetNodes.size < 3) {
          console.log('🔍 Processing Subnet:', {
            subnetId: node.data.id,
            vpcId,
            hasVpcInMap: vpcId ? vpcNodes.has(vpcId) : false,
            vpcNodesKeys: Array.from(vpcNodes.keys())
          });
        }

        if (vpcId && vpcNodes.has(vpcId)) {
          subnetNodes.set(node.data.id, vpcId);
        }
      }
    });

    // Step 3: Process other resources and link to Subnets
    elements.nodes.forEach(node => {
      if (!['aws_vpc', 'aws_subnet'].includes(node.data.resource_type)) {
        // Look for subnet_id or subnet_ids in metadata
        const metadata = node.data.metadata?.attributes;
        const subnetId = metadata?.subnet_id;
        const subnetIds = metadata?.subnet_ids;

        if (subnetId && subnetNodes.has(subnetId)) {
          resourceToSubnet.set(node.data.id, subnetId);
        } else if (Array.isArray(subnetIds) && subnetIds.length > 0) {
          // If multiple subnets, use the first one as parent
          const firstSubnet = subnetIds.find(sid => subnetNodes.has(sid));
          if (firstSubnet) {
            resourceToSubnet.set(node.data.id, firstSubnet);
          }
        }
      }
    });

    // Transform nodes to use proper types for styling and highlight drifted resources
    const transformedNodes = elements.nodes.map(node => {
      const hasDrift = driftedResourceIds.has(node.data.id) ||
                       driftedResourceIds.has(node.data.label) ||
                       modifiedResourceIds.has(node.data.id) ||
                       missingResourceIds.has(node.data.id);

      const isModified = modifiedResourceIds.has(node.data.id);
      const isMissing = missingResourceIds.has(node.data.id);

      // Determine parent for compound nodes
      // If parent is already set (e.g., from mock data), use it
      // Otherwise, derive it from metadata (for API data)
      let parent = node.data.parent;
      if (!parent) {
        if (node.data.resource_type === 'aws_subnet') {
          parent = subnetNodes.get(node.data.id);
        } else if (!['aws_vpc', 'aws_subnet'].includes(node.data.resource_type)) {
          parent = resourceToSubnet.get(node.data.id);
        }
      }

      return {
        ...node,
        data: {
          ...node.data,
          parent, // Add parent relationship
          type: typeMapping[node.data.resource_type] || 'terraform_resource',
          severity: hasDrift ? (isModified ? 'high' : isMissing ? 'critical' : 'medium') : node.data.severity
        }
      };
    });

    const result = {
      nodes: transformedNodes,
      edges: elements.edges || []
    };

    // Debug: Log parent relationships
    console.log('🔍 VPC/Subnet hierarchy debug:', {
      totalNodes: result.nodes.length,
      vpcCount: result.nodes.filter(n => n.data.resource_type === 'aws_vpc').length,
      subnetCount: result.nodes.filter(n => n.data.resource_type === 'aws_subnet').length,
      nodesWithParent: result.nodes.filter(n => n.data.parent).length,
      vpcNodes: Array.from(vpcNodes.keys()),
      subnetToVpc: Array.from(subnetNodes.entries()),
      resourceToSubnet: Array.from(resourceToSubnet.entries()).slice(0, 10),
      // Sample subnet nodes with their metadata
      sampleSubnets: result.nodes
        .filter(n => n.data.resource_type === 'aws_subnet')
        .slice(0, 2)
        .map(n => ({
          id: n.data.id,
          parent: n.data.parent,
          vpcId: n.data.metadata?.attributes?.vpc_id
        })),
      // Sample resource nodes with subnet info
      sampleResources: result.nodes
        .filter(n => !['aws_vpc', 'aws_subnet'].includes(n.data.resource_type))
        .slice(0, 3)
        .map(n => ({
          id: n.data.id,
          type: n.data.resource_type,
          parent: n.data.parent,
          subnetId: n.data.metadata?.attributes?.subnet_id,
          subnetIds: n.data.metadata?.attributes?.subnet_ids
        }))
    });

    return result;
  }, [graphData, driftDetection]);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">
                TFDrift-Falco
              </h1>
              <span className="ml-3 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
                v0.5.0
              </span>
            </div>

            {/* View Mode Tabs */}
            <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg">
              <button
                onClick={() => setViewMode('drift')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                  viewMode === 'drift'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Drift Detection
              </button>
              <button
                onClick={() => setViewMode('graph')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                  viewMode === 'graph'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Graph View
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {viewMode === 'drift' ? (
          <DriftDashboard region={region} />
        ) : (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">
                  Resource Dependency Graph
                </h2>
                <p className="text-sm text-gray-600 mt-1">
                  {cytoscapeElements && `${cytoscapeElements.nodes.length} resources, ${cytoscapeElements.edges.length} relationships`}
                </p>
              </div>
              {driftSummary && (driftSummary.counts.unmanaged > 0 || driftSummary.counts.modified > 0) && (
                <div className="flex items-center px-3 py-1.5 bg-yellow-100 text-yellow-800 rounded-lg text-sm">
                  <span className="font-medium">⚠️ {driftSummary.counts.unmanaged + driftSummary.counts.modified} resources with drift</span>
                </div>
              )}
            </div>

            {(isGraphLoading && !USE_MOCK_GRAPH_DATA) ? (
              <div className="flex items-center justify-center h-96">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                <span className="ml-3 text-gray-600">Loading graph...</span>
              </div>
            ) : cytoscapeElements && cytoscapeElements.nodes.length > 0 ? (
              <div style={{ height: '600px', width: '100%', border: '1px solid #e5e7eb', borderRadius: '0.5rem' }}>
                <CytoscapeGraph
                  elements={cytoscapeElements}
                  layout="fcose"
                  onNodeClick={(nodeId) => console.log('Node clicked:', nodeId)}
                />
              </div>
            ) : (
              <div className="text-center text-gray-500 py-12 bg-gray-50 rounded-lg">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <p className="mt-4 text-lg font-medium">No graph data available</p>
                <p className="text-sm mt-2">Terraform state must be loaded to visualize resource dependencies</p>
              </div>
            )}
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            TFDrift-Falco - Terraform Drift Detection with Falco Integration
          </p>
        </div>
      </footer>
    </div>
  );
}

export default App;
