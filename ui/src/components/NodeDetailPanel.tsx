/**
 * NodeDetailPanel - Display detailed node information and relationships
 */

import React from 'react';
import { X, Network } from 'lucide-react';
import { NodeProperties, NodeRelationships, NodeMetadata } from './NodeDetailPanelSubs';

interface NodeDetailPanelProps {
  nodeId: string;
  onClose: () => void;
  onNodeSelect?: (nodeId: string) => void;
  onShowImpactRadius?: (nodeIds: string[], depth: number) => void;
}

type TabType = 'overview' | 'relationships' | 'impact';

const NodeDetailPanel = ({ nodeId, onClose, onNodeSelect, onShowImpactRadius }: NodeDetailPanelProps) => {
  const [activeTab, setActiveTab] = React.useState<TabType>('overview');
  const [impactDepth, setImpactDepth] = React.useState(2);
  const [showImpact, setShowImpact] = React.useState(false);

  const handleShowImpact = () => {
    setShowImpact(true);
  };

  const handleHideImpact = () => {
    setShowImpact(false);
    if (onShowImpactRadius) {
      onShowImpactRadius([], 0);
    }
  };

  return (
    <div className="h-full bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex flex-col animate-in slide-in-from-right duration-300 w-full md:w-96 lg:w-[28rem] xl:w-[32rem]">
      {/* Header */}
      <div className="px-3 sm:px-4 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <Network className="w-4 h-4 sm:w-5 sm:h-5 flex-shrink-0" />
          <h3 className="font-semibold text-sm sm:text-base truncate">ãã¼ãè©³ç´°</h3>
        </div>
        <button
          onClick={onClose}
          className="p-1 hover:bg-white/20 rounded transition-colors flex-shrink-0"
          aria-label="éãã"
        >
          <X className="w-4 h-4 sm:w-5 sm:h-5" />
        </button>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex-shrink-0">
        <nav className="flex overflow-x-auto">
          <button
            onClick={() => setActiveTab('overview')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'overview'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            æ¦è¦
          </button>
          <button
            onClick={() => setActiveTab('relationships')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'relationships'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            é¢ä¿æ§
          </button>
          <button
            onClick={() => setActiveTab('impact')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'impact'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            å½±é¿ç¯å²
          </button>
        </nav>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3 sm:p-4 space-y-4">
        {/* Node Basic Info */}
        <div
          className="transition-opacity duration-200"
          style={{
            display: activeTab === 'overview' ? 'block' : 'none',
            opacity: activeTab === 'overview' ? 1 : 0,
          }}
        >
          <NodeProperties nodeId={nodeId} />
        </div>

        {/* Dependencies and Relationships */}
        <div
          className="space-y-4 transition-opacity duration-200"
          style={{
            display: activeTab === 'relationships' ? 'block' : 'none',
            opacity: activeTab === 'relationships' ? 1 : 0,
          }}
        >
          <NodeRelationships nodeId={nodeId} onNodeSelect={onNodeSelect} />
        </div>

        {/* Impact Radius */}
        <div
          className="transition-opacity duration-200"
          style={{
            display: activeTab === 'impact' ? 'block' : 'none',
            opacity: activeTab === 'impact' ? 1 : 0,
          }}
        >
          <NodeMetadata
            nodeId={nodeId}
            impactDepth={impactDepth}
            onImpactDepthChange={setImpactDepth}
            showImpact={showImpact}
            onShowImpact={handleShowImpact}
            onHideImpact={handleHideImpact}
            onShowImpactRadius={onShowImpactRadius}
          />
        </div>
      </div>
    </div>
  );
};

export default NodeDetailPanel;
