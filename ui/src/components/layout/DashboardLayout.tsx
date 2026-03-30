import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { ToastContainer } from '../toast/ToastContainer';

export function DashboardLayout() {
  return (
    <div className="flex h-screen bg-slate-50 dark:bg-slate-950 transition-colors">
      {/* Skip to main content link */}
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-50 focus:bg-blue-600 focus:text-white focus:p-2 focus:rounded"
      >
        Skip to main content
      </a>

      <Sidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <Header />
        <main id="main-content" className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
      <ToastContainer />
    </div>
  );
}
