import { useState } from 'react';

export function TopologyPage() {
  const [viewMode] = useState<'graph' | 'table'>('graph');

  return (
    <div className="space-y-4 h-full flex flex-col">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Topology</h1>
        <div className="text-sm text-slate-500">
          Existing ReactFlow / Cytoscape graph view
        </div>
      </div>
      <div className="flex-1 bg-white rounded-xl border border-slate-200 min-h-[500px] flex items-center justify-center text-slate-400">
        Infrastructure Topology Graph (Existing CytoscapeGraph component will be integrated here)
      </div>
    </div>
  );
}
