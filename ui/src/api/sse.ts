/**
 * Server-Sent Events (SSE) Client
 * Unidirectional real-time streaming from TFDrift API
 */

import { useEffect, useRef, useState, useCallback } from 'react';

const SSE_URL = import.meta.env.VITE_SSE_URL || 'http://localhost:8080/api/v1/stream';

export type SSEEventType = 'connected' | 'drift' | 'falco' | 'state_change' | 'keep-alive';

export interface SSEEvent {
  type: SSEEventType;
  data: any;
  timestamp?: string;
}

export interface UseSSEOptions {
  url?: string;
  autoConnect?: boolean;
  reconnectAttempts?: number;
  reconnectDelay?: number;
}

export interface UseSSEReturn {
  isConnected: boolean;
  isConnecting: boolean;
  error: Error | null;
  lastEvent: SSEEvent | null;
  events: SSEEvent[];
  connect: () => void;
  disconnect: () => void;
  clearEvents: () => void;
}

export function useSSE(options: UseSSEOptions = {}): UseSSEReturn {
  const {
    url = SSE_URL,
    autoConnect = true,
    reconnectAttempts = 5,
    reconnectDelay = 3000,
  } = options;

  const eventSource = useRef<EventSource | null>(null);
  const reconnectCount = useRef(0);
  const reconnectTimeout = useRef<NodeJS.Timeout | null>(null);

  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastEvent, setLastEvent] = useState<SSEEvent | null>(null);
  const [events, setEvents] = useState<SSEEvent[]>([]);

  // Schedule reconnect with exponential backoff
  const scheduleReconnect = useCallback(() => {
    if (reconnectCount.current >= reconnectAttempts) {
      setError(new Error(`Failed to connect after ${reconnectAttempts} attempts`));
      setIsConnecting(false);
      return;
    }

    reconnectCount.current++;
    const delay = reconnectDelay * Math.pow(2, reconnectCount.current - 1);

    console.log(`[SSE] Reconnecting in ${delay}ms (attempt ${reconnectCount.current}/${reconnectAttempts})`);

    reconnectTimeout.current = setTimeout(() => {
      connect();
    }, delay);
  }, [reconnectAttempts, reconnectDelay]);

  // Handle incoming event
  const handleEvent = useCallback((type: SSEEventType, data: any) => {
    const event: SSEEvent = {
      type,
      data,
      timestamp: new Date().toISOString(),
    };

    console.log('[SSE] Event received:', event);
    setLastEvent(event);
    setEvents(prev => [...prev, event].slice(-100)); // Keep last 100 events
  }, []);

  // Connect to SSE stream
  const connect = useCallback(() => {
    // Prevent multiple connections
    if (eventSource.current || isConnecting) {
      return;
    }

    setIsConnecting(true);
    setError(null);

    try {
      console.log('[SSE] Connecting to', url);
      const es = new EventSource(url);

      es.onopen = () => {
        console.log('[SSE] Connected');
        setIsConnected(true);
        setIsConnecting(false);
        setError(null);
        reconnectCount.current = 0;
      };

      // Handle 'connected' event
      es.addEventListener('connected', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          handleEvent('connected', data);
        } catch (err) {
          console.error('[SSE] Failed to parse connected event:', err);
        }
      });

      // Handle 'drift' events
      es.addEventListener('drift', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          handleEvent('drift', data);
        } catch (err) {
          console.error('[SSE] Failed to parse drift event:', err);
        }
      });

      // Handle 'falco' events
      es.addEventListener('falco', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          handleEvent('falco', data);
        } catch (err) {
          console.error('[SSE] Failed to parse falco event:', err);
        }
      });

      // Handle 'state_change' events
      es.addEventListener('state_change', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          handleEvent('state_change', data);
        } catch (err) {
          console.error('[SSE] Failed to parse state_change event:', err);
        }
      });

      // Handle generic messages
      es.onmessage = (event: MessageEvent) => {
        console.log('[SSE] Generic message:', event.data);
      };

      es.onerror = (event) => {
        console.error('[SSE] Error:', event);
        setIsConnected(false);
        setIsConnecting(false);
        setError(new Error('SSE connection error'));

        // EventSource automatically reconnects, but we track manual reconnect for control
        if (eventSource.current?.readyState === EventSource.CLOSED) {
          scheduleReconnect();
        }
      };

      eventSource.current = es;
    } catch (err) {
      console.error('[SSE] Connection failed:', err);
      setError(err instanceof Error ? err : new Error('Unknown error'));
      setIsConnecting(false);
      scheduleReconnect();
    }
  }, [url, isConnecting, handleEvent, scheduleReconnect]);

  // Disconnect from SSE stream
  const disconnect = useCallback(() => {
    console.log('[SSE] Disconnecting...');
    reconnectCount.current = reconnectAttempts; // Prevent reconnection

    if (reconnectTimeout.current) {
      clearTimeout(reconnectTimeout.current);
      reconnectTimeout.current = null;
    }

    if (eventSource.current) {
      eventSource.current.close();
      eventSource.current = null;
    }

    setIsConnected(false);
    setIsConnecting(false);
  }, [reconnectAttempts]);

  // Clear event history
  const clearEvents = useCallback(() => {
    setEvents([]);
    setLastEvent(null);
  }, []);

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, []);

  return {
    isConnected,
    isConnecting,
    error,
    lastEvent,
    events,
    connect,
    disconnect,
    clearEvents,
  };
}

/**
 * Hook for filtering SSE events by type
 */
export function useSSEEvents(eventTypes: SSEEventType[]): SSEEvent[] {
  const { events } = useSSE();

  return events.filter(event => eventTypes.includes(event.type));
}

/**
 * Hook for handling specific SSE event types with callback
 */
export function useSSEEventHandler(
  eventType: SSEEventType,
  callback: (data: any) => void,
  dependencies: any[] = []
) {
  const { lastEvent } = useSSE();

  useEffect(() => {
    if (lastEvent && lastEvent.type === eventType) {
      callback(lastEvent.data);
    }
  }, [lastEvent, eventType, ...dependencies]);
}

/**
 * Hook for real-time drift alerts
 */
export function useDriftAlerts() {
  const [driftAlerts, setDriftAlerts] = useState<any[]>([]);

  useSSEEventHandler('drift', (data) => {
    setDriftAlerts(prev => [data, ...prev].slice(0, 50)); // Keep last 50 alerts
  }, []);

  return {
    driftAlerts,
    clearAlerts: () => setDriftAlerts([]),
  };
}

/**
 * Hook for real-time Falco events
 */
export function useFalcoEvents() {
  const [falcoEvents, setFalcoEvents] = useState<any[]>([]);

  useSSEEventHandler('falco', (data) => {
    setFalcoEvents(prev => [data, ...prev].slice(0, 50)); // Keep last 50 events
  }, []);

  return {
    falcoEvents,
    clearEvents: () => setFalcoEvents([]),
  };
}
