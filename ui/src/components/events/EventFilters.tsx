import { Search, X } from 'lucide-react';
import type { DriftFilters, DriftSeverity, Provider } from '../../types/drift';

interface EventFiltersProps {
  filters: DriftFilters;
  onChange: (filters: DriftFilters) => void;
  totalCount: number;
  filteredCount: number;
}

export function EventFilters({ filters, onChange, totalCount, filteredCount }: EventFiltersProps) {
  const update = (patch: Partial<DriftFilters>) => onChange({ ...filters, ...patch });

  const hasFilters = (filters.severity?.length ?? 0) > 0
    || (filters.provider?.length ?? 0) > 0
    || !!filters.search;

  return (
    <div className="flex flex-wrap items-center gap-3">
      {/* Search */}
      <div className="relative flex-1 min-w-[200px] max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
        <input
          type="text"
          placeholder="Search events..."
          value={filters.search || ''}
          onChange={(e) => update({ search: e.target.value })}
          className="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      {/* Provider */}
      <select
        value={filters.provider?.[0] || ''}
        onChange={(e) => update({ provider: e.target.value ? [e.target.value as Provider] : undefined })}
        className="px-3 py-2 text-sm border border-slate-200 rounded-lg bg-white"
      >
        <option value="">All Providers</option>
        <option value="aws">AWS</option>
        <option value="gcp">GCP</option>
        <option value="azure">Azure</option>
      </select>

      {/* Severity */}
      <select
        value={filters.severity?.[0] || ''}
        onChange={(e) => update({ severity: e.target.value ? [e.target.value as DriftSeverity] : undefined })}
        className="px-3 py-2 text-sm border border-slate-200 rounded-lg bg-white"
      >
        <option value="">All Severities</option>
        <option value="critical">Critical</option>
        <option value="high">High</option>
        <option value="medium">Medium</option>
        <option value="low">Low</option>
      </select>

      {/* Clear */}
      {hasFilters && (
        <button
          onClick={() => onChange({})}
          className="inline-flex items-center gap-1 px-3 py-2 text-sm text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors"
        >
          <X className="h-3.5 w-3.5" />
          Clear
        </button>
      )}

      {/* Count */}
      <span className="ml-auto text-xs text-slate-400">
        {filteredCount === totalCount ? `${totalCount} events` : `${filteredCount} of ${totalCount} events`}
      </span>
    </div>
  );
}
