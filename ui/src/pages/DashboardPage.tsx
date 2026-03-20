import { Activity, AlertTriangle, CheckCircle, Cloud } from 'lucide-react';

export function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <SummaryCard
          title="Total Drift Events"
          value="24"
          change="+3 today"
          icon={Activity}
          color="indigo"
        />
        <SummaryCard
          title="Critical"
          value="5"
          change="+1 today"
          icon={AlertTriangle}
          color="red"
        />
        <SummaryCard
          title="Resolved"
          value="18"
          change="75% rate"
          icon={CheckCircle}
          color="green"
        />
        <SummaryCard
          title="Cloud Providers"
          value="2"
          change="AWS, GCP"
          icon={Cloud}
          color="blue"
        />
      </div>

      {/* Placeholder sections */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-72 flex items-center justify-center text-slate-400">
          Drift Events Timeline (Coming Soon)
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-6 h-72 flex items-center justify-center text-slate-400">
          Service Breakdown Chart (Coming Soon)
        </div>
      </div>

      <div className="bg-white rounded-xl border border-slate-200 p-6 h-64 flex items-center justify-center text-slate-400">
        Recent Drift Events Table (Coming Soon)
      </div>
    </div>
  );
}

function SummaryCard({
  title,
  value,
  change,
  icon: Icon,
  color,
}: {
  title: string;
  value: string;
  change: string;
  icon: React.ElementType;
  color: string;
}) {
  const colorClasses: Record<string, string> = {
    indigo: 'bg-indigo-50 text-indigo-600',
    red: 'bg-red-50 text-red-600',
    green: 'bg-green-50 text-green-600',
    blue: 'bg-blue-50 text-blue-600',
  };

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-slate-500">{title}</span>
        <div className={`p-2 rounded-lg ${colorClasses[color]}`}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <p className="text-2xl font-bold text-slate-900">{value}</p>
      <p className="text-xs text-slate-500 mt-1">{change}</p>
    </div>
  );
}
