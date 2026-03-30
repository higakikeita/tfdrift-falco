/**
 * Cytoscape Event Handlers
 * Encapsulates all event binding and unbinding logic for graph interactions
 */

import type { Core, NodeSingular, EdgeSingular } from 'cytoscape';

export interface EventHandlerCallbacks {
  onNodeClick?: (nodeId: string, nodeData: unknown) => void;
  onEdgeClick?: (edgeId: string, edgeData: unknown) => void;
}

export class CytoscapeEventHandler {
  private cy: Core;
  private callbacks: EventHandlerCallbacks;

  constructor(cy: Core, callbacks: EventHandlerCallbacks) {
    this.cy = cy;
    this.callbacks = callbacks;
  }

  /**
   * Bind all event handlers
   */
  public bind(): void {
    this.bindNodeTap();
    this.bindEdgeTap();
    this.bindNodeHover();
  }

  /**
   * Unbind all event handlers
   */
  public unbind(): void {
    this.cy.off('tap', 'node');
    this.cy.off('tap', 'edge');
    this.cy.off('mouseover', 'node');
    this.cy.off('mouseout', 'node');
  }

  /**
   * Bind node tap/click event
   */
  private bindNodeTap(): void {
    this.cy.on('tap', 'node', (evt) => {
      const node: NodeSingular = evt.target;
      const nodeId = node.id();

      if (this.callbacks.onNodeClick) {
        this.callbacks.onNodeClick(nodeId, node.data());
      }
    });
  }

  /**
   * Bind edge tap/click event
   */
  private bindEdgeTap(): void {
    this.cy.on('tap', 'edge', (evt) => {
      const edge: EdgeSingular = evt.target;
      const edgeId = edge.id();

      if (this.callbacks.onEdgeClick) {
        this.callbacks.onEdgeClick(edgeId, edge.data());
      }
    });
  }

  /**
   * Bind node hover effects
   */
  private bindNodeHover(): void {
    // Hover in
    this.cy.on('mouseover', 'node', (evt) => {
      const node: NodeSingular = evt.target;
      const connectedEdges = node.connectedEdges();
      connectedEdges.addClass('hover-highlight');
    });

    // Hover out
    this.cy.on('mouseout', 'node', (evt) => {
      const node: NodeSingular = evt.target;
      const connectedEdges = node.connectedEdges();
      connectedEdges.removeClass('hover-highlight');
    });
  }
}
