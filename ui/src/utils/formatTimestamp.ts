import { TIME_INTERVALS } from '../constants';

/**
 * Format a timestamp into a human-readable relative time string (Japanese).
 * Returns relative time for recent events, or formatted date for older ones.
 *
 * @param timestamp - ISO 8601 timestamp string
 * @returns Formatted time string in Japanese
 */
export function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / TIME_INTERVALS.MINUTE);
  const hours = Math.floor(diff / TIME_INTERVALS.HOUR);
  const days = Math.floor(diff / TIME_INTERVALS.DAY);

  if (minutes < 1) return 'たった今';
  if (minutes < 60) return `${minutes}分前`;
  if (hours < 24) return `${hours}時間前`;
  if (days < 7) return `${days}日前`;
  return date.toLocaleDateString('ja-JP', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}
