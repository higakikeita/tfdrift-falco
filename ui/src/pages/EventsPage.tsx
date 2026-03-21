import { useState, useMemo } from 'react';
import { EventTable } from '../components/events/EventTable';
import { EventFilters } from '../components/events/EventFilters';
import { mockDriftEvents } from '../mocks/eventData';
import type { DriftFilters, DriftEvent } from '../types/drift';

export function EventsPage() {
  const [filters, setFilters] = useState<DriftFilters>({});
  const [selectedEvent, setSelectedEvent] = useState<DriftEvent | null>(null);

  const filtered = useMemo(() => {
    return mockDriftEvents.filter((evt) => {
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
  }, [filters]);

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Drift Events</h1>

      <EventFilters
        filters={filters}
        onChange={setFilters}
        totalCount={mockDriftEvents.length}
        filteredCount={filtered.length}
      />

      <EventTable events={filtered} onEventClick={setSelectedEvent} />

      {/* Simple detail panel - will be expanded in #21 */}
      {selectedEvent && (
        <div className="fixed inset-y-0 right-0 w-96 bg-white border-l border-slate-200 shadow-xl p-6 overflow-y-auto z-50">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-slate-900">Event Detail</h2>
            <button onClick={() => setSelectedEvent(null)} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
          </div>
          <dl className="space-y-3 text-sm">
            {([
              ['ID', selectedEvent.id],
              ['Time', new Date(selectedEvent.timestamp).toLocaleString()],
              ['Severity', selectedEvent.severity],
              ['Provider', selectedEvent.provider.toUpperCase()],
              ['Resource', selectedEvent.resourceName || selectedEvent.resourceId],
              ['Type', selectedEvent.resourceType],
              ['Attribute', selectedEvent.attribute],
              ['Old Value', selectedEvent.oldValue ?? '(none)'],
              ['New Value', selectedEvent.newValue ?? '(none)'],
              ['User', selectedEvent.userIdentity.userName],
              ['Region', selectedEvent.region],
              ['Source IP', selectedEvent.sourceIP || '-'],
              ['CloudTrail Event', selectedEvent.cloudtrailEventName || '-'],
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
