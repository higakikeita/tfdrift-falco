import { useState, useMemo } from 'react';
import { ArrowUpDown, ChevronLeft, ChevronRight } from 'lucide-react';
import type { DriftEvent, DriftSeverity } from '../../types/drift';
import { cn } from '../../lib/utils';

interface EventTableProps {
  events: DriftEvent[];
  onEventClick?: (event: DriftEvent) => void;
}

type SortField = 'timestamp' | 'severity' | 'provider' | 'resourceType';
type SortDir = 'asc' | 'desc';

const severityOrder: Record<DriftSeverity, number> = { critical: 0, high: 1, medium: 2, low: 3 };
const severityColors: Record<DriftSeverity, string> = {
  critical: 'bg-red-100 text-red-700 border-red-200',
  high: 'bg-orange-100 text-orange-700 border-orange-200',
  medium: 'bg-yellow-100 text-yellow-700 border-yellow-200',
  low: 'bg-blue-100 text-blue-700 border-blue-200',
};
const providerBadge: Record<string, string> = {
  aws: 'bg-amber-50 text-amber-700',
  gcp: 'bg-blue-50 text-blue-700',
  azure: 'bg-sky-50 text-sky-700',
};

export function EventTable({ events, onEventClick }: EventTableProps) {
  const [sortField, setSortField] = useState<SortField>('timestamp');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);

  const sorted = useMemo(() => {
    return [...events].sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case 'timestamp':
          cmp = new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime();
          break;
        case 'severity':
          cmp = severityOrder[a.severity] - severityOrder[b.severity];
          break;
        case 'provider':
          cmp = a.provider.localeCompare(b.provider);
          break;
        case 'resourceType':
          cmp = a.resourceType.localeCompare(b.resourceType);
          break;
      }
      return sortDir === 'asc' ? cmp : -cmp;
    });
  }, [events, sortField, sortDir]);

  const totalPages = Math.ceil(sorted.length / pageSize);
  const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDir('desc');
    }
  };

  const fmtTime = (iso: string) => {
    const d = new Date(iso);
    return d.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200 bg-slate-50">
              {([
                ['timestamp', 'Time'],
                ['severity', 'Severity'],
                ['provider', 'Provider'],
                ['resourceType', 'Resource'],
              ] as [SortField, string][]).map(([field, label]) => (
                <th
                  key={field}
                  className="px-4 py-3 text-left font-medium text-slate-600 cursor-pointer hover:bg-slate-100 select-none"
                  onClick={() => toggleSort(field)}
                >
                  <span className="inline-flex items-center gap-1">
                    {label}
                    <ArrowUpDown className={cn('h-3.5 w-3.5', sortField === field ? 'text-indigo-500' : 'text-slate-300')} />
                  </span>
                </th>
              ))}
              <th className="px-4 py-3 text-left font-medium text-slate-600">Change</th>
              <th className="px-4 py-3 text-left font-medium text-slate-600">User</th>
              <th className="px-4 py-3 text-left font-medium text-slate-600">Region</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((evt) => (
              <tr
                key={evt.id}
                className="border-b border-slate-100 hover:bg-slate-50 cursor-pointer transition-colors"
                onClick={() => onEventClick?.(evt)}
              >
                <td className="px-4 py-3 text-slate-600 whitespace-nowrap">{fmtTime(evt.timestamp)}</td>
                <td className="px-4 py-3">
                  <span className={cn('px-2 py-0.5 rounded-full text-xs font-medium border', severityColors[evt.severity])}>
                    {evt.severity}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={cn('px-2 py-0.5 rounded text-xs font-medium uppercase', providerBadge[evt.provider])}>
                    {evt.provider}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="font-medium text-slate-900 truncate max-w-[200px]">{evt.resourceName || evt.resourceId}</div>
                  <div className="text-xs text-slate-400 font-mono truncate max-w-[200px]">{evt.resourceType}</div>
                </td>
                <td className="px-4 py-3">
                  <span className="font-mono text-xs text-slate-600 truncate max-w-[150px] block">{evt.attribute}</span>
                </td>
                <td className="px-4 py-3 text-slate-600 text-xs truncate max-w-[120px]">{evt.userIdentity.userName}</td>
                <td className="px-4 py-3 text-slate-500 text-xs">{evt.region}</td>
              </tr>
            ))}
            {paged.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-12 text-center text-slate-400">No events match your filters</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between px-4 py-3 border-t border-slate-200 bg-slate-50">
        <div className="flex items-center gap-2 text-sm text-slate-500">
          <span>Rows per page:</span>
          <select
            value={pageSize}
            onChange={(e) => { setPageSize(Number(e.target.value)); setPage(0); }}
            className="border border-slate-200 rounded px-2 py-1 text-xs bg-white"
          >
            {[10, 25, 50].map((n) => <option key={n} value={n}>{n}</option>)}
          </select>
          <span className="ml-2">
            {sorted.length === 0 ? '0' : `${page * pageSize + 1}-${Math.min((page + 1) * pageSize, sorted.length)}`} of {sorted.length}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="p-1.5 rounded hover:bg-slate-200 disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="p-1.5 rounded hover:bg-slate-200 disabled:opacity-30 disabled:cursor-not-allowed"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
