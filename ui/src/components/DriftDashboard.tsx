import { useState } from 'react';
import { useDriftSummary, useTriggerDriftDetection } from '../api/hooks/useDiscovery';
import { AlertCircle, CheckCircle, XCircle, RefreshCw, AlertTriangle } from 'lucide-react';

interface DriftDashboardProps {
  region?: string;
}

export function DriftDashboard({ region = 'us-east-1' }: DriftDashboardProps) {
  const [autoRefresh, setAutoRefresh] = useState(true);
  const { data: summary, isLoading, error } = useDriftSummary(region, { enabled: autoRefresh });
  const triggerDetection = useTriggerDriftDetection(region);

  const handleManualRefresh = () => {
    triggerDetection.mutate();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-3 text-gray-600">Scanning AWS resources...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <div className="flex items-center">
          <XCircle className="h-5 w-5 text-red-500 mr-2" />
          <span className="text-red-700">Failed to load drift information: {String(error)}</span>
        </div>
      </div>
    );
  }

  if (!summary) {
    return (
      <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
        <p className="text-gray-600">No drift data available</p>
      </div>
    );
  }

  const hasUnmanagedResources = summary.counts.unmanaged > 0;
  const hasMissingResources = summary.counts.missing > 0;
  const hasModifiedResources = summary.counts.modified > 0;
  const hasDrift = hasUnmanagedResources || hasMissingResources || hasModifiedResources;

  return (
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
            disabled={triggerDetection.isPending}
            className="flex items-center px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${triggerDetection.isPending ? 'animate-spin' : ''}`} />
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

            {/* Missing by Type */}
            {hasMissingResources && (
              <div>
                <h4 className="text-sm font-medium text-red-900 mb-2">Missing Resources</h4>
                <div className="space-y-2">
                  {Object.entries(summary.breakdown.missing_by_type).map(([type, count]) => (
                    <div key={type} className="flex items-center justify-between text-sm">
                      <span className="text-gray-700">{type}</span>
                      <span className="font-semibold text-red-600">{count}</span>
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
  );
}
