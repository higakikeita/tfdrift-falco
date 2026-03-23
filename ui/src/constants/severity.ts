/**
 * Shared severity constants used across the application.
 * Centralizes severity-related configuration to avoid duplication
 * in SettingsPage, DriftHistoryTable, and other components.
 */

export type SeverityLevel = 'critical' | 'high' | 'medium' | 'low';

export interface SeverityConfig {
  label: string;
  badgeClass: string;
  dotColor: string;
  bgColor: string;
  textColor: string;
}

/**
 * Severity badge Tailwind classes (light + dark mode).
 * Used in SettingsPage RulesTab, DriftHistoryTable, DriftDetailPanel, etc.
 */
export const SEVERITY_BADGE_CLASSES: Record<SeverityLevel, string> = {
  critical: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  medium: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  low: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
};

/**
 * Severity dot/indicator colors for charts and visual indicators.
 */
export const SEVERITY_DOT_COLORS: Record<SeverityLevel, string> = {
  critical: 'bg-red-500',
  high: 'bg-orange-500',
  medium: 'bg-yellow-500',
  low: 'bg-blue-500',
};

/**
 * Full severity configuration with all visual properties.
 */
export const SEVERITY_CONFIG: Record<SeverityLevel, SeverityConfig> = {
  critical: {
    label: 'Critical',
    badgeClass: SEVERITY_BADGE_CLASSES.critical,
    dotColor: SEVERITY_DOT_COLORS.critical,
    bgColor: '#ef4444',
    textColor: '#991b1b',
  },
  high: {
    label: 'High',
    badgeClass: SEVERITY_BADGE_CLASSES.high,
    dotColor: SEVERITY_DOT_COLORS.high,
    bgColor: '#f97316',
    textColor: '#9a3412',
  },
  medium: {
    label: 'Medium',
    badgeClass: SEVERITY_BADGE_CLASSES.medium,
    dotColor: SEVERITY_DOT_COLORS.medium,
    bgColor: '#eab308',
    textColor: '#854d0e',
  },
  low: {
    label: 'Low',
    badgeClass: SEVERITY_BADGE_CLASSES.low,
    dotColor: SEVERITY_DOT_COLORS.low,
    bgColor: '#3b82f6',
    textColor: '#1e40af',
  },
};

/**
 * Ordered severity levels from most to least severe.
 */
export const SEVERITY_ORDER: SeverityLevel[] = ['critical', 'high', 'medium', 'low'];

/**
 * Cloud provider badge configuration.
 * Used in SettingsPage ProvidersTab and DriftHistoryTable.
 */
export const PROVIDER_CONFIG = {
  aws: {
    name: 'Amazon Web Services',
    shortName: 'AWS',
    badgeClass: 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200',
  },
  gcp: {
    name: 'Google Cloud Platform',
    shortName: 'GCP',
    badgeClass: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  },
  azure: {
    name: 'Microsoft Azure',
    shortName: 'Azure',
    badgeClass: 'bg-sky-100 text-sky-800 dark:bg-sky-900 dark:text-sky-200',
  },
} as const;

export type ProviderKey = keyof typeof PROVIDER_CONFIG;

/**
 * Common time interval constants (in milliseconds).
 * Replaces magic numbers like 60000, 3600000, 86400000.
 */
export const TIME_INTERVALS = {
  SECOND: 1000,
  MINUTE: 60 * 1000,
  HOUR: 60 * 60 * 1000,
  DAY: 24 * 60 * 60 * 1000,
  WEEK: 7 * 24 * 60 * 60 * 1000,
} as const;

/**
 * Severity emoji icons for visual indicators in tables and lists.
 */
export const SEVERITY_ICONS: Record<SeverityLevel, string> = {
  critical: '\u{1F6A8}',
  high: '\u26A0\uFE0F',
  medium: '\u26A1',
  low: '\u2139\uFE0F',
};

/**
 * Drift change type configuration.
 * Used in DriftHistoryTable and drift-related components.
 */
export type ChangeType = 'added' | 'modified' | 'deleted';

export const CHANGE_TYPE_CONFIG: Record<ChangeType, { label: string; color: string }> = {
  added: { label: '\u8FFD\u52A0', color: 'text-green-600' },
  modified: { label: '\u5909\u66F4', color: 'text-yellow-600' },
  deleted: { label: '\u524A\u9664', color: 'text-red-600' },
};

/**
 * Provider filter button colors for active state.
 */
export const PROVIDER_FILTER_COLORS: Record<string, { active: string; inactive: string }> = {
  aws: { active: 'bg-orange-100 text-orange-800 border-orange-200', inactive: 'bg-white text-gray-600 border-gray-300' },
  gcp: { active: 'bg-blue-100 text-blue-800 border-blue-200', inactive: 'bg-white text-gray-600 border-gray-300' },
  azure: { active: 'bg-sky-100 text-sky-800 border-sky-200', inactive: 'bg-white text-gray-600 border-gray-300' },
};
