import { Activity, AlertTriangle, Cloud, Database } from 'lucide-react';
import { SeverityChart } from '../components/analytics/SeverityChart';
import { ServiceBreakdown } from '../components/analytics/ServiceBreakdown';
import { useStats } from '../api/hooks/useStats';

// Only the dimensions the backend /stats endpoint actually aggregates are
// shown. Time-series and per-user breakdowns have no backend aggregation
// yet, so they are intentionally omitted rather than faked (see #289).
const SEVERITY_FILL: Record<string, string> = {
  Critical: '#ef4444',
  High: '#f97316',
  Medium: '#eab308',
  Low: '#3b82f6',
};

export function AnalyticsPage() {
  const { data: stats, isLoading } = useStats();

  const breakdown = stats?.severity_breakdown ?? {};
  const severityData = (['critical', 'high', 'medium', 'low'] as const)
    .map((k) => {
      const name = k.charAt(0).toUpperCase() + k.slice(1);
      return { name, value: breakdown[k] ?? 0, fill: SEVERITY_FILL[name] };
    })
    .filter((d) => d.value > 0);

  const serviceData = (stats?.top_resource_types ?? []).map((t) => ({
    service: t.resource_type,
    count: t.count,
  }));

  const dash = (n: number | undefined) =>
    isLoading ? '…' : typeof n === 'number' ? String(n) : '—';
  const criticalCount = breakdown.critical;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Analytics</h1>

      {/* KPI Cards — live from /api/v1/stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <KpiCard icon={Activity} label="Total Drifts" value={dash(stats?.drifts?.total)} color="indigo" />
        <KpiCard icon={AlertTriangle} label="Critical" value={dash(criticalCount)} color="red" />
        <KpiCard icon={Cloud} label="CloudTrail Events" value={dash(stats?.events?.total)} color="blue" />
        <KpiCard icon={Database} label="Tracked Resources" value={dash(stats?.graph?.total_nodes)} color="amber" />
      </div>

      {/* Charts — severity distribution + top resource types */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {severityData.length > 0 ? (
          <SeverityChart data={severityData} />
        ) : (
          <EmptyPanel title="Severity Breakdown" />
        )}
        {serviceData.length > 0 ? (
          <ServiceBreakdown data={serviceData} />
        ) : (
          <EmptyPanel title="Top Resource Types" />
        )}
      </div>
    </div>
  );
}

function EmptyPanel({ title }: { title: string }) {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-6">
      <h3 className="text-sm font-semibold text-slate-700 mb-4">{title}</h3>
      <div className="h-56 flex items-center justify-center text-sm text-slate-400">
        No data yet — drift events will populate this chart.
      </div>
    </div>
  );
}

function KpiCard({ icon: Icon, label, value, color }: { icon: React.ElementType; label: string; value: string; color: string }) {
  const bg: Record<string, string> = {
    indigo: 'bg-indigo-50 text-indigo-600',
    red: 'bg-red-50 text-red-600',
    amber: 'bg-amber-50 text-amber-600',
    blue: 'bg-blue-50 text-blue-600',
  };
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">{label}</span>
        <div className={`p-1.5 rounded-lg ${bg[color]}`}><Icon className="h-4 w-4" /></div>
      </div>
      <p className="text-2xl font-bold text-slate-900">{value}</p>
    </div>
  );
}
