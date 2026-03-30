import type { Meta, StoryObj } from '@storybook/react';
import { ErrorBoundary } from './ErrorBoundary';

const meta: Meta<typeof ErrorBoundary> = {
  title: 'Components/ErrorBoundary',
  component: ErrorBoundary,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
    backgrounds: {
      default: 'light',
      values: [
        { name: 'light', value: '#ffffff' },
      ],
    },
  },
};

export default meta;
type Story = StoryObj<typeof ErrorBoundary>;

/**
 * Component that throws an error to demonstrate error boundary catching.
 */
const ErrorThrowingComponent = () => {
  throw new Error('This is a demonstration error from the component');
};

/**
 * Safe component that renders without errors.
 */
const SafeComponent = () => (
  <div style={{ padding: '20px' }}>
    <h2>This is a safe component</h2>
    <p>It renders without any errors.</p>
  </div>
);

/**
 * Default ErrorBoundary showing the error UI.
 * This story demonstrates what happens when a child component throws an error.
 */
export const CaughtError: Story = {
  render: () => (
    <ErrorBoundary>
      <ErrorThrowingComponent />
    </ErrorBoundary>
  ),
};

/**
 * ErrorBoundary wrapping safe children that render normally.
 */
export const WithSafeChildren: Story = {
  render: () => (
    <ErrorBoundary>
      <SafeComponent />
    </ErrorBoundary>
  ),
};

/**
 * Multiple safe components within ErrorBoundary.
 */
export const WithMultipleSafeChildren: Story = {
  render: () => (
    <ErrorBoundary>
      <div style={{ padding: '20px' }}>
        <h1>Application Dashboard</h1>
        <p>This is the main content area.</p>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(2, 1fr)',
          gap: '20px',
          marginTop: '20px',
        }}>
          <div style={{
            padding: '15px',
            backgroundColor: '#f0f4f8',
            borderRadius: '8px',
          }}>
            <h3>Panel 1</h3>
            <p>Some content here</p>
          </div>
          <div style={{
            padding: '15px',
            backgroundColor: '#f0f4f8',
            borderRadius: '8px',
          }}>
            <h3>Panel 2</h3>
            <p>Some content here</p>
          </div>
        </div>
      </div>
    </ErrorBoundary>
  ),
};

/**
 * ErrorBoundary with a complex UI layout that catches errors.
 */
export const ComplexLayout: Story = {
  render: () => (
    <ErrorBoundary>
      <div style={{
        display: 'flex',
        height: '100vh',
      }}>
        <aside style={{
          width: '250px',
          backgroundColor: '#f8fafc',
          borderRight: '1px solid #e2e8f0',
          padding: '20px',
        }}>
          <h2 style={{ marginTop: 0 }}>Navigation</h2>
          <ul style={{ listStyle: 'none', padding: 0 }}>
            <li><a href="#" style={{ textDecoration: 'none', color: '#0066cc' }}>Dashboard</a></li>
            <li><a href="#" style={{ textDecoration: 'none', color: '#0066cc' }}>Settings</a></li>
            <li><a href="#" style={{ textDecoration: 'none', color: '#0066cc' }}>Reports</a></li>
          </ul>
        </aside>
        <main style={{ flex: 1, padding: '20px' }}>
          <h1>Welcome to TFDrift Falco</h1>
          <p>This is the main content area protected by ErrorBoundary.</p>
        </main>
      </div>
    </ErrorBoundary>
  ),
};

/**
 * Demonstrating ErrorBoundary protecting different sections of an application.
 * This shows a realistic scenario where one component fails but others continue.
 */
export const MultipleErrorBoundaries: Story = {
  render: () => (
    <div style={{ padding: '20px' }}>
      <h1>Application with Multiple Error Boundaries</h1>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: '20px',
        marginTop: '20px',
      }}>
        <div>
          <h3>Widget A (Safe)</h3>
          <ErrorBoundary>
            <SafeComponent />
          </ErrorBoundary>
        </div>

        <div>
          <h3>Widget B (Error - Isolated)</h3>
          <ErrorBoundary>
            <ErrorThrowingComponent />
          </ErrorBoundary>
        </div>

        <div>
          <h3>Widget C (Safe)</h3>
          <ErrorBoundary>
            <SafeComponent />
          </ErrorBoundary>
        </div>

        <div>
          <h3>Widget D (Safe)</h3>
          <ErrorBoundary>
            <SafeComponent />
          </ErrorBoundary>
        </div>
      </div>
    </div>
  ),
};

/**
 * Detailed error view showing what users would see in development mode.
 */
export const DevelopmentMode: Story = {
  parameters: {
    docs: {
      description: {
        story: 'In development mode, the error boundary shows detailed error information including the error message and component stack trace.',
      },
    },
  },
  render: () => (
    <div style={{ position: 'relative' }}>
      <ErrorBoundary>
        <ErrorThrowingComponent />
      </ErrorBoundary>
      <p style={{
        marginTop: '20px',
        padding: '10px',
        backgroundColor: '#fff3cd',
        border: '1px solid #ffc107',
        borderRadius: '4px',
        fontSize: '12px',
      }}>
        Note: In development mode, you'll see the full error message and component stack trace.
        In production, users see a friendly error message.
      </p>
    </div>
  ),
};

/**
 * Error boundary demonstrating the reload functionality.
 * Users can click the "Reload Page" button to recover from the error.
 */
export const WithReloadButton: Story = {
  render: () => (
    <div style={{
      padding: '20px',
    }}>
      <h2>Error Recovery Example</h2>
      <p>When an error occurs, users can click the "Reload Page" button to recover:</p>
      <ErrorBoundary>
        <ErrorThrowingComponent />
      </ErrorBoundary>
    </div>
  ),
};

/**
 * Shows how ErrorBoundary protects async component initialization.
 */
export const AsyncComponentProtection: Story = {
  render: () => (
    <ErrorBoundary>
      <div style={{ padding: '20px' }}>
        <h3>Protected Async Operations</h3>
        <p>ErrorBoundary can catch errors during component initialization and rendering.</p>
        <div style={{
          padding: '10px',
          backgroundColor: '#e7f3ff',
          border: '1px solid #b3d9ff',
          borderRadius: '4px',
          marginTop: '10px',
        }}>
          Note: ErrorBoundary catches errors during render phase only.
          For async errors (setTimeout, promises), use try-catch or error state management.
        </div>
      </div>
    </ErrorBoundary>
  ),
};

/**
 * Full page error state demonstrating what users see when the entire page fails.
 */
export const FullPageError: Story = {
  render: () => (
    <ErrorBoundary>
      <ErrorThrowingComponent />
    </ErrorBoundary>
  ),
};

/**
 * Demonstrating that ErrorBoundary preserves the application structure
 * and only shows an error dialog over the content.
 */
export const ErrorOverlay: Story = {
  render: () => (
    <div style={{
      position: 'relative',
      minHeight: '600px',
      backgroundColor: '#f8fafc',
      padding: '20px',
    }}>
      <h1>Application Page</h1>
      <p>Some content that would normally be here...</p>

      <ErrorBoundary>
        <div style={{
          padding: '20px',
          backgroundColor: 'white',
          borderRadius: '8px',
          marginTop: '20px',
          border: '1px solid #e2e8f0',
        }}>
          <ErrorThrowingComponent />
        </div>
      </ErrorBoundary>
    </div>
  ),
};
