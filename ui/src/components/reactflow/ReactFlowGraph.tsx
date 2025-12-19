/**
 * React Flow Graph Component
 * High-quality graph visualization with official cloud provider icons
 */

import { useCallback, useEffect, useMemo, useState } from 'react';
import ReactFlow, {
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  Node,
  Edge,
  Panel,
  BackgroundVariant,
  ConnectionMode
} from 'reactflow';
import 'reactflow/dist/style.css';

import { CustomNode } from './CustomNode';
import { convertToReactFlow, highlightPath } from '../../utils/reactFlowAdapter';
import type { CytoscapeElements } from '../../types/graph';

interface ReactFlowGraphProps {
  elements: CytoscapeElements;
  onNodeClick?: (nodeId: string, nodeData: any) => void;
  highlightedPath?: string[];
  className?: string;
}

const nodeTypes = {
  custom: CustomNode,
};

const defaultEdgeOptions = {
  animated: false,
  style: { stroke: '#64748b', strokeWidth: 2 }
};

export const ReactFlowGraph: React.FC<ReactFlowGraphProps> = ({
  elements,
  onNodeClick,
  highlightedPath = [],
  className = ''
}) => {
  // Convert Cytoscape data to React Flow format
  const { nodes: initialNodes, edges: initialEdges } = useMemo(
    () => convertToReactFlow(elements),
    [elements]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Apply highlighting when path changes
  useEffect(() => {
    if (highlightedPath.length > 0) {
      const { nodes: highlightedNodes, edges: highlightedEdges } = highlightPath(
        initialNodes,
        initialEdges,
        highlightedPath
      );
      setNodes(highlightedNodes);
      setEdges(highlightedEdges);
    } else {
      setNodes(initialNodes);
      setEdges(initialEdges);
    }
  }, [highlightedPath, initialNodes, initialEdges, setNodes, setEdges]);

  // Handle node click
  const handleNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (onNodeClick) {
        onNodeClick(node.id, node.data);
      }
    },
    [onNodeClick]
  );

  return (
    <div className={`w-full h-full ${className}`}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        connectionMode={ConnectionMode.Loose}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={2}
        defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
      >
        {/* Background Grid */}
        <Background
          variant={BackgroundVariant.Dots}
          gap={16}
          size={1}
          color="#cbd5e1"
        />

        {/* Controls */}
        <Controls
          showInteractive={false}
          className="!bg-white !border !border-gray-300 !rounded-lg !shadow-lg"
        />

        {/* Mini Map */}
        <MiniMap
          nodeColor={(node) => {
            const severity = (node.data as any)?.severity;
            switch (severity) {
              case 'critical': return '#ef4444';
              case 'high': return '#f97316';
              case 'medium': return '#eab308';
              case 'low': return '#3b82f6';
              default: return '#64748b';
            }
          }}
          className="!bg-white !border !border-gray-300 !rounded-lg !shadow-lg"
          maskColor="rgba(0, 0, 0, 0.1)"
        />

        {/* Info Panel */}
        <Panel position="top-left" className="bg-white/90 backdrop-blur-sm rounded-lg shadow-lg p-3 border border-gray-200">
          <div className="text-sm font-semibold text-gray-700">
            TFDrift-Falco Graph
          </div>
          <div className="text-xs text-gray-500 mt-1">
            {nodes.length} nodes, {edges.length} edges
          </div>
        </Panel>

        {/* Export Panel */}
        <Panel position="top-right" className="flex gap-2">
          <button
            onClick={() => {
              // TODO: Implement PNG export
              console.log('Export PNG');
            }}
            className="px-3 py-2 bg-white hover:bg-gray-50 rounded-lg shadow-md border border-gray-300 text-sm font-medium text-gray-700 transition-colors"
          >
            Export PNG
          </button>
          <button
            onClick={() => {
              // TODO: Implement SVG export
              console.log('Export SVG');
            }}
            className="px-3 py-2 bg-white hover:bg-gray-50 rounded-lg shadow-md border border-gray-300 text-sm font-medium text-gray-700 transition-colors"
          >
            Export SVG
          </button>
        </Panel>
      </ReactFlow>
    </div>
  );
};

export default ReactFlowGraph;
