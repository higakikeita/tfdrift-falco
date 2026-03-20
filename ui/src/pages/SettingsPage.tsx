import { useState } from 'react';

const tabs = ['Webhooks', 'Rules', 'Cloud Providers', 'General'] as const;
type Tab = typeof tabs[number];

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<Tab>('Webhooks');

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Settings</h1>

      {/* Tab Navigation */}
      <div className="border-b border-slate-200">
        <nav className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`pb-3 text-sm font-medium transition-colors ${
                activeTab === tab
                  ? 'text-indigo-600 border-b-2 border-indigo-600'
                  : 'text-slate-500 hover:text-slate-700'
              }`}
            >
              {tab}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="bg-white rounded-xl border border-slate-200 p-6 min-h-[400px] flex items-center justify-center text-slate-400">
        {activeTab} Configuration (Issue #23)
      </div>
    </div>
  );
}
