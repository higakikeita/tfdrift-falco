import { useState } from 'react';
import {
  Plus,
  Trash2,
  TestTube,
  Webhook,
  Shield,
  Cloud,
  Settings,
} from 'lucide-react';
import { cn } from '../lib/utils';
import { toast } from '../stores/toastStore';
import { apiClient } from '../api/client';
import { SEVERITY_BADGE_CLASSES, PROVIDER_CONFIG } from '../constants';
import { SaveButton } from '../components/ui/SaveButton';

const tabs = [
  { key: 'webhooks', label: 'Webhooks', icon: Webhook },
  { key: 'rules', label: 'Drift Rules', icon: Shield },
  { key: 'providers', label: 'Cloud Providers', icon: Cloud },
  { key: 'general', label: 'General', icon: Settings },
] as const;
type TabKey = (typeof tabs)[number]['key'];

// --- Types ---
interface WebhookEntry {
  id: string;
  name: string;
  url: string;
  events: string[];
  enabled: boolean;
}

interface RuleEntry {
  id: string;
  name: string;
  resourceTypes: string[];
  watchedAttributes: string[];
  severity: string;
  enabled: boolean;
}

interface ProviderEntry {
  id: string;
  provider: string;
  regions: string[];
  enabled: boolean;
}

// --- Settings Page ---
export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('webhooks');

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Settings</h1>

      {/* Tab Navigation */}
      <div className="border-b border-slate-200 dark:border-slate-700">
        <nav className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={cn(
                'pb-3 text-sm font-medium transition-colors flex items-center gap-1.5',
                activeTab === tab.key
                  ? 'text-indigo-600 dark:text-indigo-400 border-b-2 border-indigo-600 dark:border-indigo-400'
                  : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'
              )}
            >
              <tab.icon className="h-4 w-4" />
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        {activeTab === 'webhooks' && <WebhooksTab />}
        {activeTab === 'rules' && <RulesTab />}
        {activeTab === 'providers' && <ProvidersTab />}
        {activeTab === 'general' && <GeneralTab />}
      </div>
    </div>
  );
}

// --- Webhooks Tab ---
function WebhooksTab() {
  const [webhooks, setWebhooks] = useState<WebhookEntry[]>([
    { id: '1', name: 'Slack #alerts', url: 'https://hooks.slack.com/services/xxx', events: ['drift', 'falco'], enabled: true },
    { id: '2', name: 'PagerDuty', url: 'https://events.pagerduty.com/v2/enqueue', events: ['drift'], enabled: false },
  ]);

  const addWebhook = () => {
    const id = `wh-${Date.now()}`;
    setWebhooks([...webhooks, { id, name: '', url: '', events: ['drift'], enabled: true }]);
  };

  const removeWebhook = (id: string) => {
    setWebhooks(webhooks.filter((w) => w.id !== id));
    toast.info('Webhook removed');
  };

  const updateWebhook = (id: string, patch: Partial<WebhookEntry>) => {
    setWebhooks(webhooks.map((w) => (w.id === id ? { ...w, ...patch } : w)));
  };

  const testWebhook = async (url: string) => {
    try {
      await apiClient.testWebhook(url);
      toast.success('Webhook test sent');
    } catch {
      toast.error('Test failed', 'Could not reach webhook URL');
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-500 dark:text-slate-400">Configure webhook endpoints for drift notifications.</p>
        <button onClick={addWebhook} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors">
          <Plus className="h-3.5 w-3.5" /> Add Webhook
        </button>
      </div>

      {webhooks.length === 0 ? (
        <div className="py-12 text-center text-sm text-slate-400">No webhooks configured.</div>
      ) : (
        <div className="space-y-3">
          {webhooks.map((wh) => (
            <div key={wh.id} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4 space-y-3">
              <div className="flex items-center gap-3">
                <input
                  type="text"
                  value={wh.name}
                  onChange={(e) => updateWebhook(wh.id, { name: e.target.value })}
                  placeholder="Webhook name"
                  className="flex-1 text-sm border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <label className="flex items-center gap-1.5 text-xs text-slate-500">
                  <input
                    type="checkbox"
                    checked={wh.enabled}
                    onChange={(e) => updateWebhook(wh.id, { enabled: e.target.checked })}
                    className="rounded"
                  />
                  Enabled
                </label>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="url"
                  value={wh.url}
                  onChange={(e) => updateWebhook(wh.id, { url: e.target.value })}
                  placeholder="https://hooks.example.com/..."
                  className="flex-1 text-sm font-mono border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                <button onClick={() => testWebhook(wh.url)} className="p-1.5 text-slate-400 hover:text-indigo-600 transition-colors" title="Test webhook">
                  <TestTube className="h-4 w-4" />
                </button>
                <button onClick={() => removeWebhook(wh.id)} className="p-1.5 text-slate-400 hover:text-red-600 transition-colors" title="Remove">
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
              <div className="flex gap-2">
                {['drift', 'falco', 'state_change'].map((evt) => (
                  <label key={evt} className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
                    <input
                      type="checkbox"
                      checked={wh.events.includes(evt)}
                      onChange={(e) => {
                        const events = e.target.checked ? [...wh.events, evt] : wh.events.filter((ev) => ev !== evt);
                        updateWebhook(wh.id, { events });
                      }}
                      className="rounded"
                    />
                    {evt}
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <SaveButton onClick={() => toast.success('Settings saved')} label="Save Webhooks" />
    </div>
  );
}

// --- Rules Tab ---
function RulesTab() {
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
                className={cn('text-xs font-medium px-2 py-1 rounded-full border-0', SEVERITY_BADGE_CLASSES[rule.severity])}
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

// --- Cloud Providers Tab ---
function ProvidersTab() {
  const [providers, setProviders] = useState<ProviderEntry[]>([
    { id: '1', provider: 'aws', regions: ['us-east-1', 'ap-northeast-1'], enabled: true },
    { id: '2', provider: 'gcp', regions: [], enabled: false },
    { id: '3', provider: 'azure', regions: [], enabled: false },
  ]);

  const updateProvider = (id: string, patch: Partial<ProviderEntry>) => {
    setProviders(providers.map((p) => (p.id === id ? { ...p, ...patch } : p)));
  };


  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">Enable and configure cloud providers for drift detection.</p>

      <div className="space-y-4">
        {providers.map((p) => {
          const info = PROVIDER_CONFIG[p.provider as keyof typeof PROVIDER_CONFIG] || { name: p.provider, color: 'bg-slate-100 text-slate-800' };
          return (
            <div key={p.id} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className={cn('px-2.5 py-1 rounded-lg text-xs font-bold uppercase', info.color)}>
                    {p.provider}
                  </span>
                  <span className="text-sm font-medium text-slate-700 dark:text-slate-300">{info.name}</span>
                </div>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={p.enabled}
                    onChange={(e) => updateProvider(p.id, { enabled: e.target.checked })}
                    className="rounded"
                  />
                  <span className={p.enabled ? 'text-green-600 dark:text-green-400 font-medium' : 'text-slate-400'}>
                    {p.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </label>
              </div>
              {p.enabled && (
                <div>
                  <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">Regions</label>
                  <input
                    type="text"
                    value={p.regions.join(', ')}
                    onChange={(e) => updateProvider(p.id, { regions: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
                    placeholder="us-east-1, eu-west-1"
                    className="w-full text-sm font-mono border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>

      <SaveButton onClick={() => toast.success('Providers saved')} label="Save Providers" />
    </div>
  );
}

// --- General Tab ---
function GeneralTab() {
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
