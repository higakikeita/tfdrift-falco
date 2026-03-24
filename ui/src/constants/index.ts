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

export {
  AWS_COMPUTE_COLORS,
  AWS_DATABASE_COLORS,
  AWS_STORAGE_COLORS,
  AWS_NETWORK_COLORS,
  AWS_SECURITY_COLORS,
  AWS_INTEGRATION_COLORS,
  AWS_SERVICE_COLORS,
  AWS_SERVICE_LEGEND,
  DRIFT_STATUS_COLORS,
  DRIFT_STATUS_BORDER_CLASSES,
  DRIFT_STATUS_LEGEND,
  CYTOSCAPE_DRIFT_COLORS,
  NODE_TYPE_COLORS,
  PROVIDER_COLORS,
} from './colors';

export type {
  LegendEntry,
  LegendCategory,
  DriftLegendEntry,
} from './colors';
