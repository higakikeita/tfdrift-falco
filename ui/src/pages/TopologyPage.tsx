import { GraphExportButton } from '../components/graph/GraphExportButton';
import { CytoscapeGraph } from '../components/CytoscapeGraph';
import { useGraph } from '../api/hooks/useGraph';

export function TopologyPage() {
  const { data, isLoading, isError } = useGraph();
  const elements = data ?? { nodes: [], edges: [] };
  const hasGraph = elements.nodes.length > 0;

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

      <div className="flex-1 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 min-h-[500px] overflow-hidden">
        {isLoading ? (
          <div className="h-full flex items-center justify-center text-slate-400 dark:text-slate-500">
            Loading topology…
          </div>
        ) : isError ? (
          <div className="h-full flex items-center justify-center text-slate-400 dark:text-slate-500">
            Failed to load the topology graph.
          </div>
        ) : hasGraph ? (
          <CytoscapeGraph elements={elements} className="w-full h-full" />
        ) : (
          <div className="h-full flex items-center justify-center text-slate-400 dark:text-slate-500">
            No topology yet. Detected drifts and their related resources will appear here as a graph.
          </div>
        )}
      </div>
    </div>
  );
}
