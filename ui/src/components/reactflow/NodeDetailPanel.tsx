/**
 * Node Detail Panel Component
 * Displays detailed information about selected nodes
 */

import { memo } from 'react';
import { OfficialCloudIcon } from '../icons/OfficialCloudIcons';

interface NodeDetailPanelProps {
  node: {
    id: string;
    data: {
      label: string;
      type: string;
      resource_type: string;
      severity?: string;
      resource_name?: string;
      metadata?: Record<string, any>;
    };
  } | null;
  onClose: () => void;
}

const getSeverityBadgeStyle = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return 'bg-red-100 text-red-800 border-red-300';
    case 'high':
      return 'bg-orange-100 text-orange-800 border-orange-300';
    case 'medium':
      return 'bg-yellow-100 text-yellow-800 border-yellow-300';
    case 'low':
      return 'bg-blue-100 text-blue-800 border-blue-300';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-300';
  }
};

export const NodeDetailPanel = memo(({ node, onClose }: NodeDetailPanelProps) => {
  if (!node) return null;

  const { data } = node;

  return (
    <div className="absolute right-6 top-24 w-96 bg-white rounded-2xl shadow-2xl border-2 border-gray-200 z-50 overflow-hidden animate-slide-in">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-700 p-5 text-white">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className="p-2 bg-white rounded-xl">
              <OfficialCloudIcon type={data.resource_type} size={48} />
            </div>
            <div>
              <h3 className="font-bold text-lg leading-tight">{data.label}</h3>
              <p className="text-blue-100 text-sm mt-1">Resource Details</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-white hover:bg-white/20 rounded-lg p-1.5 transition-colors"
            aria-label="Close"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="p-5 space-y-4 max-h-[500px] overflow-y-auto">
        {/* Severity */}
        {data.severity && (
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Severity</label>
            <div className="mt-1.5">
              <span className={`
                inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-bold border-2
                ${getSeverityBadgeStyle(data.severity)}
              `}>
                {data.severity.toUpperCase()}
              </span>
            </div>
          </div>
        )}

        {/* Resource Type */}
        <div>
          <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Resource Type</label>
          <div className="mt-1.5 px-4 py-3 bg-gray-50 rounded-lg border border-gray-200">
            <code className="text-sm font-mono text-gray-900">{data.resource_type}</code>
          </div>
        </div>

        {/* Resource Name */}
        {data.resource_name && (
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Resource Name</label>
            <div className="mt-1.5 px-4 py-3 bg-gray-50 rounded-lg border border-gray-200">
              <code className="text-sm font-mono text-gray-900 break-all">{data.resource_name}</code>
            </div>
          </div>
        )}

        {/* Node ID */}
        <div>
          <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Node ID</label>
          <div className="mt-1.5 px-4 py-3 bg-gray-50 rounded-lg border border-gray-200">
            <code className="text-xs font-mono text-gray-700">{node.id}</code>
          </div>
        </div>

        {/* Metadata */}
        {data.metadata && Object.keys(data.metadata).length > 0 && (
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2 block">Metadata</label>
            <div className="space-y-2">
              {Object.entries(data.metadata).map(([key, value]) => (
                <div key={key} className="px-4 py-3 bg-gray-50 rounded-lg border border-gray-200">
                  <div className="text-xs font-semibold text-gray-600 mb-1">{key}</div>
                  <div className="text-sm text-gray-900">
                    {typeof value === 'object' ? (
                      <pre className="text-xs font-mono overflow-x-auto">{JSON.stringify(value, null, 2)}</pre>
                    ) : (
                      <span className="font-mono">{String(value)}</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="px-5 py-4 bg-gray-50 border-t border-gray-200">
        <button
          onClick={onClose}
          className="w-full px-4 py-2.5 bg-gray-200 hover:bg-gray-300 text-gray-700 font-semibold rounded-lg transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  );
});

NodeDetailPanel.displayName = 'NodeDetailPanel';
