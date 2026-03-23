/**
 * Barrel export for constants module.
 * Import from '@/constants' or '../constants' for clean imports.
 */
export {
  SEVERITY_BADGE_CLASSES,
  SEVERITY_DOT_COLORS,
  SEVERITY_CONFIG,
  SEVERITY_ORDER,
  PROVIDER_CONFIG,
  TIME_INTERVALS,
} from './severity';

export type {
  SeverityLevel,
  SeverityConfig,
  ProviderKey,
} from './severity';
