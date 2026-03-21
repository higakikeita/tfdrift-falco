/**
 * GraphExportButton - Export graph as PNG, SVG, or JSON (#18)
 */

import { useState, useRef, useEffect } from 'react';
import { Download, Image, FileCode, FileJson, ChevronDown } from 'lucide-react';
import { cn } from '../../lib/utils';
import { toast } from '../../stores/toastStore';

interface GraphExportButtonProps {
  /** Cytoscape Core instance ref */
  getCyInstance?: () => unknown | null;
  /** React Flow instance ref for getting nodes/edges */
  getGraphData?: () => { nodes: unknown[]; edges: unknown[] } | null;
  className?: string;
}

export function GraphExportButton({ getCyInstance, getGraphData, className }: GraphExportButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    if (!isOpen) return;
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [isOpen]);

  const downloadBlob = (blob: Blob, filename: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const downloadText = (text: string, filename: string, mimeType: string) => {
    const blob = new Blob([text], { type: mimeType });
    downloadBlob(blob, filename);
  };

  const timestamp = () => new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);

  const exportPNG = () => {
    setIsOpen(false);
    try {
      const cy = getCyInstance?.() as { png?: (opts: Record<string, unknown>) => string } | null;
      if (cy?.png) {
        const dataUrl = cy.png({ full: true, scale: 2, bg: '#ffffff' });
        // Convert data URL to blob
        fetch(dataUrl)
          .then((res) => res.blob())
          .then((blob) => {
            downloadBlob(blob, `tfdrift-graph-${timestamp()}.png`);
            toast.success('Exported PNG', 'Graph saved as PNG image');
          });
      } else {
        toast.error('Export failed', 'Graph instance not available');
      }
    } catch {
      toast.error('Export failed', 'Could not export as PNG');
    }
  };

  const exportSVG = () => {
    setIsOpen(false);
    try {
      const cy = getCyInstance?.() as { svg?: (opts: Record<string, unknown>) => string } | null;
      if (cy?.svg) {
        const svgContent = cy.svg({ full: true, scale: 1, bg: '#ffffff' });
        downloadText(svgContent, `tfdrift-graph-${timestamp()}.svg`, 'image/svg+xml');
        toast.success('Exported SVG', 'Graph saved as SVG vector');
      } else {
        toast.error('Export failed', 'SVG export requires cytoscape-svg extension');
      }
    } catch {
      toast.error('Export failed', 'Could not export as SVG');
    }
  };

  const exportJSON = () => {
    setIsOpen(false);
    try {
      // Try Cytoscape first
      const cy = getCyInstance?.() as { json?: () => unknown } | null;
      if (cy?.json) {
        const json = cy.json();
        downloadText(
          JSON.stringify(json, null, 2),
          `tfdrift-graph-${timestamp()}.json`,
          'application/json'
        );
        toast.success('Exported JSON', 'Graph data saved as JSON');
        return;
      }

      // Try getGraphData callback
      const data = getGraphData?.();
      if (data) {
        downloadText(
          JSON.stringify(data, null, 2),
          `tfdrift-graph-${timestamp()}.json`,
          'application/json'
        );
        toast.success('Exported JSON', 'Graph data saved as JSON');
        return;
      }

      toast.error('Export failed', 'No graph data available');
    } catch {
      toast.error('Export failed', 'Could not export as JSON');
    }
  };

  const options = [
    { label: 'Export as PNG', icon: Image, onClick: exportPNG, desc: 'Raster image (2x)' },
    { label: 'Export as SVG', icon: FileCode, onClick: exportSVG, desc: 'Vector image' },
    { label: 'Export as JSON', icon: FileJson, onClick: exportJSON, desc: 'Graph data' },
  ];

  return (
    <div ref={menuRef} className={cn('relative', className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
      >
        <Download className="h-4 w-4" />
        Export
        <ChevronDown className={cn('h-3 w-3 transition-transform', isOpen && 'rotate-180')} />
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-1 w-52 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg z-50 py-1">
          {options.map((opt) => (
            <button
              key={opt.label}
              onClick={opt.onClick}
              className="w-full flex items-center gap-3 px-3 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
            >
              <opt.icon className="h-4 w-4 text-slate-400" />
              <div className="text-left">
                <div className="font-medium">{opt.label}</div>
                <div className="text-[10px] text-slate-400">{opt.desc}</div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
