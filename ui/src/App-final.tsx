/**
 * TFDrift-Falco - Final Production Version
 *
 * Modern layout with sidebar, HTML icon overlays, and shadcn/ui
 */

import { useState, useEffect } from 'react';
import { ReactFlowProvider } from 'reactflow';
import ReactFlowGraph from './components/reactflow/ReactFlowGraph';
import { Button } from './components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from './components/ui/card';
import {
  generateSampleCausalChain,
  generateComplexSampleGraph,
  generateBlastRadiusGraph
} from './utils/sampleData';
import { OfficialCloudIcon } from './components/icons/OfficialCloudIcons';
import { useGraph, useStats } from './api/hooks';
import { Loader2, AlertCircle } from 'lucide-react';
import { WelcomeModal, shouldShowWelcome } from './components/onboarding/WelcomeModal';
import { HelpOverlay } from './components/onboarding/HelpOverlay';
import { KeyboardShortcutsGuide } from './components/onboarding/KeyboardShortcutsGuide';

type DemoMode = 'api' | 'simple' | 'complex' | 'blast-radius';

function AppFinal() {
  const [demoMode, setDemoMode] = useState<DemoMode>('api');
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);
  const [showWelcome, setShowWelcome] = useState(shouldShowWelcome());
  const [showShortcuts, setShowShortcuts] = useState(false);

  // Filtering state
  const [searchTerm, setSearchTerm] = useState('');
  const [severityFilters, setSeverityFilters] = useState<string[]>([]);
  const [resourceTypeFilters, setResourceTypeFilters] = useState<string[]>([]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      // Ignore if user is typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      switch (e.key.toLowerCase()) {
        case '?':
          setShowShortcuts(true);
          break;
        case 'h':
          // Toggle help - handled by HelpOverlay component
          break;
        case 'escape':
          setShowWelcome(false);
          setShowShortcuts(false);
          break;
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, []);

  // Fetch data from API
  const { data: apiGraphData, isLoading: graphLoading, error: graphError } = useGraph();
  const { data: apiStats, isLoading: statsLoading } = useStats();

  // Get graph data based on mode
  const getGraphData = () => {
    if (demoMode === 'api') {
      if (apiGraphData) {
        return {
          nodes: apiGraphData.nodes || [],
          edges: apiGraphData.edges || [],
        };
      }
      return { nodes: [], edges: [] };
    }

    switch (demoMode) {
      case 'simple':
        return generateSampleCausalChain();
      case 'complex':
        return generateComplexSampleGraph();
      case 'blast-radius':
        return generateBlastRadiusGraph();
      default:
        return generateSampleCausalChain();
    }
  };

  const graphData = getGraphData();

  // Toggle severity filter
  const toggleSeverityFilter = (severity: string) => {
    setSeverityFilters(prev =>
      prev.includes(severity)
        ? prev.filter(s => s !== severity)
        : [...prev, severity]
    );
  };

  // Toggle resource type filter
  const toggleResourceTypeFilter = (type: string) => {
    setResourceTypeFilters(prev =>
      prev.includes(type)
        ? prev.filter(t => t !== type)
        : [...prev, type]
    );
  };

  // Clear all filters
  const clearFilters = () => {
    setSearchTerm('');
    setSeverityFilters([]);
    setResourceTypeFilters([]);
  };

  // Apply filters to graph data
  const filteredGraphData = {
    nodes: graphData.nodes.filter(node => {
      const data = node.data;

      // Search filter
      const matchesSearch = !searchTerm ||
        data.label?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        data.resource_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        data.resource_type?.toLowerCase().includes(searchTerm.toLowerCase());

      // Severity filter
      const matchesSeverity = severityFilters.length === 0 ||
        (data.severity && severityFilters.includes(data.severity));

      // Resource type filter
      const matchesResourceType = resourceTypeFilters.length === 0 ||
        (data.resource_type && resourceTypeFilters.includes(data.resource_type));

      return matchesSearch && matchesSeverity && matchesResourceType;
    }),
    edges: graphData.edges.filter(edge => {
      // Only include edges where both source and target nodes are in filtered nodes
      const filteredNodeIds = new Set(
        graphData.nodes.filter(node => {
          const data = node.data;
          const matchesSearch = !searchTerm ||
            data.label?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            data.resource_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            data.resource_type?.toLowerCase().includes(searchTerm.toLowerCase());
          const matchesSeverity = severityFilters.length === 0 ||
            (data.severity && severityFilters.includes(data.severity));
          const matchesResourceType = resourceTypeFilters.length === 0 ||
            (data.resource_type && resourceTypeFilters.includes(data.resource_type));
          return matchesSearch && matchesSeverity && matchesResourceType;
        }).map(node => node.data.id)
      );

      return filteredNodeIds.has(edge.data.source) && filteredNodeIds.has(edge.data.target);
    })
  };

  const handleNodeClick = (nodeId: string, nodeData: unknown) => {
    console.log('Node clicked:', nodeId, nodeData);
  };

  const handleHighlightPath = () => {
    if (demoMode === 'simple') {
      setHighlightedPath([
        'drift-001',
        'iam-policy-001',
        'iam-role-001',
        'sa-001',
        'pod-001',
        'container-001',
        'falco-001'
      ]);
    }
  };

  return (
    <div className="h-screen w-full flex flex-col bg-background">
      {/* Professional Header */}
      <header className="border-b bg-card shadow-sm">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <OfficialCloudIcon type="terraform_change" size={32} />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight">TFDrift-Falco</h1>
                <p className="text-sm text-muted-foreground">
                  Cloud Infrastructure Security & Drift Analysis
                </p>
              </div>
            </div>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-3 px-4 py-2 bg-slate-50 rounded-lg border">
                <OfficialCloudIcon type="aws_lambda" size={28} />
                <OfficialCloudIcon type="gcp_compute_instance" size={28} />
                <OfficialCloudIcon type="kubernetes_pod" size={28} />
              </div>
              <div className="text-right">
                <p className="text-xs font-medium text-muted-foreground">Core Value</p>
                <p className="text-sm font-semibold">「なぜ」を可視化する</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Sidebar - Controls */}
        <aside className="w-72 border-r bg-card overflow-y-auto">
          <div className="p-4 space-y-4">
            {/* Demo Mode */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Data Source</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant={demoMode === 'api' ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => setDemoMode('api')}
                  disabled={graphLoading}
                >
                  {graphLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Loading API...
                    </>
                  ) : (
                    <>
                      <span className="mr-2">🔌</span> Live API Data
                    </>
                  )}
                </Button>
                <div className="border-t pt-2 mt-2">
                  <p className="text-xs text-muted-foreground mb-2">Demo Scenarios:</p>
                  <Button
                    variant={demoMode === 'simple' ? 'default' : 'outline'}
                    size="sm"
                    className="w-full justify-start mb-1"
                    onClick={() => setDemoMode('simple')}
                  >
                    <span className="mr-2">🔗</span> Simple Chain
                  </Button>
                  <Button
                    variant={demoMode === 'complex' ? 'default' : 'outline'}
                    size="sm"
                    className="w-full justify-start mb-1"
                    onClick={() => setDemoMode('complex')}
                  >
                    <span className="mr-2">🕸️</span> Complex Graph
                  </Button>
                  <Button
                    variant={demoMode === 'blast-radius' ? 'default' : 'outline'}
                    size="sm"
                    className="w-full justify-start"
                    onClick={() => setDemoMode('blast-radius')}
                  >
                    <span className="mr-2">💥</span> Blast Radius
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Filters */}
            <Card>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm">Filters</CardTitle>
                  {(searchTerm || severityFilters.length > 0 || resourceTypeFilters.length > 0) && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs"
                      onClick={clearFilters}
                    >
                      Clear
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Search */}
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                    Search
                  </label>
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Node name, type..."
                    className="w-full px-3 py-1.5 text-sm border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                  />
                </div>

                {/* Severity Filter */}
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                    Severity
                  </label>
                  <div className="space-y-2">
                    {['critical', 'high', 'medium', 'low'].map((severity) => (
                      <label key={severity} className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={severityFilters.includes(severity)}
                          onChange={() => toggleSeverityFilter(severity)}
                          className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <span className="text-sm capitalize">{severity}</span>
                      </label>
                    ))}
                  </div>
                </div>

                {/* Resource Type Filter */}
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
                    Resource Type
                  </label>
                  <div className="space-y-2 max-h-32 overflow-y-auto">
                    {Array.from(new Set(graphData.nodes.map(n => n.data.resource_type)))
                      .filter(Boolean)
                      .sort()
                      .map((type) => (
                        <label key={type} className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={resourceTypeFilters.includes(type)}
                            onChange={() => toggleResourceTypeFilter(type)}
                            className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                          />
                          <span className="text-xs truncate">{type}</span>
                        </label>
                      ))}
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Actions */}
            {demoMode === 'simple' && (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm">Highlight</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <Button
                    size="sm"
                    className="w-full"
                    onClick={handleHighlightPath}
                  >
                    ⚡ Show Path
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full"
                    onClick={() => setHighlightedPath([])}
                  >
                    Clear
                  </Button>
                </CardContent>
              </Card>
            )}

            {/* Stats */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">
                  {demoMode === 'api' ? 'API Statistics' : 'Graph Statistics'}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {demoMode === 'api' && statsLoading ? (
                  <div className="flex items-center justify-center py-4">
                    <Loader2 className="w-4 h-4 text-gray-400 animate-spin" />
                  </div>
                ) : (
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between items-center">
                      <span className="text-muted-foreground">Nodes</span>
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-semibold">{filteredGraphData.nodes.length}</span>
                        {filteredGraphData.nodes.length !== graphData.nodes.length && (
                          <span className="text-xs text-muted-foreground">/ {graphData.nodes.length}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-muted-foreground">Edges</span>
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-semibold">{filteredGraphData.edges.length}</span>
                        {filteredGraphData.edges.length !== graphData.edges.length && (
                          <span className="text-xs text-muted-foreground">/ {graphData.edges.length}</span>
                        )}
                      </div>
                    </div>
                    {demoMode === 'api' && apiStats && (
                      <>
                        <div className="border-t pt-2 mt-2">
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground">Total Drifts</span>
                            <span className="font-mono font-semibold">{apiStats.drifts?.total || 0}</span>
                          </div>
                        </div>
                        {apiStats.drifts?.severity_counts && (
                          <div className="space-y-1 mt-2">
                            {Object.entries(apiStats.drifts.severity_counts).map(([severity, count]) => (
                              <div key={severity} className="flex justify-between items-center text-xs">
                                <span className="text-muted-foreground capitalize">{severity}</span>
                                <span className="font-mono">{count}</span>
                              </div>
                            ))}
                          </div>
                        )}
                      </>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Legend */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Resource Types</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-xs">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-purple-600"></div>
                    <span>Terraform Change</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-red-500"></div>
                    <span>IAM Policy/Role</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-blue-500"></div>
                    <span>Kubernetes</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-cyan-500"></div>
                    <span>Container</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-pink-500"></div>
                    <span>Falco Event</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </aside>

        {/* Main Graph Area */}
        <main className="flex-1 relative bg-gray-50">
          {/* Loading State */}
          {demoMode === 'api' && graphLoading && (
            <div className="absolute inset-0 flex items-center justify-center bg-gray-50 z-10">
              <div className="text-center">
                <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
                <p className="text-sm font-medium text-gray-700">Loading graph data...</p>
                <p className="text-xs text-gray-500 mt-1">Connecting to TFDrift API</p>
              </div>
            </div>
          )}

          {/* Error State */}
          {demoMode === 'api' && graphError && (
            <div className="absolute inset-0 flex items-center justify-center bg-gray-50 z-10">
              <Card className="max-w-md">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-red-600">
                    <AlertCircle className="w-5 h-5" />
                    API Connection Error
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600 mb-4">
                    Unable to connect to the TFDrift API server. Please ensure the backend is running.
                  </p>
                  <div className="bg-red-50 border border-red-200 rounded p-3 mb-4">
                    <p className="text-xs font-mono text-red-800">
                      {graphError instanceof Error ? graphError.message : 'Unknown error'}
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Button
                      size="sm"
                      onClick={() => window.location.reload()}
                      className="w-full"
                    >
                      Retry Connection
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDemoMode('simple')}
                      className="w-full"
                    >
                      Switch to Demo Mode
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {/* Graph */}
          {!(demoMode === 'api' && (graphLoading || graphError)) && (
            <ReactFlowProvider>
              <ReactFlowGraph
                elements={filteredGraphData}
                onNodeClick={handleNodeClick}
                highlightedPath={highlightedPath}
              />
            </ReactFlowProvider>
          )}
        </main>
      </div>

      {/* Welcome Modal */}
      {showWelcome && <WelcomeModal onClose={() => setShowWelcome(false)} />}

      {/* Keyboard Shortcuts Guide */}
      {showShortcuts && <KeyboardShortcutsGuide onClose={() => setShowShortcuts(false)} />}

      {/* Help Overlay */}
      <HelpOverlay
        onOpenShortcuts={() => setShowShortcuts(true)}
        onOpenWelcome={() => setShowWelcome(true)}
      />
    </div>
  );
}

export default AppFinal;
