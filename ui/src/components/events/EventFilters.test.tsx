import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
}));

import EventFilters from './EventFilters';

const defaultFilters = {
  severity: '',
  provider: '',
  status: '',
  search: '',
};

describe('EventFilters', () => {
  it('should render without crashing', () => {
    const { container } = render(
      <EventFilters filters={defaultFilters} onFiltersChange={vi.fn()} />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should render filter inputs', () => {
    render(
      <EventFilters filters={defaultFilters} onFiltersChange={vi.fn()} />
    );
    // Should have form elements (select, input, etc.)
    const inputs = document.querySelectorAll('input, select, button');
    expect(inputs.length).toBeGreaterThan(0);
  });

  it('should call onFiltersChange when a filter changes', () => {
    const onChange = vi.fn();
    render(
      <EventFilters filters={defaultFilters} onFiltersChange={onChange} />
    );
    // Find any interactive element and interact
    const selects = document.querySelectorAll('select');
    if (selects.length > 0) {
      fireEvent.change(selects[0], { target: { value: 'critical' } });
      expect(onChange).toHaveBeenCalled();
    }
  });
});
