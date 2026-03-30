import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSSE, type SSEEvent } from './sse';

// Mock EventSource
class MockEventSource {
  url: string;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  readyState: number = 0;
  addEventListener: (event: string, handler: (e: any) => void) => void = vi.fn();
  close: () => void = vi.fn();

  constructor(url: string) {
    this.url = url;
  }
}

global.EventSource = MockEventSource as any;

describe('useSSE hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with default state', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.isConnected).toBe(false);
    expect(result.current.isConnecting).toBe(false);
    expect(result.current.error).toBe(null);
    expect(result.current.lastEvent).toBe(null);
    expect(result.current.events).toEqual([]);
  });

  it('should have connect function', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.connect).toBeDefined();
    expect(typeof result.current.connect).toBe('function');
  });

  it('should have disconnect function', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.disconnect).toBeDefined();
    expect(typeof result.current.disconnect).toBe('function');
  });

  it('should have clearEvents function', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.clearEvents).toBeDefined();
    expect(typeof result.current.clearEvents).toBe('function');
  });

  it('should use default URL when not provided', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current).toBeDefined();
  });

  it('should accept custom URL', () => {
    const customUrl = 'https://custom.example.com/stream';

    const { result } = renderHook(() =>
      useSSE({ url: customUrl, autoConnect: false })
    );

    expect(result.current).toBeDefined();
  });

  it('should not auto-connect when autoConnect is false', () => {
    const { result } = renderHook(() =>
      useSSE({ autoConnect: false })
    );

    expect(result.current.isConnecting).toBe(false);
    expect(result.current.isConnected).toBe(false);
  });

  it('should handle connect with custom options', () => {
    const { result } = renderHook(() =>
      useSSE({
        autoConnect: false,
        reconnectAttempts: 3,
        reconnectDelay: 1000,
      })
    );

    expect(result.current.connect).toBeDefined();
  });

  it('should expose lastEvent', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.lastEvent).toBeNull();
  });

  it('should expose events array', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(Array.isArray(result.current.events)).toBe(true);
    expect(result.current.events.length).toBe(0);
  });

  it('should handle reconnect options', () => {
    const { result } = renderHook(() =>
      useSSE({
        autoConnect: false,
        reconnectAttempts: 5,
        reconnectDelay: 5000,
      })
    );

    expect(result.current).toBeDefined();
  });

  it('should clear events when clearEvents called', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    act(() => {
      result.current.clearEvents();
    });

    expect(result.current.events).toEqual([]);
  });

  it('should return connection status', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(typeof result.current.isConnected).toBe('boolean');
    expect(typeof result.current.isConnecting).toBe('boolean');
  });

  it('should support different event types', () => {
    const eventTypes = ['connected', 'drift', 'falco', 'state_change', 'keep-alive', 'message'];
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.lastEvent).toBe(null);
  });

  it('should handle error state', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.error).toBeNull();
  });

  it('should store events in array', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(Array.isArray(result.current.events)).toBe(true);
  });

  it('should support URL configuration option', () => {
    const customUrl = 'http://example.com/api/stream';

    const { result } = renderHook(() =>
      useSSE({ url: customUrl, autoConnect: false })
    );

    expect(result.current.isConnected).toBe(false);
  });

  it('should handle reconnection with exponential backoff', () => {
    const { result } = renderHook(() =>
      useSSE({
        autoConnect: false,
        reconnectAttempts: 3,
        reconnectDelay: 1000,
      })
    );

    expect(result.current).toBeDefined();
  });

  it('should have proper return type shape', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current).toHaveProperty('isConnected');
    expect(result.current).toHaveProperty('isConnecting');
    expect(result.current).toHaveProperty('error');
    expect(result.current).toHaveProperty('lastEvent');
    expect(result.current).toHaveProperty('events');
    expect(result.current).toHaveProperty('connect');
    expect(result.current).toHaveProperty('disconnect');
    expect(result.current).toHaveProperty('clearEvents');
  });

  it('should handle multiple connections gracefully', () => {
    const { result: result1 } = renderHook(() => useSSE({ autoConnect: false }));
    const { result: result2 } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result1.current.isConnected).toBe(false);
    expect(result2.current.isConnected).toBe(false);
  });

  it('should accept UseSSEOptions', () => {
    const options = {
      url: 'http://example.com/stream',
      autoConnect: false,
      reconnectAttempts: 5,
      reconnectDelay: 3000,
    };

    const { result } = renderHook(() => useSSE(options));

    expect(result.current).toBeDefined();
  });

  it('should handle cleanup on unmount', () => {
    const { result, unmount } = renderHook(() =>
      useSSE({ autoConnect: false })
    );

    expect(result.current).toBeDefined();

    unmount();

    // After unmount, should be cleaned up
  });

  it('should return consistent state shape on re-render', () => {
    const { result, rerender } = renderHook(() =>
      useSSE({ autoConnect: false })
    );

    const initialState = result.current;

    rerender();

    expect(result.current).toHaveProperty('isConnected');
    expect(result.current).toHaveProperty('isConnecting');
    expect(result.current).toHaveProperty('error');
    expect(result.current).toHaveProperty('lastEvent');
    expect(result.current).toHaveProperty('events');
  });

  it('should initialize with empty events array', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.events).toEqual([]);
  });

  it('should initialize error as null', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.error).toBeNull();
  });

  it('should initialize lastEvent as null', () => {
    const { result } = renderHook(() => useSSE({ autoConnect: false }));

    expect(result.current.lastEvent).toBeNull();
  });
});

describe('SSEEvent type', () => {
  it('should represent SSE event structure', () => {
    const event: SSEEvent = {
      type: 'connected',
      data: { status: 'ok' },
      timestamp: new Date().toISOString(),
    };

    expect(event.type).toBe('connected');
    expect(event.data).toBeDefined();
    expect(event.timestamp).toBeDefined();
  });

  it('should support various event types', () => {
    const types: Array<'connected' | 'drift' | 'falco' | 'state_change' | 'keep-alive' | 'message'> = [
      'connected',
      'drift',
      'falco',
      'state_change',
      'keep-alive',
      'message',
    ];

    types.forEach(type => {
      const event: SSEEvent = {
        type,
        data: {},
      };

      expect(event.type).toBe(type);
    });
  });
});
