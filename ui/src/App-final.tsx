/**
 * TFDrift-Falco - Final Production Version
 *
 * Modern layout with sidebar, HTML icon overlays, and shadcn/ui
 */

import { useState } from 'react';
import GraphWithIcons from './components/GraphWithIcons';
import { Button } from './components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from './components/ui/card';
import {
  generateSampleCausalChain,
  generateComplexSampleGraph,
  generateBlastRadiusGraph
} from './utils/sampleData';
import { LayoutType } from './types/graph';
import type { LayoutType as LayoutTypeType } from './types/graph';
import { SiTerraform, SiAmazonaws, SiGooglecloud, SiKubernetes } from 'react-icons/si';
import { FaShieldAlt } from 'react-icons/fa';

type DemoMode = 'simple' | 'complex' | 'blast-radius';

function AppFinal() {
  const [demoMode, setDemoMode] = useState<DemoMode>('simple');
  const [layout, setLayout] = useState<LayoutTypeType>(LayoutType.HIERARCHICAL);
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);

  const getGraphData = () => {
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

  const handleNodeClick = (nodeId: string, nodeData: any) => {
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
                <SiTerraform className="text-purple-600 text-2xl" />
                <FaShieldAlt className="text-red-600 text-xl" />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight">TFDrift-Falco</h1>
                <p className="text-sm text-muted-foreground">
                  Causality Graph Visualization
                </p>
              </div>
            </div>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2 text-sm">
                <SiAmazonaws className="text-orange-500 text-lg" />
                <SiGooglecloud className="text-blue-500 text-lg" />
                <SiKubernetes className="text-blue-600 text-lg" />
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
                <CardTitle className="text-sm">Scenario</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant={demoMode === 'simple' ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => setDemoMode('simple')}
                >
                  <span className="mr-2">🔗</span> Simple Chain
                </Button>
                <Button
                  variant={demoMode === 'complex' ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start"
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
              </CardContent>
            </Card>

            {/* Layout */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Layout</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant={layout === LayoutType.HIERARCHICAL ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start text-xs"
                  onClick={() => setLayout(LayoutType.HIERARCHICAL)}
                >
                  Hierarchical
                </Button>
                <Button
                  variant={layout === LayoutType.RADIAL ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start text-xs"
                  onClick={() => setLayout(LayoutType.RADIAL)}
                >
                  Radial
                </Button>
                <Button
                  variant={layout === LayoutType.FORCE ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start text-xs"
                  onClick={() => setLayout(LayoutType.FORCE)}
                >
                  Force-Directed
                </Button>
                <Button
                  variant={layout === LayoutType.GRID ? 'default' : 'outline'}
                  size="sm"
                  className="w-full justify-start text-xs"
                  onClick={() => setLayout(LayoutType.GRID)}
                >
                  Grid
                </Button>
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
                <CardTitle className="text-sm">Graph Statistics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between items-center">
                    <span className="text-muted-foreground">Nodes</span>
                    <span className="font-mono font-semibold">{graphData.nodes.length}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-muted-foreground">Edges</span>
                    <span className="font-mono font-semibold">{graphData.edges.length}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-muted-foreground">Layout</span>
                    <span className="text-xs bg-muted px-2 py-1 rounded">
                      {layout}
                    </span>
                  </div>
                </div>
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
        <main className="flex-1 relative">
          <GraphWithIcons
            elements={graphData}
            layout={layout}
            onNodeClick={handleNodeClick}
            highlightedPath={highlightedPath}
          />
        </main>
      </div>
    </div>
  );
}

export default AppFinal;
