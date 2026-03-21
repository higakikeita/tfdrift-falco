import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  AlertTriangle,
  BarChart3,
  Network,
  Settings,
  ChevronLeft,
  ChevronRight,
  Shield,
} from 'lucide-react';
import { useSidebarStore } from '../../stores/sidebarStore';
import { cn } from '../../lib/utils';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/events', icon: AlertTriangle, label: 'Events' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/topology', icon: Network, label: 'Topology' },
  { to: '/settings', icon: Settings, label: 'Settings' },
];

export function Sidebar() {
  const { isCollapsed, toggle } = useSidebarStore();

  return (
    <aside
      className={cn(
        'flex flex-col h-full bg-slate-900 text-slate-300 border-r border-slate-800 transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-60'
      )}
    >
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 py-4 border-b border-slate-800">
        <Shield className="h-7 w-7 text-indigo-400 shrink-0" />
        {!isCollapsed && (
          <span className="text-lg font-bold text-white whitespace-nowrap">
            TFDrift
          </span>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 space-y-1 px-2">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors',
                isActive
                  ? 'bg-indigo-600/20 text-indigo-400'
                  : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
              )
            }
          >
            <item.icon className="h-5 w-5 shrink-0" />
            {!isCollapsed && <span>{item.label}</span>}
          </NavLink>
        ))}
      </nav>

      {/* Collapse Toggle */}
      <button
        onClick={toggle}
        className="flex items-center justify-center py-3 border-t border-slate-800 text-slate-500 hover:text-slate-300 transition-colors"
        aria-label={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      >
        {isCollapsed ? (
          <ChevronRight className="h-5 w-5" />
        ) : (
          <ChevronLeft className="h-5 w-5" />
        )}
      </button>
    </aside>
  );
}
