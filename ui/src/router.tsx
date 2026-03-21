import { createBrowserRouter, Navigate } from 'react-router-dom';
import { DashboardLayout } from './components/layout/DashboardLayout';
import { DashboardPage } from './pages/DashboardPage';
import { EventsPage } from './pages/EventsPage';
import { AnalyticsPage } from './pages/AnalyticsPage';
import { TopologyPage } from './pages/TopologyPage';
import { SettingsPage } from './pages/SettingsPage';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <DashboardLayout />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <DashboardPage /> },
      { path: 'events', element: <EventsPage /> },
      { path: 'analytics', element: <AnalyticsPage /> },
      { path: 'topology', element: <TopologyPage /> },
      { path: 'settings', element: <SettingsPage /> },
    ],
  },
]);
