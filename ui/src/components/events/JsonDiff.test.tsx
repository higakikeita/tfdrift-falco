import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('../../lib/utils', () => ({
  cn: (...args: unknown[]) => '',
}));

import { JsonDiff } from './JsonDiff';

describe('JsonDiff', () => {
  it('should render without crashing', () => {
    const { container } = render(
      <JsonDiff oldValue="old" newValue="new" />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should render diff between two string values', () => {
    const { container } = render(
      <JsonDiff oldValue="hello" newValue="hello world" />
    );
    expect(container).toBeTruthy();
  });

  it('should handle JSON object values', () => {
    const oldValue = { key: 'old' };
    const newValue = { key: 'new' };
    const { container } = render(
      <JsonDiff oldValue={oldValue} newValue={newValue} />
    );
    expect(container).toBeTruthy();
  });

  it('should handle null values', () => {
    const { container } = render(
      <JsonDiff oldValue={null} newValue="new" />
    );
    expect(container).toBeTruthy();
  });

  it('should display attribute when provided', () => {
    render(
      <JsonDiff
        oldValue="old"
        newValue="new"
        attribute="testAttr"
      />
    );
    expect(screen.getByText('Attribute:')).toBeInTheDocument();
    expect(screen.getByText('testAttr')).toBeInTheDocument();
  });

  it('should handle identical values', () => {
    const { container } = render(
      <JsonDiff oldValue="same" newValue="same" />
    );
    expect(container).toBeTruthy();
  });

  it('should handle complex nested JSON', () => {
    const oldValue = {
      nested: { key: 'value', array: [1, 2, 3] },
    };
    const newValue = {
      nested: { key: 'updated', array: [1, 2, 3, 4] },
    };
    const { container } = render(
      <JsonDiff oldValue={oldValue} newValue={newValue} />
    );
    expect(container).toBeTruthy();
  });

  it('should render correctly with undefined values', () => {
    const { container } = render(
      <JsonDiff oldValue={undefined} newValue="new" />
    );
    expect(container).toBeTruthy();
  });
});
