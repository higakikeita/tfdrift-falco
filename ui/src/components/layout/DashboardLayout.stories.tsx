import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { DashboardLayout } from './DashboardLayout';

const meta = {
  title: 'Components/Layout/DashboardLayout',
  component: DashboardLayout,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <div style={{ width: '100%', height: '100vh', backgroundColor: '#0f172a' }}>
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta<typeof DashboardLayout>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<DashboardLayout />}>
            <Route
              index
              element={
                <div className="space-y-6">
                  <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                    Dashboard Overview
                  </h1>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <div className="text-sm font-medium text-slate-500 dark:text-slate-400">
                        Total Resources
                      </div>
                      <div className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-2">
                        1,234
                      </div>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <div className="text-sm font-medium text-slate-500 dark:text-slate-400">
                        Active Drifts
                      </div>
                      <div className="text-2xl font-bold text-red-600 dark:text-red-400 mt-2">
                        28
                      </div>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <div className="text-sm font-medium text-slate-500 dark:text-slate-400">
                        Compliant
                      </div>
                      <div className="text-2xl font-bold text-green-600 dark:text-green-400 mt-2">
                        92%
                      </div>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <div className="text-sm font-medium text-slate-500 dark:text-slate-400">
                        Last Sync
                      </div>
                      <div className="text-lg font-bold text-slate-900 dark:text-slate-100 mt-2">
                        5m ago
                      </div>
                    </div>
                  </div>
                </div>
              }
            />
          </Route>
        </Routes>
      </MemoryRouter>
    ),
  ],
};

export const WithDetailedContent: Story = {
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<DashboardLayout />}>
            <Route
              index
              element={
                <div className="space-y-6">
                  <div className="flex items-center justify-between">
                    <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                      Drift Analysis
                    </h1>
                    <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium">
                      Generate Report
                    </button>
                  </div>

                  <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="lg:col-span-2 bg-white dark:bg-slate-800 rounded-lg shadow-md p-6">
                      <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">
                        Drift Timeline
                      </h2>
                      <div className="space-y-3">
                        {Array.from({ length: 8 }).map((_, i) => (
                          <div
                            key={i}
                            className="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-700 rounded"
                          >
                            <div className="w-2 h-2 bg-orange-500 rounded-full" />
                            <div className="flex-1">
                              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">
                                Resource-{i}: Config Changed
                              </p>
                              <p className="text-xs text-slate-500 dark:text-slate-400">
                                {10 - i} hours ago
                              </p>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="space-y-4">
                      <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6">
                        <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">
                          Severity Breakdown
                        </h3>
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">Critical</span>
                            <span className="text-sm font-bold text-red-600">5</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">High</span>
                            <span className="text-sm font-bold text-orange-600">12</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">Medium</span>
                            <span className="text-sm font-bold text-yellow-600">8</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">Low</span>
                            <span className="text-sm font-bold text-blue-600">3</span>
                          </div>
                        </div>
                      </div>

                      <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md p-6">
                        <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">
                          Top Providers
                        </h3>
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">AWS</span>
                            <span className="text-sm font-bold">18</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">GCP</span>
                            <span className="text-sm font-bold">7</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-slate-600 dark:text-slate-400">Kubernetes</span>
                            <span className="text-sm font-bold">3</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              }
            />
          </Route>
        </Routes>
      </MemoryRouter>
    ),
  ],
};

export const WithGraphContent: Story = {
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<DashboardLayout />}>
            <Route
              index
              element={
                <div className="space-y-6">
                  <div className="flex items-center justify-between">
                    <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                      Infrastructure Graph
                    </h1>
                    <div className="flex gap-2">
                      <button className="px-4 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-900 dark:text-slate-100 rounded-lg font-medium text-sm">
                        Reset View
                      </button>
                      <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium text-sm">
                        Export
                      </button>
                    </div>
                  </div>

                  <div className="bg-white dark:bg-slate-800 rounded-lg shadow-md h-96 flex items-center justify-center">
                    <div className="text-center text-slate-500 dark:text-slate-400">
                      <p className="text-lg font-semibold mb-2">Graph Visualization Area</p>
                      <p className="text-sm">Placeholder for graph component</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Nodes</p>
                      <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">
                        256
                      </p>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <p className="text-xs font-medium text-slate-500 dark:text-slate-400">
                        Connections
                      </p>
                      <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">
                        483
                      </p>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <p className="text-xs font-medium text-slate-500 dark:text-slate-400">
                        Clusters
                      </p>
                      <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">
                        12
                      </p>
                    </div>
                    <div className="p-4 bg-white dark:bg-slate-800 rounded-lg shadow-md">
                      <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Status</p>
                      <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
                        Healthy
                      </p>
                    </div>
                  </div>
                </div>
              }
            />
          </Route>
        </Routes>
      </MemoryRouter>
    ),
  ],
};

export const WithEmptyState: Story = {
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<DashboardLayout />}>
            <Route
              index
              element={
                <div className="flex items-center justify-center h-full">
                  <div className="text-center">
                    <div className="text-6xl mb-4">📋</div>
                    <h2 className="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">
                      No data available
                    </h2>
                    <p className="text-slate-600 dark:text-slate-400 mb-6">
                      Start by connecting your cloud accounts or importing infrastructure
                    </p>
                    <button className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium">
                      Get Started
                    </button>
                  </div>
                </div>
              }
            />
          </Route>
        </Routes>
      </MemoryRouter>
    ),
  ],
};
