import { useLocation } from 'react-router-dom';
import { Bell, Search, ChevronRight } from 'lucide-react';

const routeTitles: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/events': 'Drift Events',
  '/analytics': 'Analytics',
  '/topology': 'Topology',
  '/settings': 'Settings',
};

export function Header() {
  const location = useLocation();
  const title = routeTitles[location.pathname] || 'TFDrift-Falco';

  // Build breadcrumb
  const segments = location.pathname.split('/').filter(Boolean);

  return (
    <header className="h-14 bg-white border-b border-slate-200 flex items-center justify-between px-6 shrink-0">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1 text-sm">
        <span className="text-slate-400">TFDrift</span>
        {segments.map((seg, i) => (
          <span key={seg} className="flex items-center gap-1">
            <ChevronRight className="h-3.5 w-3.5 text-slate-300" />
            <span
              className={
                i === segments.length - 1
                  ? 'text-slate-900 font-medium'
                  : 'text-slate-400'
              }
            >
              {seg.charAt(0).toUpperCase() + seg.slice(1)}
            </span>
          </span>
        ))}
      </div>

      {/* Right side */}
      <div className="flex items-center gap-3">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            type="text"
            placeholder="Search events..."
            className="pl-9 pr-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-slate-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent w-56"
          />
        </div>

        {/* Notifications */}
        <button className="relative p-2 text-slate-500 hover:text-slate-700 hover:bg-slate-100 rounded-lg transition-colors">
          <Bell className="h-5 w-5" />
          <span className="absolute top-1 right-1 h-2 w-2 bg-red-500 rounded-full" />
        </button>

        {/* User avatar */}
        <div className="h-8 w-8 rounded-full bg-indigo-600 flex items-center justify-center text-white text-sm font-medium">
          K
        </div>
      </div>
    </header>
  );
}
