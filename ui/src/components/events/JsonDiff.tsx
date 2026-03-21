/**
 * JsonDiff Component
 * Renders an inline diff view for JSON/string values
 * with syntax-highlighted additions and removals.
 */

import { useMemo } from 'react';
import { cn } from '../../lib/utils';

interface JsonDiffProps {
  oldValue: unknown;
  newValue: unknown;
  attribute?: string;
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return '(none)';
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  if (typeof value === 'object') return JSON.stringify(value, null, 2);
  return String(value);
}

function computeLineDiff(
  oldStr: string,
  newStr: string
): Array<{ type: 'added' | 'removed' | 'context'; text: string }> {
  const oldLines = oldStr.split('\n');
  const newLines = newStr.split('\n');
  const m = oldLines.length;
  const n = newLines.length;

  if (m + n > 200) {
    return [
      ...oldLines.map((t) => ({ type: 'removed' as const, text: t })),
      ...newLines.map((t) => ({ type: 'added' as const, text: t })),
    ];
  }

  const dp: number[][] = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0));
  for (let i = 1; i <= m; i++)
    for (let j = 1; j <= n; j++)
      dp[i][j] = oldLines[i - 1] === newLines[j - 1]
        ? dp[i - 1][j - 1] + 1
        : Math.max(dp[i - 1][j], dp[i][j - 1]);

  const diff: Array<{ type: 'added' | 'removed' | 'context'; text: string }> = [];
  let i = m, j = n;
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      diff.unshift({ type: 'context', text: oldLines[i - 1] });
      i--; j--;
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      diff.unshift({ type: 'added', text: newLines[j - 1] });
      j--;
    } else {
      diff.unshift({ type: 'removed', text: oldLines[i - 1] });
      i--;
    }
  }
  return diff;
}

export function JsonDiff({ oldValue, newValue, attribute }: JsonDiffProps) {
  const diff = useMemo(
    () => computeLineDiff(formatValue(oldValue), formatValue(newValue)),
    [oldValue, newValue]
  );

  return (
    <div className="rounded-lg border border-slate-200 overflow-hidden">
      {attribute && (
        <div className="px-3 py-2 bg-slate-50 border-b border-slate-200">
          <span className="text-xs font-medium text-slate-500">Attribute: </span>
          <code className="text-xs font-mono text-indigo-600">{attribute}</code>
        </div>
      )}
      <div className="overflow-x-auto">
        <pre className="text-xs font-mono leading-relaxed p-0 m-0">
          {diff.map((line, idx) => (
            <div
              key={idx}
              className={cn(
                'px-3 py-0.5 border-l-2',
                line.type === 'removed' && 'bg-red-50 border-l-red-400 text-red-800',
                line.type === 'added' && 'bg-green-50 border-l-green-400 text-green-800',
                line.type === 'context' && 'bg-white border-l-transparent text-slate-600'
              )}
            >
              <span className="select-none text-slate-400 mr-2 inline-block w-4 text-right">
                {line.type === 'removed' ? '-' : line.type === 'added' ? '+' : ' '}
              </span>
              {line.text}
            </div>
          ))}
          {diff.every((d) => d.type === 'context') && (
            <div className="px-3 py-2 text-slate-400 italic">No changes detected</div>
          )}
        </pre>
      </div>
    </div>
  );
}
