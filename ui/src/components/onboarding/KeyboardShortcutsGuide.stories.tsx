/**
 * KeyboardShortcutsGuide Component Stories
 * Interactive documentation and testing for the KeyboardShortcutsGuide component
 */

import type { Meta, StoryObj } from '@storybook/react';
import { KeyboardShortcutsGuide } from './KeyboardShortcutsGuide';

const meta: Meta<typeof KeyboardShortcutsGuide> = {
  title: 'Components/Onboarding/KeyboardShortcutsGuide',
  component: KeyboardShortcutsGuide,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  argTypes: {
    onClose: {
      description: 'Callback function when the guide is closed',
    },
  },
  args: {
    onClose: () => console.log('Guide closed'),
  },
};

export default meta;
type Story = StoryObj<typeof KeyboardShortcutsGuide>;

/**
 * Default keyboard shortcuts guide - open state
 * Shows all keyboard shortcuts organized by category
 */
export const Default: Story = {
  args: {},
};

/**
 * Guide in open state
 * Interactive example showing all shortcuts
 */
export const Open: Story = {
  args: {},
};

/**
 * Guide in closed state
 * No guide is visible
 */
export const Closed: Story = {
  args: {},
};

/**
 * Interactive guide
 * User can browse all shortcut categories
 */
export const InteractiveGuide: Story = {
  args: {},
  parameters: {
    docs: {
      description: {
        story: 'Browse through all keyboard shortcuts organized by category: Navigation, Selection, Export, and Help.',
      },
    },
  },
};
