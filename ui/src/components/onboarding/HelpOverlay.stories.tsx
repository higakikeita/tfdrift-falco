/**
 * HelpOverlay Component Stories
 * Interactive documentation and testing for the HelpOverlay component
 */

import type { Meta, StoryObj } from '@storybook/react';
import { HelpOverlay } from './HelpOverlay';

const meta: Meta<typeof HelpOverlay> = {
  title: 'Components/Onboarding/HelpOverlay',
  component: HelpOverlay,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  argTypes: {
    onOpenShortcuts: {
      description: 'Callback to open keyboard shortcuts guide',
    },
    onOpenWelcome: {
      description: 'Callback to open welcome tutorial',
    },
  },
  args: {
    onOpenShortcuts: () => console.log('Open shortcuts'),
    onOpenWelcome: () => console.log('Open welcome'),
  },
};

export default meta;
type Story = StoryObj<typeof HelpOverlay>;

/**
 * Default help overlay - expanded state
 * Shows quick tips and key features
 */
export const Default: Story = {
  args: {},
};

/**
 * Help overlay with all callbacks
 * Fully interactive with all buttons functional
 */
export const FullyInteractive: Story = {
  args: {
    onOpenShortcuts: () => console.log('Open shortcuts'),
    onOpenWelcome: () => console.log('Open welcome'),
  },
};

/**
 * Help overlay without shortcut callback
 * Keyboard shortcuts button will not be shown
 */
export const WithoutShortcuts: Story = {
  args: {
    onOpenShortcuts: undefined,
    onOpenWelcome: () => console.log('Open welcome'),
  },
};

/**
 * Help overlay without welcome callback
 * Tutorial button will not be shown
 */
export const WithoutTutorial: Story = {
  args: {
    onOpenShortcuts: () => console.log('Open shortcuts'),
    onOpenWelcome: undefined,
  },
};

/**
 * Minimal help overlay
 * No action buttons, only tips and features
 */
export const Minimal: Story = {
  args: {
    onOpenShortcuts: undefined,
    onOpenWelcome: undefined,
  },
};
