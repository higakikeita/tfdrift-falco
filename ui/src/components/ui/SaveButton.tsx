import type { LucideIcon } from 'lucide-react';
import { Save } from 'lucide-react';

interface SaveButtonProps {
  onClick: () => void;
  label?: string;
  icon?: LucideIcon;
  disabled?: boolean;
}

/**
 * Reusable save/action button with consistent styling.
 * Extracted from SettingsPage where it was duplicated across all 4 tabs.
 */
export function SaveButton({
  onClick,
  label = 'Save',
  icon: Icon = Save,
  disabled = false,
}: SaveButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="inline-flex items-center gap-1.5 px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors"
    >
      <Icon className="h-4 w-4" />
      {label}
    </button>
  );
}
