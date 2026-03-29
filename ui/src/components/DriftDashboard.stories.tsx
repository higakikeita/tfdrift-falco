import type { Meta, StoryObj } from '@storybook/react';
import { DriftDashboard } from './DriftDashboard';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const meta: Meta<typeof DriftDashboard> = {
  title: 'Dashboard/DriftDashboard',
  component: DriftDashboard,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0f172a' },
        { name: 'light', value: '#f8fafc' },
      ],
    },
  },
  decorators: [
    (Story) => {
      const queryClient = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
        },
      });

      return (
        <QueryClientProvider client={queryClient}>
          <div style={{ backgroundColor: '#f8fafc', minHeight: '100vh', padding: '20px' }}>
            <Story />
          </div>
        </QueryClientProvider>
      );
    },
  ],
};

export default meta;
type Story = StoryObj<typeof DriftDashboard>;

export const NoDriftDetected: Story = {
  args: {
    region: 'us-east-1',
  },
  parameters: {
    msw: {
      handlers: [
        // Mock successful API response with no drift
        // For actual implementation, you would configure MSW handlers here
      ],
    },
  },
};

export const WithDriftIssues: Story = {
  args: {
    region: 'us-west-2',
  },
  parameters: {
    msw: {
      handlers: [
        // Mock successful API response with drift data
      ],
    },
  },
};

export const CriticalDriftState: Story = {
  args: {
    region: 'eu-west-1',
  },
};

export const MultipleResourceTypes: Story = {
  args: {
    region: 'ap-southeast-1',
  },
};

export const LargeScaleDrift: Story = {
  args: {
    region: 'us-east-1',
  },
};

export const UsEast1Region: Story = {
  args: {
    region: 'us-east-1',
  },
};

export const EuWest1Region: Story = {
  args: {
    region: 'eu-west-1',
  },
};

export const ApSoutheast1Region: Story = {
  args: {
    region: 'ap-southeast-1',
  },
};
