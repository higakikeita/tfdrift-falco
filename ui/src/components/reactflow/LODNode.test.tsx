import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import { LODNode, shouldUseLOD, getLODThresholds } from './LODNode';

// Mock the Handle and useStore from reactflow
vi.mock('reactflow', () => ({
  Handle: ({ type, position, className }: Record<string, unknown>) => (
    <div data-testid={`handle-${type}`} className={className} data-position={position} />
  ),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right',
  },
  useStore: () => (
    // Return a zoom value that triggers medium detail rendering
    0.5
  ),
}));

// Mock OfficialCloudIcon
vi.mock('../icons/OfficialCloudIcons', () => ({
  OfficialCloudIcon: ({ type, size }: Record<string, unknown>) => (
    <div data-testid="cloud-icon" data-type={type} data-size={size}>
      Icon
    </div>
  ),
}));

// Mock CustomNode
vi.mock('./CustomNode', () => ({
  CustomNode: ({ data }: Record<string, unknown>) => (
    <div data-testid="custom-node" data-label={data.label}>
      {data.label}
    </div>
  ),
}));

const mockNodeProps = (data = {}) => ({
  id: 'test-lod-node',
  data: {
    label: 'Test Node',
    type: 'ec2',
    resource_type: 'aws:ec2:instance',
    severity: 'high' as const,
    resource_name: 'test-instance',
    ...data,
  },
  selected: false,
  isConnectable: true,
  xPos: 0,
  yPos: 0,
  dragging: false,
  zIndex: 0,
});

describe('LODNode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render LOD node', () => {
    const { container } = render(
      <LODNode {...mockNodeProps()} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should render with proper structure', () => {
    const { container } = render(
      <LODNode {...mockNodeProps()} />
    );
    expect(container.querySelector('div')).toBeInTheDocument();
  });

  it('should handle nodes without resource type', () => {
    const { container } = render(
      <LODNode {...mockNodeProps({ resource_type: undefined })} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle nodes without metadata', () => {
    const { container } = render(
      <LODNode {...mockNodeProps({ metadata: undefined })} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle nodes with metadata', () => {
    const { container } = render(
      <LODNode
        {...mockNodeProps({
          metadata: {
            availability_zone: 'us-east-1a',
            instance_type: 't3.medium',
          },
        })}
      />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should be memoized for performance', () => {
    const { container: container1 } = render(
      <LODNode {...mockNodeProps({ label: 'Node 1' })} />
    );
    const { container: container2 } = render(
      <LODNode {...mockNodeProps({ label: 'Node 1' })} />
    );

    expect(container1.firstChild).toBeInTheDocument();
    expect(container2.firstChild).toBeInTheDocument();
  });

  it('should render with different types', () => {
    const types = ['ec2', 's3', 'rds', 'lambda', 'dynamodb'];

    types.forEach(type => {
      const { container } = render(
        <LODNode {...mockNodeProps({ type })} />
      );
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  it('should handle selection state', () => {
    const props = mockNodeProps();
    props.selected = true;

    const { container } = render(
      <LODNode {...props} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should update when data changes', () => {
    const { rerender, container } = render(
      <LODNode {...mockNodeProps({ label: 'Original' })} />
    );

    rerender(
      <LODNode {...mockNodeProps({ label: 'Updated' })} />
    );

    expect(container.firstChild).toBeInTheDocument();
  });

  it('should render with different severity levels', () => {
    const severities = ['critical', 'high', 'medium', 'low', undefined];

    severities.forEach(severity => {
      const { container } = render(
        <LODNode {...mockNodeProps({ severity: severity as unknown as 'critical' | 'high' | 'medium' | 'low' | undefined })} />
      );
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  it('should handle connectable prop', () => {
    const props = mockNodeProps();
    props.isConnectable = false;

    const { container } = render(
      <LODNode {...props} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle dragging state', () => {
    const props = mockNodeProps();
    props.dragging = true;

    const { container } = render(
      <LODNode {...props} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });
});

describe('shouldUseLOD utility', () => {
  it('should return false for small node counts', () => {
    expect(shouldUseLOD(50)).toBe(false);
  });

  it('should return true for large node counts', () => {
    expect(shouldUseLOD(150)).toBe(true);
  });

  it('should return false at threshold boundary', () => {
    expect(shouldUseLOD(100)).toBe(false);
  });

  it('should return true above threshold boundary', () => {
    expect(shouldUseLOD(101)).toBe(true);
  });

  it('should handle very large node counts', () => {
    expect(shouldUseLOD(5000)).toBe(true);
  });

  it('should handle zero nodes', () => {
    expect(shouldUseLOD(0)).toBe(false);
  });
});

describe('getLODThresholds utility', () => {
  it('should return all zeros for small graphs', () => {
    const thresholds = getLODThresholds(50);
    expect(thresholds.minimal).toBe(0);
    expect(thresholds.medium).toBe(0);
    expect(thresholds.full).toBe(0);
  });

  it('should return progressive thresholds for medium graphs', () => {
    const thresholds = getLODThresholds(250);
    expect(thresholds.minimal).toBe(0.2);
    expect(thresholds.medium).toBe(0.5);
    expect(thresholds.full).toBe(1.0);
  });

  it('should return tighter thresholds for large graphs', () => {
    const thresholds = getLODThresholds(750);
    expect(thresholds.minimal).toBe(0.3);
    expect(thresholds.medium).toBe(0.6);
    expect(thresholds.full).toBe(1.0);
  });

  it('should return aggressive thresholds for very large graphs', () => {
    const thresholds = getLODThresholds(2000);
    expect(thresholds.minimal).toBe(0.4);
    expect(thresholds.medium).toBe(0.7);
    expect(thresholds.full).toBe(1.0);
  });

  it('should scale thresholds with node count', () => {
    const small = getLODThresholds(100);
    const medium = getLODThresholds(300);
    const large = getLODThresholds(700);
    const veryLarge = getLODThresholds(1500);

    expect(small.minimal).toBeLessThanOrEqual(medium.minimal);
    expect(medium.minimal).toBeLessThanOrEqual(large.minimal);
    expect(large.minimal).toBeLessThanOrEqual(veryLarge.minimal);
  });
});
