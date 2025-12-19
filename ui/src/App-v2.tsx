/**
 * TFDrift-Falco Graph UI - Redesigned
 *
 * Clean, modern interface with proper shadcn/ui integration
 */

import { useState } from 'react';
import CytoscapeGraph from './components/CytoscapeGraph';
import { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs';
import { Button } from './components/ui/button';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './components/ui/card';
import {
  generateSampleCausalChain,
  generateComplexSampleGraph,
  generateBlastRadiusGraph
} from './utils/sampleData';
import { LayoutType } from './types/graph';
import type { LayoutType as LayoutTypeType } from './types/graph';

type DemoMode = 'simple' | 'complex' | 'blast-radius';

function AppV2() {
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
      {/* Modern Header */}
      <header className="border-b bg-card px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">TFDrift-Falco</h1>
            <p className="text-sm text-muted-foreground">
              Causality Graph Visualization
            </p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-xs font-medium text-muted-foreground">Core Value</p>
              <p className="text-sm font-semibold">「なぜ」を可視化</p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar Controls */}
        <aside className="w-80 border-r bg-card p-4 overflow-y-auto">
          <div className="space-y-6">
            {/* Demo Mode Card */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Demo Mode</CardTitle>
                <CardDescription>Select visualization scenario</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant={demoMode === 'simple' ? 'default' : 'outline'}
                  className="w-full justify-start"
                  onClick={() => setDemoMode('simple')}
                >
                  Simple Chain
                </Button>
                <Button
                  variant={demoMode === 'complex' ? 'default' : 'outline'}
                  className="w-full justify-start"
                  onClick={() => setDemoMode('complex')}
                >
                  Complex Graph
                </Button>
                <Button
                  variant={demoMode === 'blast-radius' ? 'default' : 'outline'}
                  className="w-full justify-start"
                  onClick={() => setDemoMode('blast-radius')}
                >
                  Blast Radius
                </Button>
              </CardContent>
            </Card>

            {/* Layout Card */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Layout</CardTitle>
                <CardDescription>Change graph arrangement</CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <Button
                  variant={layout === LayoutType.HIERARCHICAL ? 'default' : 'outline'}
                  className="w-full justify-start text-sm"
                  onClick={() => setLayout(LayoutType.HIERARCHICAL)}
                >
                  Hierarchical
                </Button>
                <Button
                  variant={layout === LayoutType.RADIAL ? 'default' : 'outline'}
                  className="w-full justify-start text-sm"
                  onClick={() => setLayout(LayoutType.RADIAL)}
                >
                  Radial
                </Button>
                <Button
                  variant={layout === LayoutType.FORCE ? 'default' : 'outline'}
                  className="w-full justify-start text-sm"
                  onClick={() => setLayout(LayoutType.FORCE)}
                >
                  Force-Directed
                </Button>
                <Button
                  variant={layout === LayoutType.GRID ? 'default' : 'outline'}
                  className="w-full justify-start text-sm"
                  onClick={() => setLayout(LayoutType.GRID)}
                >
                  Grid
                </Button>
              </CardContent>
            </Card>

            {/* Actions Card */}
            {demoMode === 'simple' && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Actions</CardTitle>
                  <CardDescription>Highlight causal path</CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  <Button
                    className="w-full"
                    onClick={handleHighlightPath}
                  >
                    Highlight Path
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full"
                    onClick={() => setHighlightedPath([])}
                  >
                    Clear
                  </Button>
                </CardContent>
              </Card>
            )}

            {/* Stats Card */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Graph Stats</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Nodes:</span>
                    <span className="font-semibold">{graphData.nodes.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Edges:</span>
                    <span className="font-semibold">{graphData.edges.length}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </aside>

        {/* Graph Visualization Area */}
        <main className="flex-1 relative bg-muted/20">
          <CytoscapeGraph
            elements={graphData}
            layout={layout}
            onNodeClick={handleNodeClick}
            highlightedPath={highlightedPath}
            className="w-full h-full"
          />
        </main>
      </div>
    </div>
  );
}

export default AppV2;
