/**
 * Cytoscape Toolbar Component
 * Provides layout buttons, zoom controls, and display options
 */

import React, { useState, useCallback } from 'react';
import type { Core, LayoutOptions } from 'cytoscape';
import { layoutConfigs } from '../../styles/cytoscapeStyles';
import { AWS_SERVICE_LEGEND, DRIFT_STATUS_LEGEND } from '../../constants/colors';

export type LayoutType = 'fcose' | 'dagre' | 'concentric' | 'cose' | 'grid';

export interface CytoscapeToolbarProps {
  cy: Core | null | undefined;
  currentLayout: LayoutType;
  onLayoutChange: (layout: LayoutType) => void;
  nodeScale: number;
  onNodeScaleChange: (scale: number) => void;
  filterMode: 'all' | 'drift-only' | 'vpc-only';
  onFilterModeChange: (mode: 'all' | 'drift-only' | 'vpc-only') => void;
  showLegend: boolean;
  onShowLegendChange: (show: boolean) => void;
}

export const CytoscapeToolbar: React.FC<CytoscapeToolbarProps> = ({
  cy,
  currentLayout,
  onLayoutChange,
  nodeScale,
  onNodeScaleChange,
  filterMode,
  onFilterModeChange,
  showLegend,
  onShowLegendChange,
}) => {
  const [showOptionsPanel, setShowOptionsPanel] = useState(false);
  const [panelPosition, setPanelPosition] = useState({ x: 0, y: 60 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const panelRef = React.useRef<HTMLDivElement>(null);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (!panelRef.current) return;
    setIsDragging(true);
    const rect = panelRef.current.getBoundingClientRect();
    setDragOffset({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    });
  }, []);

  React.useEffect(() => {
    if (isDragging) {
      const handleMouseMove = (e: MouseEvent) => {
        setPanelPosition({
          x: e.clientX - dragOffset.x,
          y: e.clientY - dragOffset.y,
        });
      };
      const handleMouseUp = () => {
        setIsDragging(false);
      };
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, dragOffset]);

  const handleExport = useCallback(() => {
    if (cy) {
      const png = cy.png({ full: true, scale: 2 });
      if (png) {
        const link = document.createElement('a');
        link.href = png;
        link.download = 'tfdrift-graph.png';
        link.click();
      }
    }
  }, [cy]);

  return (
    <>
      {/* Controls Overlay */}
      <div className="absolute top-4 right-4 flex flex-col gap-2">
        <button
          onClick={() => cy?.fit()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Fit to view"
          aria-label="Fit graph to view"
        >
          🔍 Fit
        </button>
        <button
          onClick={() => cy?.center()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Center view"
          aria-label="Center graph view"
        >
          🎯 Center
        </button>
        <button
          onClick={() => setShowOptionsPanel(!showOptionsPanel)}
          className={`px-3 py-2 rounded-lg shadow-md text-sm font-medium ${
            showOptionsPanel ? 'bg-blue-500 text-white' : 'bg-white hover:bg-gray-100'
          }`}
          title="Show/Hide Options"
          aria-label="Toggle display options panel"
          aria-expanded={showOptionsPanel}
        >
          ⚙️ Options
        </button>
        <button
          onClick={handleExport}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Export as PNG"
          aria-label="Export graph as PNG image"
        >
          📸 Export
        </button>
      </div>

      {/* Options Panel - Draggable */}
      {showOptionsPanel && (
        <div
          ref={panelRef}
          className="fixed bg-white rounded-lg shadow-xl p-4 w-64 z-50 border-2 border-gray-200"
          style={{
            left: `${panelPosition.x}px`,
            top: `${panelPosition.y}px`,
            cursor: isDragging ? 'grabbing' : 'grab',
          }}
        >
          <div
            className="flex items-center justify-between mb-3 pb-2 border-b border-gray-200"
            onMouseDown={handleMouseDown}
          >
            <h3 className="text-sm font-bold text-gray-900 select-none">⚙️ Display Options</h3>
            <button
              onClick={() => setShowOptionsPanel(false)}
              className="text-gray-400 hover:text-gray-600 text-lg leading-none"
              title="Close"
              aria-label="Close options panel"
            >
              ×
            </button>
          </div>

          {/* Filter Mode */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">Filter</label>
            <select
              value={filterMode}
              onChange={(e) => onFilterModeChange(e.target.value as 'all' | 'drift-only' | 'vpc-only')}
              className="w-full px-2 py-1 text-sm border border-gray-300 rounded"
            >
              <option value="all">All Resources</option>
              <option value="drift-only">Drift Only</option>
              <option value="vpc-only">VPC/Network Only</option>
            </select>
          </div>

          {/* Node Scale Slider */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">
              Node Scale: {nodeScale.toFixed(1)}x
            </label>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">小</span>
              <input
                type="range"
                min="0.5"
                max="2.0"
                step="0.1"
                value={nodeScale}
                onChange={(e) => onNodeScaleChange(parseFloat(e.target.value))}
                className="flex-1"
              />
              <span className="text-xs text-gray-500">大</span>
            </div>
            <div className="flex gap-1 mt-2">
              <button
                onClick={() => onNodeScaleChange(0.7)}
                className={`flex-1 px-2 py-1 text-xs rounded ${
                  nodeScale === 0.7 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'
                }`}
              >
                小
              </button>
              <button
                onClick={() => onNodeScaleChange(1.0)}
                className={`flex-1 px-2 py-1 text-xs rounded ${
                  nodeScale === 1.0 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'
                }`}
              >
                標準
              </button>
              <button
                onClick={() => onNodeScaleChange(1.3)}
                className={`flex-1 px-2 py-1 text-xs rounded ${
                  nodeScale === 1.3 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'
                }`}
              >
                大
              </button>
            </div>
          </div>

          {/* Legend Toggle */}
          <div className="mb-4">
            <label className="flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={showLegend}
                onChange={(e) => onShowLegendChange(e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm text-gray-700">Show Legend</span>
            </label>
          </div>

          {/* Layout Selection */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">Layout</label>
            <div className="space-y-1">
              {(['fcose', 'dagre', 'concentric', 'cose', 'grid'] as LayoutType[]).map((layoutType) => (
                <label key={layoutType} className="flex items-center cursor-pointer">
                  <input
                    type="radio"
                    name="layout"
                    value={layoutType}
                    checked={currentLayout === layoutType}
                    onChange={() => {
                      onLayoutChange(layoutType);
                      if (cy) {
                        const layoutConfig = (layoutConfigs as Record<string, LayoutOptions>)[layoutType];
                        cy.layout(layoutConfig).run();
                      }
                    }}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700 capitalize">
                    {layoutType}
                    {layoutType === 'fcose' && ' (推奨)'}
                  </span>
                </label>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Legend */}
      {showLegend && (
        <div className="absolute bottom-4 left-4 bg-white rounded-lg shadow-lg p-4 max-h-96 overflow-y-auto w-64">
          <h3 className="text-sm font-bold mb-2">AWS Services</h3>
          <div className="grid grid-cols-2 gap-1 text-xs">
            {AWS_SERVICE_LEGEND.flatMap((category) =>
              category.items.map((item) => (
                <div key={item.label} className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded" style={{ backgroundColor: item.color }} />
                  <span className="text-[10px]">{item.label}</span>
                </div>
              ))
            )}
          </div>
          <hr className="my-2" />
          <h3 className="text-sm font-bold mb-2">Drift Status</h3>
          <div className="space-y-1 text-xs">
            {DRIFT_STATUS_LEGEND.map((status) => (
              <div key={status.label} className="flex items-center gap-2">
                <div className={`w-4 h-4 rounded border-4 ${status.borderClass}`} />
                <span>{status.label} ({status.description})</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
};
