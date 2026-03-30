/**
 * General Settings Tab
 * Configuration for TFDrift-Falco general settings
 */

import { useState } from 'react';
import { SaveButton } from '../../components/ui/SaveButton';
import { toast } from '../../stores/toastStore';

export function GeneralSettings() {
  const [dryRun, setDryRun] = useState(false);
  const [autoImport, setAutoImport] = useState(false);
  const [requireApproval, setRequireApproval] = useState(true);
  const [pollingInterval, setPollingInterval] = useState(30);

  return (
    <div className="space-y-6 max-w-lg">
      <p className="text-sm text-slate-500 dark:text-slate-400">General configuration for TFDrift-Falco.</p>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <div className="text-sm font-medium text-slate-900 dark:text-slate-100">Dry Run Mode</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Detect drifts without sending notifications</div>
          </div>
          <input type="checkbox" checked={dryRun} onChange={(e) => setDryRun(e.target.checked)} className="rounded" />
        </div>

        <div className="flex items-center justify-between">
          <div>
            <div className="text-sm font-medium text-slate-900 dark:text-slate-100">Auto-Import Unmanaged Resources</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Automatically import resources not in Terraform state</div>
          </div>
          <input type="checkbox" checked={autoImport} onChange={(e) => setAutoImport(e.target.checked)} className="rounded" />
        </div>

        {autoImport && (
          <div className="flex items-center justify-between ml-4">
            <div>
              <div className="text-sm font-medium text-slate-900 dark:text-slate-100">Require Approval</div>
              <div className="text-xs text-slate-500 dark:text-slate-400">Manual approval before importing</div>
            </div>
            <input type="checkbox" checked={requireApproval} onChange={(e) => setRequireApproval(e.target.checked)} className="rounded" />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-slate-900 dark:text-slate-100 mb-1">Polling Interval (seconds)</label>
          <input
            type="number"
            min={5}
            max={300}
            value={pollingInterval}
            onChange={(e) => setPollingInterval(Number(e.target.value))}
            className="w-32 text-sm border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
      </div>

      <SaveButton onClick={() => toast.success('Settings saved')} label="Save Settings" />
    </div>
  );
}
