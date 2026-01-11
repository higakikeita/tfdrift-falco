/**
 * WelcomeModal Component Stories
 * Interactive documentation and testing for the WelcomeModal component
 */

import type { Meta, StoryObj } from '@storybook/react';
import { WelcomeModal } from './WelcomeModal';

const meta: Meta<typeof WelcomeModal> = {
  title: 'Components/Onboarding/WelcomeModal',
  component: WelcomeModal,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  argTypes: {
    onClose: {
      description: 'Callback function when the modal is closed',
    },
  },
  args: {
    onClose: () => console.log('Modal closed'),
  },
};

export default meta;
type Story = StoryObj<typeof WelcomeModal>;

/**
 * Default welcome modal - Step 1
 * Shows the welcome screen with project overview
 */
export const Default: Story = {
  args: {
  },
};

/**
 * Modal in open state
 * Interactive example showing the full tutorial flow
 */
export const Open: Story = {
  args: {
  },
};

/**
 * Modal in closed state
 * No modal is visible
 */
export const Closed: Story = {
  args: {
  },
};

/**
 * Interactive tutorial flow
 * User can navigate through all steps
 */
export const InteractiveTutorial: Story = {
  args: {
  },
  parameters: {
    docs: {
      description: {
        story: 'Click "次へ" to navigate through the tutorial steps, or "戻る" to go back.',
      },
    },
  },
};
