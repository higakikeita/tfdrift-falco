import { useState } from 'react';
import { AlertCircle, CheckCircle, XCircle, RefreshCw, AlertTriangle } from 'lucide-react';

// Demo data to showcase the UI without needing real AWS credentials
const demoSummary = {
  region: 'us-east-1',
  timestamp: new Date().toISOString(),
  counts: {
    terraform_resources: 119,
    aws_resources: 125,
    unmanaged: 6,
    missing: 0,
    modified: 3,
  },
  breakdown: {
    unmanaged_by_type: {
      'aws_security_group': 3,
      'aws_instance': 2,
      'aws_subnet': 1,
    },
    missing_by_type: {},
    modified_by_type: {
      'aws_eks_cluster': 1,
      'aws_db_instance': 2,
    },
  },
};

type ViewMode = 'drift' | 'graph';

function App() {
  const [viewMode, setViewMode] = useState<ViewMode>('drift');
  const [isScanning, setIsScanning] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const summary = demoSummary;
  const hasUnmanagedResources = summary.counts.unmanaged > 0;
  const hasMissingResources = summary.counts.missing > 0;
  const hasModifiedResources = summary.counts.modified > 0;
  const hasDrift = hasUnmanagedResources || hasMissingResources || hasModifiedResources;

  const handleManualRefresh = () => {
    setIsScanning(true);
    setTimeout(() => setIsScanning(false), 2000);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">
                TFDrift-Falco
              </h1>
              <span className="ml-3 px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
                v0.4.1 DEMO
              </span>
            </div>

            {/* View Mode Tabs */}
            <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg">
              <button
                onClick={() => setViewMode('drift')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                  viewMode === 'drift'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Drift Detection
              </button>
              <button
                onClick={() => setViewMode('graph')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                  viewMode === 'graph'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Graph View
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {viewMode === 'drift' ? (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">AWS Drift Detection</h2>
                <p className="text-sm text-gray-600 mt-1">Region: {summary.region}</p>
                <p className="text-xs text-gray-500">Last updated: {new Date(summary.timestamp).toLocaleString()}</p>
              </div>
              <div className="flex items-center space-x-3">
                <label className="flex items-center text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={autoRefresh}
                    onChange={(e) => setAutoRefresh(e.target.checked)}
                    className="mr-2"
                  />
                  Auto-refresh (5min)
                </label>
                <button
                  onClick={handleManualRefresh}
                  disabled={isScanning}
                  className="flex items-center px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  <RefreshCw className={`h-4 w-4 mr-2 ${isScanning ? 'animate-spin' : ''}`} />
                  Scan Now
                </button>
              </div>
            </div>

            {/* Overall Status */}
            <div className={`border-l-4 p-4 rounded-r-lg ${
              hasDrift
                ? 'bg-yellow-50 border-yellow-500'
                : 'bg-green-50 border-green-500'
            }`}>
              <div className="flex items-center">
                {hasDrift ? (
                  <>
                    <AlertTriangle className="h-6 w-6 text-yellow-600 mr-3" />
                    <div>
                      <h3 className="text-lg font-semibold text-yellow-900">Drift Detected</h3>
                      <p className="text-sm text-yellow-700">
                        {summary.counts.unmanaged + summary.counts.missing + summary.counts.modified} resource(s) differ from Terraform state
                      </p>
                    </div>
                  </>
                ) : (
                  <>
                    <CheckCircle className="h-6 w-6 text-green-600 mr-3" />
                    <div>
                      <h3 className="text-lg font-semibold text-green-900">No Drift Detected</h3>
                      <p className="text-sm text-green-700">
                        All AWS resources match Terraform state
                      </p>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Total Resources */}
              <div className="bg-white border border-gray-200 rounded-lg p-4 shadow-sm">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-600">Terraform Resources</p>
                    <p className="text-3xl font-bold text-gray-900">{summary.counts.terraform_resources}</p>
                  </div>
                  <div className="bg-blue-100 p-3 rounded-lg">
                    <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                </div>
                <p className="text-xs text-gray-500 mt-2">Managed by Terraform</p>
              </div>

              {/* Unmanaged Resources */}
              <div className={`border rounded-lg p-4 shadow-sm ${
                hasUnmanagedResources
                  ? 'bg-orange-50 border-orange-300'
                  : 'bg-white border-gray-200'
              }`}>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-600">Unmanaged</p>
                    <p className={`text-3xl font-bold ${hasUnmanagedResources ? 'text-orange-600' : 'text-gray-900'}`}>
                      {summary.counts.unmanaged}
                    </p>
                  </div>
                  <div className={`p-3 rounded-lg ${hasUnmanagedResources ? 'bg-orange-200' : 'bg-gray-100'}`}>
                    <AlertCircle className={`h-6 w-6 ${hasUnmanagedResources ? 'text-orange-600' : 'text-gray-400'}`} />
                  </div>
                </div>
                <p className="text-xs text-gray-600 mt-2">Manually created resources</p>
              </div>

              {/* Missing Resources */}
              <div className={`border rounded-lg p-4 shadow-sm ${
                hasMissingResources
                  ? 'bg-red-50 border-red-300'
                  : 'bg-white border-gray-200'
              }`}>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-600">Missing</p>
                    <p className={`text-3xl font-bold ${hasMissingResources ? 'text-red-600' : 'text-gray-900'}`}>
                      {summary.counts.missing}
                    </p>
                  </div>
                  <div className={`p-3 rounded-lg ${hasMissingResources ? 'bg-red-200' : 'bg-gray-100'}`}>
                    <XCircle className={`h-6 w-6 ${hasMissingResources ? 'text-red-600' : 'text-gray-400'}`} />
                  </div>
                </div>
                <p className="text-xs text-gray-600 mt-2">Manually deleted resources</p>
              </div>

              {/* Modified Resources */}
              <div className={`border rounded-lg p-4 shadow-sm ${
                hasModifiedResources
                  ? 'bg-yellow-50 border-yellow-300'
                  : 'bg-white border-gray-200'
              }`}>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-600">Modified</p>
                    <p className={`text-3xl font-bold ${hasModifiedResources ? 'text-yellow-600' : 'text-gray-900'}`}>
                      {summary.counts.modified}
                    </p>
                  </div>
                  <div className={`p-3 rounded-lg ${hasModifiedResources ? 'bg-yellow-200' : 'bg-gray-100'}`}>
                    <AlertTriangle className={`h-6 w-6 ${hasModifiedResources ? 'text-yellow-600' : 'text-gray-400'}`} />
                  </div>
                </div>
                <p className="text-xs text-gray-600 mt-2">Manually modified resources</p>
              </div>
            </div>

            {/* Resource Type Breakdown */}
            {hasDrift && (
              <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Resource Type Breakdown</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Unmanaged by Type */}
                  {hasUnmanagedResources && (
                    <div>
                      <h4 className="text-sm font-medium text-orange-900 mb-2">Unmanaged Resources</h4>
                      <div className="space-y-2">
                        {Object.entries(summary.breakdown.unmanaged_by_type).map(([type, count]) => (
                          <div key={type} className="flex items-center justify-between text-sm">
                            <span className="text-gray-700">{type}</span>
                            <span className="font-semibold text-orange-600">{count}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Modified by Type */}
                  {hasModifiedResources && (
                    <div>
                      <h4 className="text-sm font-medium text-yellow-900 mb-2">Modified Resources</h4>
                      <div className="space-y-2">
                        {Object.entries(summary.breakdown.modified_by_type).map(([type, count]) => (
                          <div key={type} className="flex items-center justify-between text-sm">
                            <span className="text-gray-700">{type}</span>
                            <span className="font-semibold text-yellow-600">{count}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              Resource Dependency Graph
            </h2>
            <div className="text-center text-gray-500 py-12">
              <p>Graph view coming soon</p>
              <p className="text-sm mt-2">Switch to Drift Detection to see AWS resource drift</p>
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            TFDrift-Falco - Terraform Drift Detection with Falco Integration (Demo Mode)
          </p>
        </div>
      </footer>
    </div>
  );
}

export default App;
