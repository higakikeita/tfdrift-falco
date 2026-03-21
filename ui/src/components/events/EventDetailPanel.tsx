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

import { useState, useCallback } from 'react';
import {
  X,
  ChevronLeft,
  ChevronRight,
  CheckCircle2,
  EyeOff,
  Clock,
  Shield,
  User,
  MapPin,
  Activity,
} from 'lucide-react';
import { cn } from '../../lib/utils';
import { JsonDiff } from './JsonDiff';
import type { FalcoEvent, EventStatus, RelatedDrift } from '../../api/types';

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

const statusConfig: Record<EventStatus, { label: string; color: string; icon: typeof Clock }> = {
  open: { label: 'Open', color: 'text-amber-600 bg-amber-50 border-amber-200', icon: Clock },
  acknowledged: { label: 'Acknowledged', color: 'text-blue-600 bg-blue-50 border-blue-200', icon: CheckCircle2 },
  ignored: { label: 'Ignored', color: 'text-slate-500 bg-slate-50 border-slate-200', icon: EyeOff },
  resolved: { label: 'Resolved', color: 'text-green-600 bg-green-50 border-green-200', icon: CheckCircle2 },
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
  const [ignoreReason, setIgnoreReason] = useState('');
  const [showIgnoreForm, setShowIgnoreForm] = useState(false);

  const handleAcknowledge = useCallback(() => {
    if (event && onStatusChange) {
      onStatusChange(event.resource_id, 'acknowledged');
    }
  }, [event, onStatusChange]);

  const handleIgnore = useCallback(() => {
    if (event && onStatusChange && ignoreReason.trim()) {
      onStatusChange(event.resource_id, 'ignored', ignoreReason.trim());
      setShowIgnoreForm(false);
      setIgnoreReason('');
    }
  }, [event, onStatusChange, ignoreReason]);

  const handleReopen = useCallback(() => {
    if (event && onStatusChange) {
      onStatusChange(event.resource_id, 'open');
    }
  }, [event, onStatusChange]);

  if (!event) return null;

  const severity = severityConfig[event.severity] || severityConfig.medium;
  const status = statusConfig[event.status || 'open'];
  const StatusIcon = status.icon;

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
          {/* Status & Actions */}
          <section>
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <StatusIcon className="h-4 w-4 text-slate-500" />
                <span
                  className={cn(
                    'px-2 py-0.5 rounded-full text-xs font-medium border',
                    status.color
                  )}
                >
                  {status.label}
                </span>
                {event.status_reason && (
                  <span className="text-xs text-slate-400 italic ml-1">
                    — {event.status_reason}
                  </span>
                )}
              </div>
            </div>

            {/* Action buttons */}
            {event.status === 'open' && (
              <div className="flex gap-2">
                <button
                  onClick={handleAcknowledge}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 border border-blue-200 rounded-lg transition-colors"
                >
                  <CheckCircle2 className="h-3.5 w-3.5" />
                  Acknowledge
                </button>
                <button
                  onClick={() => setShowIgnoreForm(!showIgnoreForm)}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-600 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-lg transition-colors"
                >
                  <EyeOff className="h-3.5 w-3.5" />
                  Ignore
                </button>
              </div>
            )}

            {(event.status === 'acknowledged' || event.status === 'ignored') && (
              <button
                onClick={handleReopen}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-amber-700 bg-amber-50 hover:bg-amber-100 border border-amber-200 rounded-lg transition-colors"
              >
                <Clock className="h-3.5 w-3.5" />
                Reopen
              </button>
            )}

            {/* Ignore reason form */}
            {showIgnoreForm && (
              <div className="mt-3 p-3 bg-slate-50 rounded-lg border border-slate-200">
                <label className="block text-xs font-medium text-slate-600 mb-1">
                  Reason for ignoring (required)
                </label>
                <textarea
                  value={ignoreReason}
                  onChange={(e) => setIgnoreReason(e.target.value)}
                  placeholder="e.g., Known test environment change"
                  rows={2}
                  className="w-full text-sm border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                />
                <div className="flex justify-end gap-2 mt-2">
                  <button
                    onClick={() => { setShowIgnoreForm(false); setIgnoreReason(''); }}
                    className="px-3 py-1 text-xs text-slate-500 hover:text-slate-700"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleIgnore}
                    disabled={!ignoreReason.trim()}
                    className="px-3 py-1 text-xs font-medium text-white bg-slate-600 hover:bg-slate-700 rounded disabled:opacity-40 disabled:cursor-not-allowed"
                  >
                    Confirm Ignore
                  </button>
                </div>
              </div>
            )}
          </section>

          {/* Event Info */}
          <section>
            <h3 className="text-sm font-semibold text-slate-700 mb-3 flex items-center gap-1.5">
              <Activity className="h-4 w-4" />
              Event Info
            </h3>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs text-slate-400 uppercase tracking-wide">Event</dt>
                <dd className="font-mono text-xs text-slate-900 mt-0.5">{event.event_name}</dd>
              </div>
              <div>
                <dt className="text-xs text-slate-400 uppercase tracking-wide">Provider</dt>
                <dd className="font-medium text-slate-900 mt-0.5 uppercase">{event.provider}</dd>
              </div>
              <div>
                <dt className="text-xs text-slate-400 uppercase tracking-wide">Timestamp</dt>
                <dd className="text-slate-900 mt-0.5">
                  {event.timestamp ? new Date(event.timestamp).toLocaleString() : '-'}
                </dd>
              </div>
              <div>
                <dt className="text-xs text-slate-400 uppercase tracking-wide">Resource ID</dt>
                <dd className="font-mono text-xs text-slate-900 mt-0.5 break-all">{event.resource_id}</dd>
              </div>
            </div>
          </section>

          {/* User Identity */}
          <section>
            <h3 className="text-sm font-semibold text-slate-700 mb-3 flex items-center gap-1.5">
              <User className="h-4 w-4" />
              User Identity
            </h3>
            <div className="space-y-2 text-sm">
              {event.user_identity?.UserName && (
                <div className="flex justify-between">
                  <span className="text-slate-400">User</span>
                  <span className="font-medium text-slate-900">{event.user_identity.UserName}</span>
                </div>
              )}
              {event.user_identity?.ARN && (
                <div>
                  <span className="text-xs text-slate-400">ARN</span>
                  <code className="block text-xs bg-slate-50 text-slate-700 px-2 py-1 rounded mt-0.5 break-all">
                    {event.user_identity.ARN}
                  </code>
                </div>
              )}
              {event.user_identity?.AccountID && (
                <div className="flex justify-between">
                  <span className="text-slate-400">Account</span>
                  <code className="text-xs text-slate-700">{event.user_identity.AccountID}</code>
                </div>
              )}
            </div>
          </section>

          {/* Region */}
          {event.region && (
            <section>
              <h3 className="text-sm font-semibold text-slate-700 mb-2 flex items-center gap-1.5">
                <MapPin className="h-4 w-4" />
                Region
              </h3>
              <span className="inline-flex items-center px-2 py-0.5 rounded bg-slate-100 text-sm font-mono text-slate-700">
                {event.region}
              </span>
            </section>
          )}

          {/* Related Drifts with Diff View */}
          {event.related_drifts && event.related_drifts.length > 0 && (
            <section>
              <h3 className="text-sm font-semibold text-slate-700 mb-3 flex items-center gap-1.5">
                <Shield className="h-4 w-4" />
                Resource Diff ({event.related_drifts.length})
              </h3>
              <div className="space-y-4">
                {event.related_drifts.map((drift: RelatedDrift, idx: number) => (
                  <div key={idx}>
                    <div className="flex items-center gap-2 mb-2">
                      <span
                        className={cn(
                          'px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase border',
                          (severityConfig[drift.severity] || severityConfig.medium).bg,
                          (severityConfig[drift.severity] || severityConfig.medium).color,
                          (severityConfig[drift.severity] || severityConfig.medium).border
                        )}
                      >
                        {drift.severity}
                      </span>
                      {drift.matched_rules?.length > 0 && (
                        <span className="text-[10px] text-slate-400">
                          Rules: {drift.matched_rules.join(', ')}
                        </span>
                      )}
                    </div>
                    <JsonDiff
                      oldValue={drift.old_value}
                      newValue={drift.new_value}
                      attribute={drift.attribute}
                    />
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* Raw Changes */}
          {event.changes && Object.keys(event.changes).length > 0 && (
            <section>
              <h3 className="text-sm font-semibold text-slate-700 mb-3">Changes</h3>
              <pre className="text-xs font-mono bg-slate-50 border border-slate-200 rounded-lg p-3 overflow-auto max-h-60">
                {JSON.stringify(event.changes, null, 2)}
              </pre>
            </section>
          )}
        </div>
      </div>
    </>
  );
}
