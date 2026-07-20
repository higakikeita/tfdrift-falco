import { Activity, AlertTriangle, Database, Cloud } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useStats } from '../api/hooks/useStats';
import { useEvents } from '../api/hooks/useEvents';
import type { FalcoEvent } from '../api/types';

export function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useStats();
  const { data: eventsPage, isLoading: eventsLoading } = useEvents({
    limit: 8,
    sort: 'timestamp',
    order: 'desc',
  });

  const events = eventsPage?.data ?? [];

  // Safe accessors: the stats shape may be partially populated while the
  // backend warms up, so every field falls back to a dash rather than NaN.
  const dash = (n: number | undefined) => (typeof n === 'number' ? String(n) : '—');
  const criticalCount =
    stats?.drifts?.severity_counts?.critical ??
    stats?.severity_breakdown?.critical;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>

      {/* Summary Cards — live from /api/v1/stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <SummaryCard
          title="Total Drift Events"
          value={statsLoading ? '…' : dash(stats?.drifts?.total)}
          icon={Activity}
          color="indigo"
        />
        <SummaryCard
          title="Critical"
          value={statsLoading ? '…' : dash(criticalCount)}
          icon={AlertTriangle}
          color="red"
        />
        <SummaryCard
          title="CloudTrail Events"
          value={statsLoading ? '…' : dash(stats?.events?.total)}
          icon={Cloud}
          color="blue"
        />
        <SummaryCard
          title="Tracked Resources"
          value={statsLoading ? '…' : dash(stats?.graph?.total_nodes)}
          icon={Database}
          color="green"
        />
      </div>

      {/* Recent drift events — the who / when / what feed */}
      <div className="bg-white rounded-xl border border-slate-200">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h2 className="text-sm font-semibold text-slate-700">Recent Drift Events</h2>
          <Link to="/events" className="text-xs font-medium text-indigo-600 hover:text-indigo-700">
            View all →
          </Link>
        </div>

        {eventsLoading ? (
          <div className="p-6 text-sm text-slate-400">Loading events…</div>
        ) : events.length === 0 ? (
          <div className="p-6 text-sm text-slate-400">
            No drift events yet. Changes made outside of Terraform/OpenTofu will appear here.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-slate-500 border-b border-slate-100">
                  <th className="px-6 py-2 font-medium">Event</th>
                  <th className="px-6 py-2 font-medium">Resource</th>
                  <th className="px-6 py-2 font-medium">User</th>
                  <th className="px-6 py-2 font-medium">Severity</th>
                  <th className="px-6 py-2 font-medium">When</th>
                </tr>
              </thead>
              <tbody>
                {events.map((e) => (
                  <EventRow key={e.id} event={e} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function EventRow({ event }: { event: FalcoEvent }) {
  return (
    <tr className="border-b border-slate-50 last:border-0 hover:bg-slate-50">
      <td className="px-6 py-3 font-medium text-slate-800">{event.event_name}</td>
      <td className="px-6 py-3 text-slate-600">
        {event.resource_type}
        {event.resource_id ? <span className="text-slate-400"> · {event.resource_id}</span> : null}
      </td>
      <td className="px-6 py-3 text-slate-600">{event.user_identity?.UserName || '—'}</td>
      <td className="px-6 py-3">
        <SeverityBadge severity={event.severity} />
      </td>
      <td className="px-6 py-3 text-slate-500">{formatTime(event.timestamp)}</td>
    </tr>
  );
}

function SeverityBadge({ severity }: { severity: string }) {
  const s = (severity || '').toLowerCase();
  const cls: Record<string, string> = {
    critical: 'bg-red-50 text-red-700',
    high: 'bg-orange-50 text-orange-700',
    medium: 'bg-amber-50 text-amber-700',
    low: 'bg-slate-100 text-slate-600',
  };
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${cls[s] || 'bg-slate-100 text-slate-600'}`}>
      {severity || 'unknown'}
    </span>
  );
}

function formatTime(ts: string): string {
  if (!ts) return '—';
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return ts;
  return d.toLocaleString();
}

function SummaryCard({
  title,
  value,
  icon: Icon,
  color,
}: {
  title: string;
  value: string;
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
    </div>
  );
}
