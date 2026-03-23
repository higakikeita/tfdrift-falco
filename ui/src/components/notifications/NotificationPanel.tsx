/**
 * NotificationPanel - Real-time drift notification panel (#16)
 * Connects to SSE stream and displays live drift alerts.
 */

import { useState, useEffect, useCallback } from 'react';
import { Bell, X, Wifi, WifiOff, Trash2 } from 'lucide-react';
import { useSSE, type SSEEvent } from '../../api/sse';
import { toast } from '../../stores/toastStore';
import { cn } from '../../lib/utils';

const severityColors: Record<string, string> = {
  critical: 'border-l-red-500 bg-red-50 dark:bg-red-950',
  high: 'border-l-orange-500 bg-orange-50 dark:bg-orange-950',
  medium: 'border-l-yellow-500 bg-yellow-50 dark:bg-yellow-950',
  low: 'border-l-blue-500 bg-blue-50 dark:bg-blue-950',
};

interface Notification {
  id: string;
  type: string;
  severity: string;
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
}

function sseEventToNotification(event: SSEEvent): Notification | null {
  const payload = event.data as Record<string, unknown> | null;
  if (!payload) return null;

  const id = `${event.type}-${event.timestamp}-${Math.random().toString(36).slice(2, 8)}`;
  const severity = (payload.severity as string) || 'medium';

  if (event.type === 'drift') {
    return {
      id,
      type: 'drift',
      severity,
      title: `Drift: ${(payload.resource_type as string) || 'Unknown'}`,
      message: `${(payload.attribute as string) || ''} changed on ${(payload.resource_id as string) || 'unknown'}`,
      timestamp: event.timestamp || new Date().toISOString(),
      read: false,
    };
  }

  if (event.type === 'falco') {
    return {
      id,
      type: 'falco',
      severity: 'high',
      title: `Falco: ${(payload.event_name as string) || 'Event'}`,
      message: `${(payload.resource_type as string) || ''} ${(payload.resource_id as string) || ''}`,
      timestamp: event.timestamp || new Date().toISOString(),
      read: false,
    };
  }

  return null;
}

export function NotificationPanel() {
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const { isConnected, lastEvent } = useSSE({ autoConnect: true });

  // Process new SSE events into notifications + toast
  useEffect(() => {
    if (!lastEvent) return;
    const notif = sseEventToNotification(lastEvent);
    if (!notif) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setNotifications((prev) => [notif, ...prev].slice(0, 100));

    // Show toast for critical/high severity
    if (notif.severity === 'critical') {
      toast.error(notif.title, notif.message);
    } else if (notif.severity === 'high') {
      toast.warning(notif.title, notif.message);
    }
  }, [lastEvent]);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const markAllRead = useCallback(() => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  }, []);

  const clearAll = useCallback(() => {
    setNotifications([]);
  }, []);

  const fmtTime = (iso: string) => {
    try {
      const d = new Date(iso);
      const now = new Date();
      const diff = now.getTime() - d.getTime();
      if (diff < 60000) return 'just now';
      if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
      if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
      return d.toLocaleDateString();
    } catch {
      return '';
    }
  };

  return (
    <div className="relative">
      {/* Bell button */}
      <button
        onClick={() => { setIsOpen(!isOpen); if (!isOpen) markAllRead(); }}
        className="relative p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
      >
        <Bell className="h-5 w-5" />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] flex items-center justify-center bg-red-500 text-white text-[10px] font-bold rounded-full px-1">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown panel */}
      {isOpen && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setIsOpen(false)} />
          <div className="absolute right-0 top-full mt-2 w-96 max-h-[480px] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-xl shadow-xl z-50 flex flex-col overflow-hidden">
            {/* Header */}
            <div className="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Notifications</h3>
                <span className={cn(
                  'flex items-center gap-1 text-[10px] font-medium px-1.5 py-0.5 rounded-full',
                  isConnected
                    ? 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300'
                    : 'bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300'
                )}>
                  {isConnected ? <Wifi className="h-2.5 w-2.5" /> : <WifiOff className="h-2.5 w-2.5" />}
                  {isConnected ? 'Live' : 'Offline'}
                </span>
              </div>
              <div className="flex items-center gap-1">
                {notifications.length > 0 && (
                  <button onClick={clearAll} className="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300" title="Clear all">
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                )}
                <button onClick={() => setIsOpen(false)} className="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300">
                  <X className="h-4 w-4" />
                </button>
              </div>
            </div>

            {/* List */}
            <div className="flex-1 overflow-y-auto">
              {notifications.length === 0 ? (
                <div className="py-12 text-center text-sm text-slate-400 dark:text-slate-500">
                  No notifications yet.{isConnected ? ' Listening for drift events...' : ' Connect to receive events.'}
                </div>
              ) : (
                notifications.map((n) => (
                  <div
                    key={n.id}
                    className={cn(
                      'px-4 py-3 border-b border-slate-100 dark:border-slate-800 border-l-4 transition-colors',
                      severityColors[n.severity] || 'border-l-slate-300 bg-white dark:bg-slate-900',
                      !n.read && 'font-medium'
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="text-sm text-slate-900 dark:text-slate-100 truncate">{n.title}</p>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 truncate">{n.message}</p>
                      </div>
                      <span className="text-[10px] text-slate-400 shrink-0 mt-0.5">{fmtTime(n.timestamp)}</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
