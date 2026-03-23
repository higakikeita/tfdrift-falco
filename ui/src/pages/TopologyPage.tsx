import { useState } from 'react';
import { GraphExportButton } from '../components/graph/GraphExportButton';

export function TopologyPage() {
  const [_viewMode] = useState<'graph' | 'table'>('graph');

  return (
    <div className="space-y-4 h-full flex flex-col">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Topology</h1>
        <div className="flex items-center gap-3">
          <GraphExportButton />
          <div className="text-sm text-slate-500 dark:text-slate-400">
            Infrastructure topology view
          </div>
        </div>
      </div>
      <div className="flex-1 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 min-h-[500px] flex items-center justify-center text-slate-400 dark:text-slate-500">
        Infrastructure Topology Graph (Existing CytoscapeGraph component will be integrated here)
      </div>
    </div>
  );
}
