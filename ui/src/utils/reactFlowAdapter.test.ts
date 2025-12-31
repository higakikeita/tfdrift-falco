/**
 * React Flow Adapter Tests
 * Tests for converting Cytoscape data to React Flow format
 */

import { describe, it, expect } from 'vitest';
import { convertToReactFlow, highlightPath } from './reactFlowAdapter';
import type { CytoscapeElements } from '@/types/graph';
import { createMockCytoscapeNode, createMockCytoscapeEdge } from '@/__tests__/utils/testUtils';

describe('reactFlowAdapter', () => {
  describe('convertToReactFlow', () => {
    // Test data setup
    const mockCytoscapeData: CytoscapeElements = {
      nodes: [
        createMockCytoscapeNode({
          data: {
            id: 'node-1',
            label: 'IAM Role',
            type: 'aws_iam_role',
            severity: 'high',
            metadata: {}
          }
        }),
        createMockCytoscapeNode({
          data: {
            id: 'node-2',
            label: 'EC2 Instance',
            type: 'aws_ec2_instance',
            severity: 'medium',
            metadata: {}
          }
        }),
        createMockCytoscapeNode({
          data: {
            id: 'node-3',
            label: 'S3 Bucket',
            type: 'aws_s3_bucket',
            severity: 'low',
            metadata: {}
          }
        })
      ],
      edges: [
        createMockCytoscapeEdge({
          data: {
            id: 'edge-1',
            source: 'node-1',
            target: 'node-2',
            label: 'depends_on'
          }
        }),
        createMockCytoscapeEdge({
          data: {
            id: 'edge-2',
            source: 'node-2',
            target: 'node-3',
            label: 'connects_to'
          }
        })
      ]
    };

    it('should convert Cytoscape data to React Flow format with default dagre layout', () => {
      const result = convertToReactFlow(mockCytoscapeData);

      // Check nodes structure
      expect(result.nodes).toHaveLength(3);
      expect(result.nodes[0]).toMatchObject({
        id: 'node-1',
        type: 'custom',
        data: {
          label: 'IAM Role',
          type: 'aws_iam_role',
          severity: 'high'
        }
      });

      // Check that positions are set (not 0,0 after layout)
      const hasValidPositions = result.nodes.every(
        node => node.position.x !== 0 || node.position.y !== 0
      );
      expect(hasValidPositions).toBe(true);

      // Check edges structure
      expect(result.edges).toHaveLength(2);
      expect(result.edges[0]).toMatchObject({
        id: 'edge-1',
        source: 'node-1',
        target: 'node-2',
        label: 'depends_on',
        type: 'smoothstep'
      });
    });

    it('should apply dagre layout algorithm correctly', () => {
      const result = convertToReactFlow(mockCytoscapeData, 'dagre');

      // Dagre should create a hierarchical top-to-bottom layout
      // The first node should be at the top
      const node1 = result.nodes.find(n => n.id === 'node-1');
      const node2 = result.nodes.find(n => n.id === 'node-2');
      const node3 = result.nodes.find(n => n.id === 'node-3');

      expect(node1).toBeDefined();
      expect(node2).toBeDefined();
      expect(node3).toBeDefined();

      // Check that nodes have positions
      expect(node1!.position).toHaveProperty('x');
      expect(node1!.position).toHaveProperty('y');
    });

    it('should apply force-directed (cose) layout algorithm', () => {
      const result = convertToReactFlow(mockCytoscapeData, 'cose');

      expect(result.nodes).toHaveLength(3);
      expect(result.edges).toHaveLength(2);

      // Force layout should spread nodes apart
      const positions = result.nodes.map(n => n.position);
      const allUnique = positions.every((pos, i) =>
        positions.every((otherPos, j) =>
          i === j || pos.x !== otherPos.x || pos.y !== otherPos.y
        )
      );
      expect(allUnique).toBe(true);
    });

    it('should apply radial (concentric) layout algorithm', () => {
      const result = convertToReactFlow(mockCytoscapeData, 'concentric');

      expect(result.nodes).toHaveLength(3);

      // Radial layout should place the most connected node at center (0,0)
      const centerNode = result.nodes.find(
        n => n.position.x === 0 && n.position.y === 0
      );
      expect(centerNode).toBeDefined();
    });

    it('should apply grid layout algorithm', () => {
      const result = convertToReactFlow(mockCytoscapeData, 'grid');

      expect(result.nodes).toHaveLength(3);

      // Grid layout should arrange nodes in a grid pattern
      // With 3 nodes, we should have a 2x2 grid
      const positions = result.nodes.map(n => n.position);

      // Check that positions follow grid pattern (multiples of spacing)
      const spacing = 300;
      positions.forEach(pos => {
        expect(pos.x % spacing).toBe(0);
        expect(pos.y % spacing).toBe(0);
      });
    });

    it('should apply hierarchical network-diagram layout', () => {
      // Create hierarchical test data
      const hierarchicalData: CytoscapeElements = {
        nodes: [
          createMockCytoscapeNode({
            data: {
              id: 'region-1',
              label: 'us-east-1',
              type: 'aws_region',
              severity: 'low',
              metadata: { hierarchical_level: 'region' }
            }
          }),
          createMockCytoscapeNode({
            data: {
              id: 'vpc-1',
              label: 'VPC',
              type: 'aws_vpc',
              severity: 'low',
              metadata: { hierarchical_level: 'vpc', parent: 'region-1' }
            }
          }),
          createMockCytoscapeNode({
            data: {
              id: 'subnet-1',
              label: 'Subnet',
              type: 'aws_subnet',
              severity: 'low',
              metadata: {
                hierarchical_level: 'subnet',
                parent: 'vpc-1',
                subnet_type: 'public'
              }
            }
          }),
          createMockCytoscapeNode({
            data: {
              id: 'resource-1',
              label: 'EC2',
              type: 'aws_ec2_instance',
              severity: 'medium',
              metadata: { hierarchical_level: 'resource', parent: 'subnet-1' }
            }
          })
        ],
        edges: []
      };

      const result = convertToReactFlow(hierarchicalData, 'network-diagram');

      expect(result.nodes).toHaveLength(4);

      // Check that hierarchical relationships are preserved
      const regionNode = result.nodes.find(n => n.id === 'region-1');
      const vpcNode = result.nodes.find(n => n.id === 'vpc-1');
      const subnetNode = result.nodes.find(n => n.id === 'subnet-1');
      const resourceNode = result.nodes.find(n => n.id === 'resource-1');

      expect(regionNode?.type).toBe('region-group');
      expect(vpcNode?.type).toBe('vpc-group');
      expect(vpcNode?.parentNode).toBe('region-1');
      expect(subnetNode?.type).toBe('subnet-group-public');
      expect(subnetNode?.parentNode).toBe('vpc-1');
      expect(resourceNode?.type).toBe('custom');
      expect(resourceNode?.parentNode).toBe('subnet-1');
    });

    it('should handle empty data', () => {
      const emptyData: CytoscapeElements = {
        nodes: [],
        edges: []
      };

      const result = convertToReactFlow(emptyData);

      expect(result.nodes).toHaveLength(0);
      expect(result.edges).toHaveLength(0);
    });

    it('should preserve node metadata', () => {
      const dataWithMetadata: CytoscapeElements = {
        nodes: [
          createMockCytoscapeNode({
            data: {
              id: 'node-1',
              label: 'Test Node',
              type: 'aws_iam_role',
              severity: 'critical',
              resource_name: 'test-role',
              metadata: {
                arn: 'arn:aws:iam::123456789012:role/test',
                custom_field: 'custom_value'
              }
            }
          })
        ],
        edges: []
      };

      const result = convertToReactFlow(dataWithMetadata);

      expect(result.nodes[0].data.metadata).toEqual({
        arn: 'arn:aws:iam::123456789012:role/test',
        custom_field: 'custom_value'
      });
      expect(result.nodes[0].data.resource_name).toBe('test-role');
    });

    it('should configure edge styling correctly', () => {
      const result = convertToReactFlow(mockCytoscapeData);

      result.edges.forEach(edge => {
        expect(edge.type).toBe('smoothstep');
        expect(edge.animated).toBe(false);
        expect(edge.style).toMatchObject({
          stroke: '#64748b',
          strokeWidth: 2.5
        });
        expect(edge.markerEnd).toMatchObject({
          type: 'arrowclosed',
          color: '#64748b'
        });
      });
    });
  });

  describe('highlightPath', () => {
    const mockNodes = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 0, y: 0 },
        data: { label: 'Node 1' }
      },
      {
        id: 'node-2',
        type: 'custom',
        position: { x: 100, y: 0 },
        data: { label: 'Node 2' }
      },
      {
        id: 'node-3',
        type: 'custom',
        position: { x: 200, y: 0 },
        data: { label: 'Node 3' }
      },
      {
        id: 'node-4',
        type: 'custom',
        position: { x: 300, y: 0 },
        data: { label: 'Node 4' }
      }
    ];

    const mockEdges = [
      {
        id: 'edge-1',
        source: 'node-1',
        target: 'node-2',
        type: 'smoothstep',
        animated: false,
        style: { stroke: '#64748b', strokeWidth: 2 }
      },
      {
        id: 'edge-2',
        source: 'node-2',
        target: 'node-3',
        type: 'smoothstep',
        animated: false,
        style: { stroke: '#64748b', strokeWidth: 2 }
      },
      {
        id: 'edge-3',
        source: 'node-3',
        target: 'node-4',
        type: 'smoothstep',
        animated: false,
        style: { stroke: '#64748b', strokeWidth: 2 }
      }
    ];

    it('should highlight nodes in the specified path', () => {
      const path = ['node-1', 'node-2', 'node-3'];
      const result = highlightPath(mockNodes, mockEdges, path);

      // Check highlighted nodes
      const highlightedNodes = result.nodes.filter(n => n.className === 'highlighted');
      expect(highlightedNodes).toHaveLength(3);
      expect(highlightedNodes.map(n => n.id)).toEqual(path);

      // Check non-highlighted nodes
      const nonHighlightedNodes = result.nodes.filter(n => n.className !== 'highlighted');
      expect(nonHighlightedNodes).toHaveLength(1);
      expect(nonHighlightedNodes[0].id).toBe('node-4');
    });

    it('should apply box shadow to highlighted nodes', () => {
      const path = ['node-1', 'node-2'];
      const result = highlightPath(mockNodes, mockEdges, path);

      result.nodes.forEach(node => {
        if (path.includes(node.id)) {
          expect(node.style).toMatchObject({
            boxShadow: '0 0 20px rgba(59, 130, 246, 0.6)'
          });
        } else {
          expect(node.style).toBeUndefined();
        }
      });
    });

    it('should highlight edges between consecutive path nodes', () => {
      const path = ['node-1', 'node-2', 'node-3'];
      const result = highlightPath(mockNodes, mockEdges, path);

      // edge-1 (node-1 -> node-2) should be highlighted
      const edge1 = result.edges.find(e => e.id === 'edge-1');
      expect(edge1?.animated).toBe(true);
      expect(edge1?.style?.stroke).toBe('#3b82f6');
      expect(edge1?.style?.strokeWidth).toBe(3);

      // edge-2 (node-2 -> node-3) should be highlighted
      const edge2 = result.edges.find(e => e.id === 'edge-2');
      expect(edge2?.animated).toBe(true);
      expect(edge2?.style?.stroke).toBe('#3b82f6');

      // edge-3 (node-3 -> node-4) should NOT be highlighted
      const edge3 = result.edges.find(e => e.id === 'edge-3');
      expect(edge3?.animated).toBe(false);
      expect(edge3?.style?.stroke).toBe('#64748b');
      expect(edge3?.style?.strokeWidth).toBe(2);
    });

    it('should handle empty path', () => {
      const result = highlightPath(mockNodes, mockEdges, []);

      // No nodes should be highlighted
      const highlightedNodes = result.nodes.filter(n => n.className === 'highlighted');
      expect(highlightedNodes).toHaveLength(0);

      // No edges should be animated
      const animatedEdges = result.edges.filter(e => e.animated);
      expect(animatedEdges).toHaveLength(0);
    });

    it('should handle single node path', () => {
      const path = ['node-2'];
      const result = highlightPath(mockNodes, mockEdges, path);

      // One node should be highlighted
      const highlightedNodes = result.nodes.filter(n => n.className === 'highlighted');
      expect(highlightedNodes).toHaveLength(1);
      expect(highlightedNodes[0].id).toBe('node-2');

      // No edges should be highlighted (need at least 2 nodes for an edge)
      const highlightedEdges = result.edges.filter(e => e.animated);
      expect(highlightedEdges).toHaveLength(0);
    });

    it('should not highlight non-consecutive path nodes', () => {
      // Path with gap: node-1 -> node-3 (skipping node-2)
      const path = ['node-1', 'node-3'];
      const result = highlightPath(mockNodes, mockEdges, path);

      // Both nodes should be highlighted
      const highlightedNodes = result.nodes.filter(n => n.className === 'highlighted');
      expect(highlightedNodes).toHaveLength(2);

      // But no edges should be highlighted (they're not consecutive in the path)
      const highlightedEdges = result.edges.filter(e => e.animated);
      expect(highlightedEdges).toHaveLength(0);
    });

    it('should preserve original data when highlighting', () => {
      const path = ['node-1', 'node-2'];
      const result = highlightPath(mockNodes, mockEdges, path);

      // Original node data should be preserved
      expect(result.nodes[0].data).toEqual(mockNodes[0].data);
      expect(result.nodes[0].position).toEqual(mockNodes[0].position);

      // Original edge data should be preserved
      expect(result.edges[0].source).toBe(mockEdges[0].source);
      expect(result.edges[0].target).toBe(mockEdges[0].target);
      expect(result.edges[0].type).toBe(mockEdges[0].type);
    });
  });
});
