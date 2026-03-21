import { useState, useMemo, useCallback } from 'react';
import { EventTable } from '../components/events/EventTable';
import { EventFilters } from '../components/events/EventFilters';
import { EventDetailPanel } from '../components/events/EventDetailPanel';
import { useEvents, useUpdateEventStatus } from '../api/hooks/useEvents';
import { mockDriftEvents } from '../mocks/eventData';
import type { DriftFilters, DriftEvent } from '../types/drift';
import type { FalcoEvent, EventStatus } from '../api/types';

/**
 * Adapt a FalcoEvent (API shape) to DriftEvent (UI shape) so existing
 * EventTable / EventFilters keep working with the same interface.
 */
function apiEventToDriftEvent(e: FalcoEvent): DriftEvent {
  return {
    id: e.resource_id,
    timestamp: e.timestamp || new Date().toISOString(),
    severity: (e.severity as DriftEvent['severity']) || 'medium',
    provider: (e.provider as DriftEvent['provider']) || 'aws',
    resourceType: e.resource_type,
    resourceId: e.resource_id,
    resourceName: e.resource_id,
    changeType: 'modified',
    attribute: e.event_name,
    oldValue: null,
    newValue: null,
    userIdentity: {
      type: e.user_identity?.Type || '',
      userName: e.user_identity?.UserName || 'unknown',
      arn: e.user_identity?.ARN,
      accountId: e.user_identity?.AccountID,
    },
    region: e.region || '',
  };
}

export function EventsPage() {
  const [filters, setFilters] = useState<DriftFilters>({});
  const [selectedDriftEvent, setSelectedDriftEvent] = useState<DriftEvent | null>(null);
  const [selectedApiEvent, setSelectedApiEvent] = useState<FalcoEvent | null>(null);

  // Fetch real events from API
  const { data: apiData } = useEvents({
    severity: filters.severity?.[0],
    provider: filters.provider?.[0],
    search: filters.search,
  });
  const updateStatus = useUpdateEventStatus();

  // Merge API events with mock data for display
  const apiEvents: FalcoEvent[] = apiData?.data || [];
  const apiAsDrift = useMemo(() => apiEvents.map(apiEventToDriftEvent), [apiEvents]);

  // Use API events if available, fall back to mock data
  const baseEvents = apiEvents.length > 0 ? apiAsDrift : mockDriftEvents;

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

  // Track selected index for prev/next navigation
  const selectedIndex = useMemo(() => {
    if (!selectedDriftEvent) return -1;
    return filtered.findIndex((e) => e.id === selectedDriftEvent.id);
  }, [filtered, selectedDriftEvent]);

  const handleEventClick = useCallback(
    (driftEvt: DriftEvent) => {
      setSelectedDriftEvent(driftEvt);
      // Find corresponding API event for the detail panel
      const apiEvt = apiEvents.find((e) => e.resource_id === driftEvt.id);
      setSelectedApiEvent(apiEvt || null);
    },
    [apiEvents]
  );

  const handleClose = useCallback(() => {
    setSelectedDriftEvent(null);
    setSelectedApiEvent(null);
  }, []);

  const handlePrev = useCallback(() => {
    if (selectedIndex > 0) {
      const prev = filtered[selectedIndex - 1];
      handleEventClick(prev);
    }
  }, [selectedIndex, filtered, handleEventClick]);

  const handleNext = useCallback(() => {
    if (selectedIndex < filtered.length - 1) {
      const next = filtered[selectedIndex + 1];
      handleEventClick(next);
    }
  }, [selectedIndex, filtered, handleEventClick]);

  const handleStatusChange = useCallback(
    (id: string, status: EventStatus, reason?: string) => {
      updateStatus.mutate({ id, status, reason });
    },
    [updateStatus]
  );

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

      {/* Detail Panel — uses API event if available, otherwise shows basic info */}
      {selectedDriftEvent && selectedApiEvent && (
        <EventDetailPanel
          event={selectedApiEvent}
          onClose={handleClose}
          onStatusChange={handleStatusChange}
          onPrev={handlePrev}
          onNext={handleNext}
          hasPrev={selectedIndex > 0}
          hasNext={selectedIndex < filtered.length - 1}
          currentIndex={selectedIndex}
          totalCount={filtered.length}
        />
      )}

      {/* Fallback: simple panel for mock events without API data */}
      {selectedDriftEvent && !selectedApiEvent && (
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
