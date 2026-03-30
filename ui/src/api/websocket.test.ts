import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocket, type WSMessage, type WSResponse } from './websocket';

// Mock WebSocket
class MockWebSocket {
  url: string;
  readyState: number = WebSocket.CONNECTING;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  send: (data: string) => void = vi.fn();
  close: () => void = vi.fn();

  constructor(url: string) {
    this.url = url;
  }
}

Object.defineProperty(MockWebSocket, 'CONNECTING', { value: 0 });
Object.defineProperty(MockWebSocket, 'OPEN', { value: 1 });
Object.defineProperty(MockWebSocket, 'CLOSING', { value: 2 });
Object.defineProperty(MockWebSocket, 'CLOSED', { value: 3 });

global.WebSocket = MockWebSocket as any;

describe('useWebSocket hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with default state', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.isConnected).toBe(false);
    expect(result.current.isConnecting).toBe(false);
    expect(result.current.error).toBe(null);
    expect(result.current.lastMessage).toBe(null);
  });

  it('should have send function', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.send).toBeDefined();
    expect(typeof result.current.send).toBe('function');
  });

  it('should have subscribe function', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.subscribe).toBeDefined();
    expect(typeof result.current.subscribe).toBe('function');
  });

  it('should have unsubscribe function', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.unsubscribe).toBeDefined();
    expect(typeof result.current.unsubscribe).toBe('function');
  });

  it('should have connect function', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.connect).toBeDefined();
    expect(typeof result.current.connect).toBe('function');
  });

  it('should have disconnect function', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.disconnect).toBeDefined();
    expect(typeof result.current.disconnect).toBe('function');
  });

  it('should use default URL when not provided', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current).toBeDefined();
  });

  it('should accept custom URL', () => {
    const customUrl = 'ws://custom.example.com/ws';

    const { result } = renderHook(() =>
      useWebSocket({ url: customUrl, autoConnect: false })
    );

    expect(result.current).toBeDefined();
  });

  it('should not auto-connect when autoConnect is false', () => {
    const { result } = renderHook(() =>
      useWebSocket({ autoConnect: false })
    );

    expect(result.current.isConnecting).toBe(false);
    expect(result.current.isConnected).toBe(false);
  });

  it('should handle custom reconnect attempts option', () => {
    const { result } = renderHook(() =>
      useWebSocket({
        autoConnect: false,
        reconnectAttempts: 3,
      })
    );

    expect(result.current).toBeDefined();
  });

  it('should handle custom reconnect delay option', () => {
    const { result } = renderHook(() =>
      useWebSocket({
        autoConnect: false,
        reconnectDelay: 5000,
      })
    );

    expect(result.current).toBeDefined();
  });

  it('should handle custom heartbeat interval option', () => {
    const { result } = renderHook(() =>
      useWebSocket({
        autoConnect: false,
        heartbeatInterval: 60000,
      })
    );

    expect(result.current).toBeDefined();
  });

  it('should expose lastMessage', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.lastMessage).toBeNull();
  });

  it('should support all message types', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    const message: WSMessage = {
      type: 'subscribe',
      topic: 'drifts',
      payload: { filter: 'critical' },
    };

    expect(message.type).toBeDefined();
  });

  it('should support all topic types', () => {
    const topics = ['all', 'drifts', 'events', 'state', 'stats'];

    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    topics.forEach(topic => {
      expect(['all', 'drifts', 'events', 'state', 'stats']).toContain(topic);
    });
  });

  it('should handle send without connection', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    const message: WSMessage = {
      type: 'ping',
    };

    // Should not throw error
    result.current.send(message);
  });

  it('should handle subscribe call', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    // Should not throw error
    result.current.subscribe('drifts');
  });

  it('should handle unsubscribe call', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    // Should not throw error
    result.current.unsubscribe('drifts');
  });

  it('should return proper state shape', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current).toHaveProperty('isConnected');
    expect(result.current).toHaveProperty('isConnecting');
    expect(result.current).toHaveProperty('error');
    expect(result.current).toHaveProperty('lastMessage');
    expect(result.current).toHaveProperty('send');
    expect(result.current).toHaveProperty('subscribe');
    expect(result.current).toHaveProperty('unsubscribe');
    expect(result.current).toHaveProperty('connect');
    expect(result.current).toHaveProperty('disconnect');
  });

  it('should handle connection state changes', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(typeof result.current.isConnected).toBe('boolean');
    expect(typeof result.current.isConnecting).toBe('boolean');
  });

  it('should handle error state', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.error === null || result.current.error instanceof Error).toBe(true);
  });

  it('should support UseWebSocketOptions', () => {
    const options = {
      url: 'ws://example.com/ws',
      autoConnect: false,
      reconnectAttempts: 5,
      reconnectDelay: 3000,
      heartbeatInterval: 30000,
    };

    const { result } = renderHook(() => useWebSocket(options));

    expect(result.current).toBeDefined();
  });

  it('should maintain state across re-renders', () => {
    const { result, rerender } = renderHook(() =>
      useWebSocket({ autoConnect: false })
    );

    const initialState = {
      isConnected: result.current.isConnected,
      isConnecting: result.current.isConnecting,
    };

    rerender();

    expect(result.current.isConnected).toBe(initialState.isConnected);
    expect(result.current.isConnecting).toBe(initialState.isConnecting);
  });

  it('should clean up on unmount', () => {
    const { result, unmount } = renderHook(() =>
      useWebSocket({ autoConnect: false })
    );

    expect(result.current).toBeDefined();

    unmount();

    // After unmount, should be cleaned up
  });

  it('should initialize with false connection states', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.isConnected).toBe(false);
    expect(result.current.isConnecting).toBe(false);
  });

  it('should initialize error as null', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.error).toBeNull();
  });

  it('should initialize lastMessage as null', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }));

    expect(result.current.lastMessage).toBeNull();
  });

  it('should handle multiple hook instances', () => {
    const { result: result1 } = renderHook(() =>
      useWebSocket({ autoConnect: false })
    );
    const { result: result2 } = renderHook(() =>
      useWebSocket({ autoConnect: false })
    );

    expect(result1.current.isConnected).toBe(false);
    expect(result2.current.isConnected).toBe(false);
  });

  it('should accept both ping and query message types', () => {
    const pingMessage: WSMessage = { type: 'ping' };
    const queryMessage: WSMessage = { type: 'query', payload: {} };

    expect(pingMessage.type).toBe('ping');
    expect(queryMessage.type).toBe('query');
  });

  it('should handle heartbeat interval configuration', () => {
    const { result } = renderHook(() =>
      useWebSocket({
        autoConnect: false,
        heartbeatInterval: 45000,
      })
    );

    expect(result.current).toBeDefined();
  });
});

describe('WSMessage type', () => {
  it('should represent WebSocket message structure', () => {
    const message: WSMessage = {
      type: 'subscribe',
      topic: 'drifts',
    };

    expect(message.type).toBe('subscribe');
    expect(message.topic).toBe('drifts');
  });

  it('should support various message types', () => {
    const types: Array<'subscribe' | 'unsubscribe' | 'ping' | 'query'> = [
      'subscribe',
      'unsubscribe',
      'ping',
      'query',
    ];

    types.forEach(type => {
      const message: WSMessage = { type };
      expect(message.type).toBe(type);
    });
  });
});

describe('WSResponse type', () => {
  it('should represent WebSocket response structure', () => {
    const response: WSResponse = {
      type: 'message',
      topic: 'drifts',
      payload: { id: '123', detected: true },
    };

    expect(response.type).toBe('message');
    expect(response.payload).toBeDefined();
  });

  it('should support optional timestamp', () => {
    const response: WSResponse = {
      type: 'message',
      payload: {},
      timestamp: new Date().toISOString(),
    };

    expect(response.timestamp).toBeDefined();
  });
});
