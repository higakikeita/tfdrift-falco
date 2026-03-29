import type { Meta, StoryObj } from '@storybook/react';
import { NotificationPanel } from './NotificationPanel';

const meta = {
  title: 'Components/Notifications/NotificationPanel',
  component: NotificationPanel,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <div style={{
        width: '100%',
        minHeight: '400px',
        backgroundColor: '#0f172a',
        padding: '20px',
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'flex-start',
      }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof NotificationPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <NotificationPanel />,
};

export const WithNotifications: Story = {
  render: () => {
    const Panel = NotificationPanel;
    // This would normally receive notifications via SSE
    // For storybook, we're showing the component in its default state
    return <Panel />;
  },
};

export const OpenPanel: Story = {
  render: () => {
    // Note: In a real implementation, this would show the panel open
    // with notifications displayed. Since the panel uses internal state
    // and SSE, we're showing the button that opens it.
    return (
      <div className="relative">
        <NotificationPanel />
        <div className="mt-4 text-xs text-gray-400 max-w-xs">
          Click the notification bell to see notifications (if any are received via SSE)
        </div>
      </div>
    );
  },
};

export const DarkMode: Story = {
  decorators: [
    (Story) => (
      <div style={{
        width: '100%',
        minHeight: '400px',
        backgroundColor: '#1e293b',
        padding: '20px',
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'flex-start',
        color: '#e2e8f0',
      }}>
        <Story />
      </div>
    ),
  ],
  render: () => <NotificationPanel />,
};

export const InHeader: Story = {
  render: () => (
    <div className="w-full bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 px-6 py-4">
      <div className="flex items-center justify-between max-w-7xl mx-auto">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            TFDrift Dashboard
          </h1>
        </div>
        <div className="flex items-center gap-4">
          <input
            type="text"
            placeholder="Search..."
            className="px-4 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100"
          />
          <NotificationPanel />
        </div>
      </div>
    </div>
  ),
};

export const WithMultipleNotificationsSimulation: Story = {
  render: () => (
    <div className="space-y-4">
      <NotificationPanel />
      <div className="bg-slate-800 text-slate-100 text-sm p-4 rounded">
        <p className="font-semibold mb-2">To see notifications:</p>
        <ol className="list-decimal list-inside space-y-1 text-xs">
          <li>Connect to a server with SSE stream</li>
          <li>Drift events will appear in the panel</li>
          <li>Click the bell icon to see all notifications</li>
        </ol>
      </div>
    </div>
  ),
};

export const MinimalLayout: Story = {
  decorators: [
    (Story) => (
      <div style={{
        width: '100%',
        height: '60px',
        backgroundColor: '#0f172a',
        padding: '10px 20px',
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'center',
        borderBottom: '1px solid #1e293b',
      }}>
        <Story />
      </div>
    ),
  ],
  render: () => <NotificationPanel />,
};

export const InSidebar: Story = {
  render: () => (
    <div className="flex h-screen bg-slate-100 dark:bg-slate-950">
      <div className="w-64 bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-700 p-6">
        <nav className="space-y-4">
          <div className="pb-4 border-b border-slate-200 dark:border-slate-700">
            <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">
              Menu
            </h2>
            <ul className="space-y-2">
              <li>
                <a href="#" className="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100">
                  Dashboard
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100">
                  Infrastructure
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100">
                  Reports
                </a>
              </li>
            </ul>
          </div>

          <div className="pt-4">
            <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">
              Settings
            </h2>
            <ul className="space-y-2">
              <li>
                <a href="#" className="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100">
                  Preferences
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100">
                  Account
                </a>
              </li>
            </ul>
          </div>
        </nav>
      </div>

      <div className="flex-1 flex flex-col">
        <div className="bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 px-6 py-4 flex items-center justify-end">
          <NotificationPanel />
        </div>
        <main className="flex-1 p-6">
          <div className="bg-white dark:bg-slate-800 rounded-lg shadow p-6">
            <p className="text-slate-600 dark:text-slate-400">
              Main content area
            </p>
          </div>
        </main>
      </div>
    </div>
  ),
};
