import { useState, useMemo, useCallback } from 'react';
import { EventTable } from '../components/events/EventTable';
import { EventFilters } from '../components/events/EventFilters';
import { useDrifts } from '../api/hooks/useDrifts';
import type { DriftFilters, DriftEvent } from '../types/drift';
import type { DriftAlert } from '../api/types';

/**
 * Adapt a DriftAlert (API /drifts shape) to DriftEvent (UI shape) so the
 * existing EventTable / EventFilters keep working. This page reads /drifts
 * (detected drifts), not /events (raw CloudTrail feed, usually empty) — the
 * "Drift Events" list must show drifts (#364).
 */
function apiDriftToDriftEvent(d: DriftAlert): DriftEvent {
  const hasNew = d.new_value && d.new_value !== 'null';
  const hasOld = d.old_value && d.old_value !== 'null';
  return {
    id: d.id || d.resource_id,
    timestamp: d.timestamp || new Date().toISOString(),
    severity: (d.severity as DriftEvent['severity']) || 'medium',
    provider: 'aws',
    resourceType: d.resource_type,
    resourceId: d.resource_id,
    resourceName: d.resource_name || d.resource_id,
    changeType: 'modified',
    attribute: d.attribute || 'modified',
    oldValue: hasOld ? d.old_value : null,
    newValue: hasNew ? d.new_value : null,
    userIdentity: {
      type: d.user_identity?.Type || '',
      userName: d.user_identity?.UserName || d.user_identity?.ARN || 'unknown',
      arn: d.user_identity?.ARN,
      accountId: d.user_identity?.AccountID,
    },
    region: '',
  };
}

export function EventsPage() {
  const [filters, setFilters] = useState<DriftFilters>({});
  const [selectedDriftEvent, setSelectedDriftEvent] = useState<DriftEvent | null>(null);

  // Fetch detected drifts from /api/v1/drifts (#364).
  const { data: apiData } = useDrifts({
    severity: filters.severity?.[0],
    provider: filters.provider?.[0],
    search: filters.search,
  });

  const apiDrifts = useMemo<DriftAlert[]>(() => apiData?.data || [], [apiData]);
  const baseEvents = useMemo(() => apiDrifts.map(apiDriftToDriftEvent), [apiDrifts]);

  const filtered = useMemo(() => {
    return baseEvents.filter((evt) => {
      if (filters.provider?.length && !filters.provider.includes(evt.provider)) return false;
      if (filters.severity?.length && !filters.severity.includes(evt.severity)) return false;
      if (filters.search) {
        const q = filters.search.toLowerCase();
        const hay = [evt.resourceName, evt.resourceId, evt.resourceType, evt.attribute, evt.userIdentity.userName, evt.region]
          .filter(Boolean)
          .join(' ')
          .toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }, [baseEvents, filters]);

  const handleEventClick = useCallback((driftEvt: DriftEvent) => {
    setSelectedDriftEvent(driftEvt);
  }, []);

  const handleClose = useCallback(() => {
    setSelectedDriftEvent(null);
  }, []);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Drift Events</h1>

      <EventFilters
        filters={filters}
        onChange={setFilters}
        totalCount={baseEvents.length}
        filteredCount={filtered.length}
      />

      <EventTable events={filtered} onEventClick={handleEventClick} />

      {/* Detail panel for the selected drift */}
      {selectedDriftEvent && (
        <div className="fixed inset-y-0 right-0 w-96 bg-white border-l border-slate-200 shadow-xl p-6 overflow-y-auto z-50">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-slate-900">Event Detail</h2>
            <button onClick={handleClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
          </div>
          <dl className="space-y-3 text-sm">
            {([
              ['ID', selectedDriftEvent.id],
              ['Time', new Date(selectedDriftEvent.timestamp).toLocaleString()],
              ['Severity', selectedDriftEvent.severity],
              ['Provider', selectedDriftEvent.provider.toUpperCase()],
              ['Resource', selectedDriftEvent.resourceName || selectedDriftEvent.resourceId],
              ['Type', selectedDriftEvent.resourceType],
              ['Attribute', selectedDriftEvent.attribute],
              ['Old Value', selectedDriftEvent.oldValue ?? '(none)'],
              ['New Value', selectedDriftEvent.newValue ?? '(none)'],
              ['User', selectedDriftEvent.userIdentity.userName],
              ['Region', selectedDriftEvent.region],
            ] as [string, string][]).map(([label, val]) => (
              <div key={label}>
                <dt className="text-slate-400 text-xs uppercase tracking-wide">{label}</dt>
                <dd className="text-slate-900 font-mono text-xs break-all">{val}</dd>
              </div>
            ))}
          </dl>
        </div>
      )}
    </div>
  );
}
