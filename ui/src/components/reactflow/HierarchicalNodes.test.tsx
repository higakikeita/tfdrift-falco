import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import { RegionGroupNode, VPCGroupNode } from './HierarchicalNodes';

// Mock reactflow's NodeProps type
const mockNodeProps = (data = {}) => ({
  id: 'test-node',
  data: {
    label: 'Test Node',
    type: 'region',
    level: 'region' as const,
    ...data,
  },
  selected: false,
  isConnectable: true,
  xPos: 0,
  yPos: 0,
  dragging: false,
  zIndex: 0,
});

describe('RegionGroupNode', () => {
  it('should render region node', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps({ label: 'us-east-1' })} />
    );
    const node = container.querySelector('div');
    expect(node).toBeInTheDocument();
  });

  it('should display region label', () => {
    const { container } = render(
      <RegionGroupNode
        {...mockNodeProps({
          label: 'us-west-2',
        })}
      />
    );
    expect(container.textContent).toContain('us-west-2');
  });

  it('should have AWS orange border color', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps({ label: 'eu-west-1' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ borderColor: '#FF9900' });
  });

  it('should have AWS orange background color', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps({ label: 'ap-south-1' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ backgroundColor: '#FFF8F0' });
  });

  it('should have minimum width of 1950px', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ minWidth: '1950px' });
  });

  it('should have minimum height of 1300px', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ minHeight: '1300px' });
  });

  it('should have proper padding', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('p-8');
  });

  it('should have rounded border', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('rounded-xl');
  });

  it('should display globe emoji', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    expect(container.textContent).toContain('🌎');
  });

  it('should render with shadow', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('shadow-xl');
  });
});

describe('VPCGroupNode', () => {
  it('should render VPC node', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ label: 'prod-vpc', level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toBeInTheDocument();
  });

  it('should display VPC label', () => {
    const { container } = render(
      <VPCGroupNode
        {...mockNodeProps({
          label: 'dev-vpc',
          level: 'vpc',
        })}
      />
    );
    expect(container.textContent).toContain('dev-vpc');
  });

  it('should have AWS VPC blue border color', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ label: 'test-vpc', level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ borderColor: '#147EBA' });
  });

  it('should have AWS VPC blue background color', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ label: 'test-vpc', level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ backgroundColor: '#E6F2F8' });
  });

  it('should have minimum width of 1820px', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ minWidth: '1820px' });
  });

  it('should have minimum height of 1120px', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveStyle({ minHeight: '1120px' });
  });

  it('should have proper padding', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('p-6');
  });

  it('should have rounded border', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('rounded-lg');
  });

  it('should display cloud emoji', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    expect(container.textContent).toContain('☁️');
  });

  it('should render with shadow', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const node = container.querySelector('div');
    expect(node).toHaveClass('shadow-lg');
  });

  it('should display CIDR when metadata provided', () => {
    const { container } = render(
      <VPCGroupNode
        {...mockNodeProps({
          label: 'vpc',
          level: 'vpc',
          metadata: { cidr: '10.0.0.0/16' },
        })}
      />
    );
    expect(container.textContent).toContain('10.0.0.0/16');
  });

  it('should not display CIDR when metadata not provided', () => {
    const { container } = render(
      <VPCGroupNode
        {...mockNodeProps({
          label: 'vpc',
          level: 'vpc',
        })}
      />
    );
    expect(container.textContent).not.toContain('10.0.0.0');
  });

  it('should display CIDR with mono font', () => {
    const { container } = render(
      <VPCGroupNode
        {...mockNodeProps({
          label: 'vpc',
          level: 'vpc',
          metadata: { cidr: '10.0.0.0/16' },
        })}
      />
    );
    const cidrElement = container.querySelector('.font-mono');
    expect(cidrElement).toHaveTextContent('10.0.0.0/16');
  });

  it('should have flex layout for header', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    const header = container.querySelector('.flex');
    expect(header).toHaveClass('items-center');
    expect(header).toHaveClass('justify-between');
  });
});

describe('HierarchicalNodes component composition', () => {
  it('should render RegionGroupNode as a memo component', () => {
    const { container } = render(
      <RegionGroupNode {...mockNodeProps()} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should render VPCGroupNode as a memo component', () => {
    const { container } = render(
      <VPCGroupNode {...mockNodeProps({ level: 'vpc' })} />
    );
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle multiple regions', () => {
    const { container: container1 } = render(
      <RegionGroupNode {...mockNodeProps({ label: 'us-east-1' })} />
    );
    const { container: container2 } = render(
      <RegionGroupNode {...mockNodeProps({ label: 'us-west-2' })} />
    );

    expect(container1.textContent).toContain('us-east-1');
    expect(container2.textContent).toContain('us-west-2');
  });

  it('should handle multiple VPCs', () => {
    const { container: container1 } = render(
      <VPCGroupNode {...mockNodeProps({ label: 'prod-vpc', level: 'vpc' })} />
    );
    const { container: container2 } = render(
      <VPCGroupNode {...mockNodeProps({ label: 'dev-vpc', level: 'vpc' })} />
    );

    expect(container1.textContent).toContain('prod-vpc');
    expect(container2.textContent).toContain('dev-vpc');
  });
});
