/**
 * Connection Settings Tab
 * Webhooks and Cloud Providers configuration
 */

import { useState } from 'react';
import { Plus, Trash2, TestTube } from 'lucide-react';
import { cn } from '../../lib/utils';
import { toast } from '../../stores/toastStore';
import { apiClient } from '../../api/client';
import { PROVIDER_CONFIG, type ProviderKey } from '../../constants';
import { SaveButton } from '../../components/ui/SaveButton';

interface WebhookEntry {
  id: string;
  name: string;
  url: string;
  events: string[];
  enabled: boolean;
}

interface ProviderEntry {
  id: string;
  provider: string;
  regions: string[];
  enabled: boolean;
}

interface ConnectionSettingsProps {
  activeTab: 'webhooks' | 'providers';
}

export function ConnectionSettings({ activeTab }: ConnectionSettingsProps) {
  const [webhooks, setWebhooks] = useState<WebhookEntry[]>([
    { id: '1', name: 'Slack #alerts', url: 'https://hooks.slack.com/services/xxx', events: ['drift', 'falco'], enabled: true },
    { id: '2', name: 'PagerDuty', url: 'https://events.pagerduty.com/v2/enqueue', events: ['drift'], enabled: false },
  ]);

  const [providers, setProviders] = useState<ProviderEntry[]>([
    { id: '1', provider: 'aws', regions: ['us-east-1', 'ap-northeast-1'], enabled: true },
    { id: '2', provider: 'gcp', regions: [], enabled: false },
    { id: '3', provider: 'azure', regions: [], enabled: false },
  ]);

  // Webhook handlers
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

  // Provider handlers
  const updateProvider = (id: string, patch: Partial<ProviderEntry>) => {
    setProviders(providers.map((p) => (p.id === id ? { ...p, ...patch } : p)));
  };

  if (activeTab === 'webhooks') {
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

  // Providers tab
  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-500 dark:text-slate-400">Enable and configure cloud providers for drift detection.</p>

      <div className="space-y-4">
        {providers.map((p) => {
          const info = PROVIDER_CONFIG[p.provider as ProviderKey] || { name: p.provider, shortName: p.provider.toUpperCase(), badgeClass: 'bg-slate-100 text-slate-800' };
          return (
            <div key={p.id} className="border border-slate-200 dark:border-slate-700 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className={cn('px-2.5 py-1 rounded-lg text-xs font-bold uppercase', info.badgeClass)}>
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
