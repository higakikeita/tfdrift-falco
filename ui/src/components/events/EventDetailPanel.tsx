/**
 * EventDetailPanel - Slide-over panel for drift event details
 * Issue #21: Drift event detail panel with resource diff view
 *
 * Features:
 * - Slide-over animation from the right
 * - Resource info, severity badge, status
 * - JSON diff view for related drifts
 * - Acknowledge / Ignore action buttons
 * - Prev / Next navigation between events
 */

import { X, ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '../../lib/utils';
import {
  EventMetadata,
  EventChanges,
  EventUserInfo,
  StatusActions,
} from './EventDetailPanelSubs';
import type { FalcoEvent, EventStatus } from '../../api/types';

interface EventDetailPanelProps {
  event: FalcoEvent | null;
  onClose: () => void;
  onStatusChange?: (id: string, status: EventStatus, reason?: string) => void;
  onPrev?: () => void;
  onNext?: () => void;
  hasPrev?: boolean;
  hasNext?: boolean;
  currentIndex?: number;
  totalCount?: number;
}

const severityConfig: Record<string, { color: string; bg: string; border: string }> = {
  critical: { color: 'text-red-700', bg: 'bg-red-100', border: 'border-red-200' },
  high: { color: 'text-orange-700', bg: 'bg-orange-100', border: 'border-orange-200' },
  medium: { color: 'text-yellow-700', bg: 'bg-yellow-100', border: 'border-yellow-200' },
  low: { color: 'text-blue-700', bg: 'bg-blue-100', border: 'border-blue-200' },
};

export function EventDetailPanel({
  event,
  onClose,
  onStatusChange,
  onPrev,
  onNext,
  hasPrev = false,
  hasNext = false,
  currentIndex,
  totalCount,
}: EventDetailPanelProps) {
  if (!event) return null;

  const severity = severityConfig[event.severity] || severityConfig.medium;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/20 z-40 transition-opacity"
        onClick={onClose}
      />

      {/* Panel */}
      <div className="fixed inset-y-0 right-0 w-[520px] max-w-full bg-white shadow-2xl z-50 flex flex-col animate-slide-in-right">
        {/* Header */}
        <div className="flex-shrink-0 px-6 py-4 border-b border-slate-200 bg-slate-50">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              {/* Prev / Next */}
              <button
                onClick={onPrev}
                disabled={!hasPrev}
                className="p-1 rounded hover:bg-slate-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                title="Previous event"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              <button
                onClick={onNext}
                disabled={!hasNext}
                className="p-1 rounded hover:bg-slate-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                title="Next event"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
              {currentIndex !== undefined && totalCount !== undefined && (
                <span className="text-xs text-slate-400 ml-1">
                  {currentIndex + 1} / {totalCount}
                </span>
              )}
            </div>
            <button
              onClick={onClose}
              className="p-1 rounded hover:bg-slate-200 text-slate-400 hover:text-slate-600 transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {/* Title */}
          <div className="flex items-start gap-3">
            <div className="flex-1 min-w-0">
              <h2 className="text-lg font-bold text-slate-900 truncate">
                {event.resource_id}
              </h2>
              <p className="text-sm text-slate-500 font-mono truncate">{event.resource_type}</p>
            </div>
            <span
              className={cn(
                'flex-shrink-0 px-2.5 py-0.5 rounded-full text-xs font-semibold border',
                severity.bg, severity.color, severity.border
              )}
            >
              {event.severity?.toUpperCase()}
            </span>
          </div>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
          <StatusActions event={event} onStatusChange={onStatusChange} />
          <EventMetadata event={event} />
          <EventUserInfo event={event} />
          <EventChanges event={event} />
        </div>
      </div>
    </>
  );
}
