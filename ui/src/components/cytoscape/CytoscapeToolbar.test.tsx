/**
 * Tests for CytoscapeToolbar component
 */

import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { Core, LayoutOptions } from 'cytoscape';
import { CytoscapeToolbar, type LayoutType } from './CytoscapeToolbar';

// Mock cytoscape layoutConfigs
vi.mock('../../styles/cytoscapeStyles', () => ({
  layoutConfigs: {
    fcose: {},
    dagre: {},
    concentric: {},
    cose: {},
    grid: {},
  },
}));

// Mock colors
vi.mock('../../constants/colors', () => ({
  AWS_SERVICE_LEGEND: [
    {
      items: [
        { label: 'Compute', color: '#FF9900' },
        { label: 'Storage', color: '#569A31' },
      ],
    },
  ],
  DRIFT_STATUS_LEGEND: [
    {
      label: 'In Drift',
      description: 'Resource differs from Terraform',
      borderClass: 'border-red-500',
    },
  ],
}));

describe('CytoscapeToolbar', () => {
  const mockCy = {
    fit: vi.fn(),
    center: vi.fn(),
    png: vi.fn(() => 'data:image/png;base64,fake'),
    zoom: vi.fn(),
    layout: vi.fn(() => ({ run: vi.fn() })),
  } as unknown as Core;

  const defaultProps = {
    cy: mockCy,
    currentLayout: 'dagre' as LayoutType,
    onLayoutChange: vi.fn(),
    nodeScale: 1.0,
    onNodeScaleChange: vi.fn(),
    filterMode: 'all' as const,
    onFilterModeChange: vi.fn(),
    showLegend: true,
    onShowLegendChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders all toolbar buttons', () => {
    render(<CytoscapeToolbar {...defaultProps} />);

    expect(screen.getByRole('button', { name: /Fit/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Center/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Options/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Export/i })).toBeInTheDocument();
  });

  it('fit button calls cy.fit()', async () => {
    const user = userEvent.setup();
    render(<CytoscapeToolbar {...defaultProps} />);

    const fitButton = screen.getByRole('button', { name: /Fit/i });
    await user.click(fitButton);

    expect(mockCy.fit).toHaveBeenCalled();
  });

  it('center button calls cy.center()', async () => {
    const user = userEvent.setup();
    render(<CytoscapeToolbar {...defaultProps} />);

    const centerButton = screen.getByRole('button', { name: /Center/i });
    await user.click(centerButton);

    expect(mockCy.center).toHaveBeenCalled();
  });

  it('layout change calls onLayoutChange', async () => {
    const user = userEvent.setup();
    const onLayoutChange = vi.fn();

    render(
      <CytoscapeToolbar
        {...defaultProps}
        onLayoutChange={onLayoutChange}
      />
    );

    // Open options panel
    const optionsButton = screen.getByRole('button', { name: /Options/i });
    await user.click(optionsButton);

    // Click a layout radio button
    const fcoseLayout = screen.getByRole('radio', { name: /fcose/i });
    await user.click(fcoseLayout);

    expect(onLayoutChange).toHaveBeenCalledWith('fcose');
  });

  it('filter mode change calls onFilterModeChange', async () => {
    const user = userEvent.setup();
    const onFilterModeChange = vi.fn();

    render(
      <CytoscapeToolbar
        {...defaultProps}
        onFilterModeChange={onFilterModeChange}
      />
    );

    // Open options panel
    const optionsButton = screen.getByRole('button', { name: /Options/i });
    await user.click(optionsButton);

    // Find and change the filter select
    const filterSelect = screen.getByDisplayValue('All Resources') as HTMLSelectElement;
    fireEvent.change(filterSelect, { target: { value: 'drift-only' } });

    expect(onFilterModeChange).toHaveBeenCalledWith('drift-only');
  });

  it('node scale slider calls onNodeScaleChange', async () => {
    const user = userEvent.setup();
    const onNodeScaleChange = vi.fn();

    render(
      <CytoscapeToolbar
        {...defaultProps}
        onNodeScaleChange={onNodeScaleChange}
      />
    );

    // Open options panel
    const optionsButton = screen.getByRole('button', { name: /Options/i });
    await user.click(optionsButton);

    // Find and interact with the range slider
    const slider = screen.getByRole('slider') as HTMLInputElement;
    fireEvent.change(slider, { target: { value: '1.5' } });

    expect(onNodeScaleChange).toHaveBeenCalledWith(1.5);
  });

  it('node scale preset buttons call onNodeScaleChange', async () => {
    const user = userEvent.setup();
    const onNodeScaleChange = vi.fn();

    render(
      <CytoscapeToolbar
        {...defaultProps}
        onNodeScaleChange={onNodeScaleChange}
      />
    );

    // Open options panel
    const optionsButton = screen.getByRole('button', { name: /Options/i });
    await user.click(optionsButton);

    // Click small button
    const smallButton = screen.getByRole('button', { name: /小/ });
    await user.click(smallButton);

    expect(onNodeScaleChange).toHaveBeenCalledWith(0.7);
  });

  it('legend checkbox calls onShowLegendChange', async () => {
    const user = userEvent.setup();
    const onShowLegendChange = vi.fn();

    render(
      <CytoscapeToolbar
        {...defaultProps}
        onShowLegendChange={onShowLegendChange}
      />
    );

    // Open options panel
    const optionsButton = screen.getByRole('button', { name: /Options/i });
    await user.click(optionsButton);

    // Find and click the legend checkbox
    const legendCheckbox = screen.getByRole('checkbox', { name: /Show Legend/i });
    await user.click(legendCheckbox);

    expect(onShowLegendChange).toHaveBeenCalledWith(false);
  });

  it('displays legend when showLegend is true', () => {
    render(
      <CytoscapeToolbar
        {...defaultProps}
        showLegend={true}
      />
    );

    expect(screen.getByText('AWS Services')).toBeInTheDocument();
    expect(screen.getByText('Drift Status')).toBeInTheDocument();
  });

  it('does not display legend when showLegend is false', () => {
    render(
      <CytoscapeToolbar
        {...defaultProps}
        showLegend={false}
      />
    );

    expect(screen.queryByText('AWS Services')).not.toBeInTheDocument();
  });

  it('options panel is hidden initially', () => {
    render(<CytoscapeToolbar {...defaultProps} />);

    expect(screen.queryByText('Display Options')).not.toBeInTheDocument();
  });

  it('options button is always visible', async () => {
    render(<CytoscapeToolbar {...defaultProps} />);

    const optionsButton = screen.getByRole('button', { name: /Options/i });
    expect(optionsButton).toBeInTheDocument();
  });

  it('handles null cy gracefully', async () => {
    const user = userEvent.setup();
    render(
      <CytoscapeToolbar
        {...defaultProps}
        cy={null}
      />
    );

    const fitButton = screen.getByRole('button', { name: /Fit/i });
    // Should not throw when clicking fit with null cy
    await user.click(fitButton);

    expect(screen.getByRole('button', { name: /Fit/i })).toBeInTheDocument();
  });

  it('export button calls cy.png()', async () => {
    const user = userEvent.setup();
    render(<CytoscapeToolbar {...defaultProps} />);

    const exportButton = screen.getByRole('button', { name: /Export/i });
    await user.click(exportButton);

    expect(mockCy.png).toHaveBeenCalledWith({ full: true, scale: 2 });
  });
});
