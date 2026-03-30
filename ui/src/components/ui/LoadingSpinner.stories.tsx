import type { Meta, StoryObj } from '@storybook/react';
import { LoadingSpinner } from './LoadingSpinner';

const meta: Meta<typeof LoadingSpinner> = {
  title: 'UI/LoadingSpinner',
  component: LoadingSpinner,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
    backgrounds: {
      default: 'light',
      values: [
        { name: 'light', value: '#f8fafc' },
        { name: 'dark', value: '#0f172a' },
      ],
    },
  },
  args: {
    size: 'md',
  },
};

export default meta;
type Story = StoryObj<typeof LoadingSpinner>;

/**
 * Default LoadingSpinner with medium size and no text.
 */
export const Default: Story = {
  args: {
    size: 'md',
  },
};

/**
 * Small loading spinner, useful for inline loading states.
 */
export const Small: Story = {
  args: {
    size: 'sm',
  },
};

/**
 * Medium loading spinner (default size).
 */
export const Medium: Story = {
  args: {
    size: 'md',
  },
};

/**
 * Large loading spinner, prominent display for full-page loading.
 */
export const Large: Story = {
  args: {
    size: 'lg',
  },
};

/**
 * LoadingSpinner with loading text below the spinner.
 */
export const WithText: Story = {
  args: {
    size: 'md',
    text: 'Loading...',
  },
};

/**
 * Small spinner with custom loading text.
 */
export const SmallWithText: Story = {
  args: {
    size: 'sm',
    text: 'Processing',
  },
};

/**
 * Medium spinner with detailed loading message.
 */
export const MediumWithDetailedText: Story = {
  args: {
    size: 'md',
    text: 'Analyzing infrastructure drift...',
  },
};

/**
 * Large spinner with descriptive text for prominent loading state.
 */
export const LargeWithText: Story = {
  args: {
    size: 'lg',
    text: 'Fetching terraform state',
  },
};

/**
 * Multiple spinners at different sizes to show scale comparison.
 */
export const AllSizes: Story = {
  render: () => (
    <div
      style={{
        display: 'flex',
        gap: '40px',
        alignItems: 'flex-start',
        justifyContent: 'center',
        padding: '40px',
      }}
    >
      <div style={{ textAlign: 'center' }}>
        <LoadingSpinner size="sm" />
        <p style={{ marginTop: '10px', fontSize: '14px' }}>Small</p>
      </div>
      <div style={{ textAlign: 'center' }}>
        <LoadingSpinner size="md" />
        <p style={{ marginTop: '10px', fontSize: '14px' }}>Medium</p>
      </div>
      <div style={{ textAlign: 'center' }}>
        <LoadingSpinner size="lg" />
        <p style={{ marginTop: '10px', fontSize: '14px' }}>Large</p>
      </div>
    </div>
  ),
};

/**
 * Spinner with custom long text that wraps naturally.
 */
export const WithLongText: Story = {
  args: {
    size: 'md',
    text: 'This is a longer loading message that may wrap to multiple lines',
  },
};

/**
 * Demo of typical use case: full-page loading overlay.
 */
export const FullPageLoadingState: Story = {
  render: () => (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '400px',
        backgroundColor: '#f8fafc',
        borderRadius: '8px',
      }}
    >
      <LoadingSpinner size="lg" text="Loading TFDrift Falco Dashboard..." />
    </div>
  ),
};

/**
 * Demo of typical use case: inline loading state in a card.
 */
export const InlineLoadingState: Story = {
  render: () => (
    <div
      style={{
        padding: '20px',
        backgroundColor: '#f8fafc',
        borderRadius: '8px',
        border: '1px solid #e2e8f0',
        maxWidth: '400px',
      }}
    >
      <h3 style={{ margin: '0 0 20px 0' }}>Drift Analysis</h3>
      <LoadingSpinner size="sm" text="Analyzing resources..." />
    </div>
  ),
};

/**
 * Demo on dark background to show contrast.
 */
export const OnDarkBackground: Story = {
  parameters: {
    backgrounds: { default: 'dark' },
  },
  render: () => (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '300px',
        backgroundColor: '#0f172a',
        borderRadius: '8px',
      }}
    >
      <LoadingSpinner size="lg" text="Loading..." />
    </div>
  ),
};

/**
 * Spinner without text for minimal visual footprint.
 */
export const NoText: Story = {
  args: {
    size: 'md',
    text: undefined,
  },
};
