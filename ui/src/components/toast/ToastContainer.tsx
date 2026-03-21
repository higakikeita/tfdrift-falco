/**
 * Toast Notification Container
 * Renders toast notifications in the bottom-right corner with auto-dismiss.
 */

import { X, CheckCircle2, AlertTriangle, AlertCircle, Info } from 'lucide-react';
import { useToastStore, type ToastType } from '../../stores/toastStore';
import { cn } from '../../lib/utils';

const iconMap: Record<ToastType, typeof CheckCircle2> = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

const styleMap: Record<ToastType, string> = {
  success: 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950',
  error: 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950',
  warning: 'border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950',
  info: 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950',
};

const iconColorMap: Record<ToastType, string> = {
  success: 'text-green-600 dark:text-green-400',
  error: 'text-red-600 dark:text-red-400',
  warning: 'text-amber-600 dark:text-amber-400',
  info: 'text-blue-600 dark:text-blue-400',
};

export function ToastContainer() {
  const { toasts, removeToast } = useToastStore();

  if (toasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
      {toasts.map((t) => {
        const Icon = iconMap[t.type];
        return (
          <div
            key={t.id}
            className={cn(
              'flex items-start gap-3 px-4 py-3 rounded-lg border shadow-lg animate-slide-in-right',
              styleMap[t.type]
            )}
            role="alert"
          >
            <Icon className={cn('h-5 w-5 mt-0.5 shrink-0', iconColorMap[t.type])} />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{t.title}</p>
              {t.message && (
                <p className="text-xs text-slate-600 dark:text-slate-400 mt-0.5">{t.message}</p>
              )}
            </div>
            <button
              onClick={() => removeToast(t.id)}
              className="shrink-0 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        );
      })}
    </div>
  );
}
