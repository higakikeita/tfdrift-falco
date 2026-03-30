/**
 * Settings Page
 * Tab-based wrapper that composes the individual settings components
 */

import { useState } from 'react';
import { Webhook, Shield, Cloud, Settings } from 'lucide-react';
import { cn } from '../lib/utils';
import { GeneralSettings, ConnectionSettings, NotificationSettings } from './settings';

const tabs = [
  { key: 'webhooks', label: 'Webhooks', icon: Webhook },
  { key: 'rules', label: 'Drift Rules', icon: Shield },
  { key: 'providers', label: 'Cloud Providers', icon: Cloud },
  { key: 'general', label: 'General', icon: Settings },
] as const;

type TabKey = (typeof tabs)[number]['key'];

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
        {activeTab === 'webhooks' && <ConnectionSettings activeTab="webhooks" />}
        {activeTab === 'rules' && <NotificationSettings />}
        {activeTab === 'providers' && <ConnectionSettings activeTab="providers" />}
        {activeTab === 'general' && <GeneralSettings />}
      </div>
    </div>
  );
}
