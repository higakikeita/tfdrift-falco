/**
 * Cytoscape Styles - Re-exports and Combined Configuration
 */

import type { LayoutOptions } from 'cytoscape';
import { nodeStyles } from './nodeStyles';
import { edgeStyles } from './edgeStyles';
import { clusterStyles } from './clusterStyles';

interface CytoscapeStyle {
  selector: string;
  style: Record<string, unknown>;
}

// Combine all stylesheets
export const cytoscapeStylesheet: CytoscapeStyle[] = [
  ...nodeStyles,
  ...edgeStyles,
  ...clusterStyles,
];

// Layout configurations
export const layoutConfigs: Record<string, LayoutOptions> = {
  fcose: {
    name: 'fcose',
    animate: true,
    animationDuration: 1000,
    animationEasing: 'ease-out-cubic',
    fit: true,
    padding: 50,
    nodeSpacing: 20,
    nodeRepulsion: 4500,
    idealEdgeLength: 50,
    edgeElasticity: 0.45,
    nestingFactor: 0.1,
    gravity: 0.25,
    numIter: 2500,
    tile: true,
    tilingPaddingVertical: 15,
    tilingPaddingHorizontal: 15,
    gravityRangeCompound: 2.0,
    gravityCompound: 1.5,
    gravityRange: 4.0,
    initialEnergyOnIncremental: 0.3,
  } as LayoutOptions,

  dagre: {
    name: 'dagre',
    rankDir: 'TB',
    nodeSep: 120,
    rankSep: 180,
    animate: true,
    animationDuration: 500,
    animationEasing: 'ease-out',
    nestingFactor: 1.2,
    edgeSep: 20,
    ranker: 'longest-path',
    fit: true,
    padding: 50,
  } as LayoutOptions,

  concentric: {
    name: 'concentric',
    concentric: (node: unknown): number => {
      const nodeWithData = node as { data: (key: string) => number | undefined };
      return nodeWithData.data('distance') || 1;
    },
    levelWidth: (): number => 2,
    animate: true,
    animationDuration: 500,
    minNodeSpacing: 100,
  } as LayoutOptions,

  cose: {
    name: 'cose',
    nodeRepulsion: 8000,
    idealEdgeLength: 100,
    animate: true,
    animationDuration: 1000,
    nodeOverlap: 20,
    gravity: 0.1,
  } as LayoutOptions,

  grid: {
    name: 'grid',
    rows: undefined,
    cols: undefined,
    animate: true,
    animationDuration: 500,
  } as LayoutOptions,
};

// Core Cytoscape configuration
export const cytoscapeConfig = {
  style: cytoscapeStylesheet,
  minZoom: 0.1,
  maxZoom: 3,
  boxSelectionEnabled: true,
  autounselectify: false,
  autoungrabify: false,
  compound: true,
};

// Re-exports
export { nodeStyles } from './nodeStyles';
export { edgeStyles } from './edgeStyles';
export { clusterStyles } from './clusterStyles';
