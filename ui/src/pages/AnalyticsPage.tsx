import { Activity, AlertTriangle, TrendingUp, Users } from 'lucide-react';
import { TimelineChart } from '../components/analytics/TimelineChart';
import { SeverityChart } from '../components/analytics/SeverityChart';
import { ServiceBreakdown } from '../components/analytics/ServiceBreakdown';
import { TopUsersChart } from '../components/analytics/TopUsersChart';
import { timelineData, severityData, serviceData, topUsersData } from '../mocks/analyticsData';

export function AnalyticsPage() {
  const totalEvents = severityData.reduce((s, d) => s + d.value, 0);
  const criticalCount = severityData.find((d) => d.name === 'Critical')?.value ?? 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Analytics</h1>
        <select className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white">
          <option>Last 7 days</option>
          <option>Last 30 days</option>
          <option>Last 90 days</option>
        </select>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard icon={Activity} label="Total Events" value={String(totalEvents)} sub="Last 7 days" color="indigo" />
        <KpiCard icon={AlertTriangle} label="Critical" value={String(criticalCount)} sub={`${((criticalCount / totalEvents) * 100).toFixed(0)}% of total`} color="red" />
        <KpiCard icon={TrendingUp} label="Avg/Day" value={(totalEvents / 7).toFixed(1)} sub="Trend: +12%" color="amber" />
        <KpiCard icon={Users} label="Unique Users" value="5" sub="Caused drift" color="blue" />
      </div>

      {/* Charts - 2x2 grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <TimelineChart data={timelineData} />
        <SeverityChart data={severityData} />
        <ServiceBreakdown data={serviceData} />
        <TopUsersChart data={topUsersData} />
      </div>
    </div>
  );
}

function KpiCard({ icon: Icon, label, value, sub, color }: { icon: React.ElementType; label: string; value: string; sub: string; color: string }) {
  const bg: Record<string, string> = { indigo: 'bg-indigo-50 text-indigo-600', red: 'bg-red-50 text-red-600', amber: 'bg-amber-50 text-amber-600', blue: 'bg-blue-50 text-blue-600' };
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">{label}</span>
        <div className={`p-1.5 rounded-lg ${bg[color]}`}><Icon className="h-4 w-4" /></div>
      </div>
      <p className="text-2xl font-bold text-slate-900">{value}</p>
      <p className="text-xs text-slate-400 mt-0.5">{sub}</p>
    </div>
  );
}
