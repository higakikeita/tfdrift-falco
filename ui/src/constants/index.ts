/**
 * Barrel export for constants module.
 * Import from '@/constants' or '../constants' for clean imports.
 */
export {
  SEVERITY_BADGE_CLASSES,
  SEVERITY_DOT_COLORS,
  SEVERITY_CONFIG,
  SEVERITY_ORDER,
  SEVERITY_ICONS,
  PROVIDER_CONFIG,
  PROVIDER_FILTER_COLORS,
  TIME_INTERVALS,
  CHANGE_TYPE_CONFIG,
} from './severity';

export type {
  SeverityLevel,
  SeverityConfig,
  ProviderKey,
  ChangeType,
} from './severity';
