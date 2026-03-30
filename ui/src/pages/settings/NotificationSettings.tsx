/**
 * Notification Settings Tab
 * Drift Rules configuration
 */

import { useState } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { cn } from '../../lib/utils';
import { toast } from '../../stores/toastStore';
import { SEVERITY_BADGE_CLASSES, type SeverityLevel } from '../../constants';
import { SaveButton } from '../../components/ui/SaveButton';

interface RuleEntry {
  id: string;
  name: string;
  resourceTypes: string[];
  watchedAttributes: string[];
  severity: string;
  enabled: boolean;
}

export function NotificationSettings() {
  const [rules, setRules] = useState<RuleEntry[]>([
    { id: '1', name: 'SG Ingress Change', resourceTypes: ['aws_security_group'], watchedAttributes: ['ingress', 'egress'], severity: 'critical', enabled: true },
    { id: '2', name: 'Instance Type Change', resourceTypes: ['aws_instance'], watchedAttributes: ['instance_type'], severity: 'high', enabled: true },
    { id: '3', name: 'S3 Public Access', resourceTypes: ['aws_s3_bucket'], watchedAttributes: ['acl', 'policy'], severity: 'critical', enabled: true },
  ]);

  const addRule = () => {
    const id = `rule-${Date.now()}`;
    setRules([...rules, { id, name: '', resourceTypes: [], watchedAttributes: [], severity: 'medium', enabled: true }]);
  };

  const removeRule = (id: string) => {
    setRules(rules.filter((r) => r.id !== id));
    toast.info('Rule removed');
  };

  const updateRule = (id: string, patch: Partial<RuleEntry>) => {
    setRules(rules.map((r) => (r.id === id ? { ...r, ...patch } : r)));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-500 dark:text-slate-400">Define rules that trigger drift alerts for specific resource types and attributes.</p>
        <button onClick={addRule} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors">
          <Plus className="h-3.5 w-3.5" /> Add Rule
        </button>
      </div>

      <div className="space-y-3">
        {rules.map((rule) => (
          <div key={rule.id} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4 space-y-3">
            <div className="flex items-center gap-3">
              <input
                type="text"
                value={rule.name}
                onChange={(e) => updateRule(rule.id, { name: e.target.value })}
                placeholder="Rule name"
                className="flex-1 text-sm border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
              <select
                value={rule.severity}
                onChange={(e) => updateRule(rule.id, { severity: e.target.value })}
                className={cn('text-xs font-medium px-2 py-1 rounded-full border-0', SEVERITY_BADGE_CLASSES[rule.severity as SeverityLevel])}
              >
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
              <label className="flex items-center gap-1.5 text-xs text-slate-500">
                <input type="checkbox" checked={rule.enabled} onChange={(e) => updateRule(rule.id, { enabled: e.target.checked })} className="rounded" />
                Enabled
              </label>
              <button onClick={() => removeRule(rule.id)} className="p-1.5 text-slate-400 hover:text-red-600 transition-colors">
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">Resource Types</label>
                <input
                  type="text"
                  value={rule.resourceTypes.join(', ')}
                  onChange={(e) => updateRule(rule.id, { resourceTypes: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
                  placeholder="aws_instance, aws_s3_bucket"
                  className="w-full text-xs font-mono border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">Watched Attributes</label>
                <input
                  type="text"
                  value={rule.watchedAttributes.join(', ')}
                  onChange={(e) => updateRule(rule.id, { watchedAttributes: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
                  placeholder="instance_type, ingress"
                  className="w-full text-xs font-mono border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      <SaveButton onClick={() => toast.success('Rules saved')} label="Save Rules" />
    </div>
  );
}
