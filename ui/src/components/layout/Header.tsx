import { useLocation } from 'react-router-dom';
import { Bell, Search, ChevronRight, Sun, Moon } from 'lucide-react';
import { useTheme } from '../../hooks/useTheme';

const routeTitles: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/events': 'Drift Events',
  '/analytics': 'Analytics',
  '/topology': 'Topology',
  '/settings': 'Settings',
};

export function Header() {
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();

  // Build breadcrumb
  const segments = location.pathname.split('/').filter(Boolean);

  return (
    <header className="h-14 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between px-6 shrink-0 transition-colors">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1 text-sm">
        <span className="text-slate-400 dark:text-slate-500">TFDrift</span>
        {segments.map((seg, i) => (
          <span key={seg} className="flex items-center gap-1">
            <ChevronRight className="h-3.5 w-3.5 text-slate-300 dark:text-slate-600" />
            <span
              className={
                i === segments.length - 1
                  ? 'text-slate-900 dark:text-slate-100 font-medium'
                  : 'text-slate-400 dark:text-slate-500'
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
            className="pl-9 pr-3 py-1.5 text-sm border border-slate-200 dark:border-slate-600 rounded-lg bg-slate-50 dark:bg-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent w-56 transition-colors"
          />
        </div>

        {/* Theme toggle */}
        <button
          onClick={toggleTheme}
          className="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
          aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          title={theme === 'dark' ? 'Light mode' : 'Dark mode'}
        >
          {theme === 'dark' ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
        </button>

        {/* Notifications */}
        <button className="relative p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
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
