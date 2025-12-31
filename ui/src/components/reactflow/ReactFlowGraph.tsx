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
  Panel,
  BackgroundVariant,
  ConnectionMode,
  useReactFlow,
  getRectOfNodes,
  getTransformForBounds
} from 'reactflow';
import type { Node } from 'reactflow';
import { toPng, toSvg } from 'html-to-image';
import 'reactflow/dist/style.css';

import { CustomNode } from './CustomNode';
import { RegionGroupNode, VPCGroupNode, AZGroupNode, SubnetGroupNode } from './HierarchicalNodes';
import { NodeDetailPanel } from './NodeDetailPanel';
import { convertToReactFlow, highlightPath } from '../../utils/reactFlowAdapter';
import type { CytoscapeElements } from '../../types/graph';

interface ReactFlowGraphProps {
  elements: CytoscapeElements;
  layout?: 'dagre' | 'cose' | 'concentric' | 'grid' | 'network-diagram';
  onNodeClick?: (nodeId: string, nodeData: any) => void;
  highlightedPath?: string[];
  highlightedNodes?: string[];
  criticalNodes?: string[];
  className?: string;
}

const nodeTypes = {
  custom: CustomNode,
  'region-group': RegionGroupNode,
  'vpc-group': VPCGroupNode,
  'az-group': AZGroupNode,
  'subnet-group-public': SubnetGroupNode,
  'subnet-group-private': SubnetGroupNode,
};

const defaultEdgeOptions = {
  animated: false,
  style: { stroke: '#64748b', strokeWidth: 2 }
};

const imageWidth = 2400;
const imageHeight = 1600;

export const ReactFlowGraph: React.FC<ReactFlowGraphProps> = ({
  elements,
  layout = 'dagre',
  onNodeClick,
  highlightedPath = [],
  highlightedNodes = [],
  criticalNodes = [],
  className = ''
}) => {
  const { getNodes } = useReactFlow();

  // Convert Cytoscape data to React Flow format
  const { nodes: initialNodes, edges: initialEdges } = useMemo(
    () => convertToReactFlow(elements, layout),
    [elements, layout]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  // Export to PNG
  const handleExportPNG = useCallback(() => {
    const nodesBounds = getRectOfNodes(getNodes());
    const transform = getTransformForBounds(nodesBounds, imageWidth, imageHeight, 0.5, 2);

    const viewport = document.querySelector('.react-flow__viewport') as HTMLElement;
    if (!viewport) return;

    toPng(viewport, {
      backgroundColor: '#f8fafc',
      width: imageWidth,
      height: imageHeight,
      style: {
        width: `${imageWidth}px`,
        height: `${imageHeight}px`,
        transform: `translate(${transform[0]}px, ${transform[1]}px) scale(${transform[2]})`,
      },
      pixelRatio: 2 // High quality
    }).then((dataUrl) => {
      const link = document.createElement('a');
      link.download = `tfdrift-graph-${Date.now()}.png`;
      link.href = dataUrl;
      link.click();
    }).catch((error) => {
      console.error('Error exporting PNG:', error);
      alert('Failed to export PNG. Please try again.');
    });
  }, [getNodes]);

  // Export to SVG
  const handleExportSVG = useCallback(() => {
    const nodesBounds = getRectOfNodes(getNodes());
    const transform = getTransformForBounds(nodesBounds, imageWidth, imageHeight, 0.5, 2);

    const viewport = document.querySelector('.react-flow__viewport') as HTMLElement;
    if (!viewport) return;

    toSvg(viewport, {
      backgroundColor: '#f8fafc',
      width: imageWidth,
      height: imageHeight,
      style: {
        width: `${imageWidth}px`,
        height: `${imageHeight}px`,
        transform: `translate(${transform[0]}px, ${transform[1]}px) scale(${transform[2]})`,
      },
    }).then((dataUrl) => {
      const link = document.createElement('a');
      link.download = `tfdrift-graph-${Date.now()}.svg`;
      link.href = dataUrl;
      link.click();
    }).catch((error) => {
      console.error('Error exporting SVG:', error);
      alert('Failed to export SVG. Please try again.');
    });
  }, [getNodes]);

  // Apply highlighting when path, nodes, or critical nodes change
  useEffect(() => {
    if (highlightedPath.length > 0) {
      const { nodes: pathNodes, edges: pathEdges } = highlightPath(
        initialNodes,
        initialEdges,
        highlightedPath
      );
      setNodes(pathNodes);
      setEdges(pathEdges);
    } else if (highlightedNodes.length > 0) {
      // Highlight nodes with impact radius style
      const updatedNodes = initialNodes.map((node) => {
        const isHighlighted = highlightedNodes.includes(node.id);
        const isCritical = criticalNodes.includes(node.id);
        return {
          ...node,
          style: {
            ...node.style,
            opacity: isHighlighted ? 1 : 0.3,
            border: isHighlighted
              ? '3px solid #dc2626'
              : isCritical
              ? '3px solid #f59e0b'
              : node.style?.border,
            boxShadow: isHighlighted
              ? '0 0 20px rgba(220, 38, 38, 0.5)'
              : isCritical
              ? '0 0 15px rgba(245, 158, 11, 0.4)'
              : node.style?.boxShadow,
          },
        };
      });
      setNodes(updatedNodes);
      setEdges(initialEdges);
    } else if (criticalNodes.length > 0) {
      // Highlight critical nodes only
      const updatedNodes = initialNodes.map((node) => {
        const isCritical = criticalNodes.includes(node.id);
        return {
          ...node,
          style: {
            ...node.style,
            border: isCritical ? '3px solid #f59e0b' : node.style?.border,
            boxShadow: isCritical ? '0 0 15px rgba(245, 158, 11, 0.4)' : node.style?.boxShadow,
            backgroundColor: isCritical ? '#fef3c7' : node.style?.backgroundColor,
          },
        };
      });
      setNodes(updatedNodes);
      setEdges(initialEdges);
    } else {
      setNodes(initialNodes);
      setEdges(initialEdges);
    }
  }, [highlightedPath, highlightedNodes, criticalNodes, initialNodes, initialEdges, setNodes, setEdges]);

  // Handle node click
  const handleNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      setSelectedNode(node);
      if (onNodeClick) {
        onNodeClick(node.id, node.data);
      }
    },
    [onNodeClick]
  );

  return (
    <div className={`w-full h-full ${className} relative`}>
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
            onClick={handleExportPNG}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg shadow-lg font-medium text-sm transition-all hover:shadow-xl flex items-center gap-2"
            title="Export as high-resolution PNG (2400x1600)"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            PNG
          </button>
          <button
            onClick={handleExportSVG}
            className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg shadow-lg font-medium text-sm transition-all hover:shadow-xl flex items-center gap-2"
            title="Export as scalable SVG"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
            SVG
          </button>
        </Panel>
      </ReactFlow>

      {/* Node Detail Panel */}
      <NodeDetailPanel
        node={selectedNode}
        onClose={() => setSelectedNode(null)}
      />
    </div>
  );
};

export default ReactFlowGraph;
