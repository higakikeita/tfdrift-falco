/**
 * WebSocket Client
 * Bidirectional real-time communication with TFDrift API
 */

import { useEffect, useRef, useState, useCallback } from 'react';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

export type WSMessageType = 'subscribe' | 'unsubscribe' | 'ping' | 'query';
export type WSTopic = 'all' | 'drifts' | 'events' | 'state' | 'stats';

export interface WSMessage {
  type: WSMessageType;
  topic?: WSTopic;
  payload?: any;
}

export interface WSResponse {
  type: string;
  topic?: string;
  payload: any;
  timestamp?: string;
}

export interface UseWebSocketOptions {
  url?: string;
  autoConnect?: boolean;
  reconnectAttempts?: number;
  reconnectDelay?: number;
  heartbeatInterval?: number;
}

export interface UseWebSocketReturn {
  isConnected: boolean;
  isConnecting: boolean;
  error: Error | null;
  lastMessage: WSResponse | null;
  send: (message: WSMessage) => void;
  subscribe: (topic: WSTopic) => void;
  unsubscribe: (topic: WSTopic) => void;
  connect: () => void;
  disconnect: () => void;
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    url = WS_URL,
    autoConnect = true,
    reconnectAttempts = 5,
    reconnectDelay = 3000,
    heartbeatInterval = 30000,
  } = options;

  const ws = useRef<WebSocket | null>(null);
  const reconnectCount = useRef(0);
  const reconnectTimeout = useRef<NodeJS.Timeout | null>(null);
  const heartbeatTimeout = useRef<NodeJS.Timeout | null>(null);

  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastMessage, setLastMessage] = useState<WSResponse | null>(null);

  // Send message
  const send = useCallback((message: WSMessage) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.warn('[WebSocket] Cannot send message: connection not open');
    }
  }, []);

  // Subscribe to topic
  const subscribe = useCallback((topic: WSTopic) => {
    send({ type: 'subscribe', topic });
  }, [send]);

  // Unsubscribe from topic
  const unsubscribe = useCallback((topic: WSTopic) => {
    send({ type: 'unsubscribe', topic });
  }, [send]);

  // Start heartbeat
  const startHeartbeat = useCallback(() => {
    if (heartbeatTimeout.current) {
      clearInterval(heartbeatTimeout.current);
    }

    heartbeatTimeout.current = setInterval(() => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        send({ type: 'ping' });
      }
    }, heartbeatInterval);
  }, [send, heartbeatInterval]);

  // Stop heartbeat
  const stopHeartbeat = useCallback(() => {
    if (heartbeatTimeout.current) {
      clearInterval(heartbeatTimeout.current);
      heartbeatTimeout.current = null;
    }
  }, []);

  // Schedule reconnect
  const scheduleReconnect = useCallback(() => {
    if (reconnectCount.current >= reconnectAttempts) {
      setError(new Error(`Failed to connect after ${reconnectAttempts} attempts`));
      setIsConnecting(false);
      return;
    }

    reconnectCount.current++;
    const delay = reconnectDelay * Math.pow(2, reconnectCount.current - 1); // Exponential backoff

    console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${reconnectCount.current}/${reconnectAttempts})`);

    reconnectTimeout.current = setTimeout(() => {
      connect();
    }, delay);
  }, [reconnectAttempts, reconnectDelay]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    // Prevent multiple connections
    if (ws.current?.readyState === WebSocket.OPEN || isConnecting) {
      return;
    }

    setIsConnecting(true);
    setError(null);

    try {
      console.log('[WebSocket] Connecting to', url);
      const socket = new WebSocket(url);

      socket.onopen = () => {
        console.log('[WebSocket] Connected');
        setIsConnected(true);
        setIsConnecting(false);
        setError(null);
        reconnectCount.current = 0;
        startHeartbeat();

        // Auto-subscribe to all events
        socket.send(JSON.stringify({ type: 'subscribe', topic: 'all' }));
      };

      socket.onmessage = (event) => {
        try {
          const message: WSResponse = JSON.parse(event.data);
          console.log('[WebSocket] Message received:', message);
          setLastMessage(message);
        } catch (err) {
          console.error('[WebSocket] Failed to parse message:', err);
        }
      };

      socket.onerror = (event) => {
        console.error('[WebSocket] Error:', event);
        setError(new Error('WebSocket connection error'));
      };

      socket.onclose = (event) => {
        console.log('[WebSocket] Disconnected:', event.code, event.reason);
        setIsConnected(false);
        setIsConnecting(false);
        stopHeartbeat();

        // Attempt reconnect if not intentionally closed
        if (event.code !== 1000 && event.code !== 1001) {
          scheduleReconnect();
        }
      };

      ws.current = socket;
    } catch (err) {
      console.error('[WebSocket] Connection failed:', err);
      setError(err instanceof Error ? err : new Error('Unknown error'));
      setIsConnecting(false);
      scheduleReconnect();
    }
  }, [url, isConnecting, startHeartbeat, stopHeartbeat, scheduleReconnect]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    console.log('[WebSocket] Disconnecting...');
    reconnectCount.current = reconnectAttempts; // Prevent reconnection

    if (reconnectTimeout.current) {
      clearTimeout(reconnectTimeout.current);
      reconnectTimeout.current = null;
    }

    stopHeartbeat();

    if (ws.current) {
      ws.current.close(1000, 'Client disconnect');
      ws.current = null;
    }

    setIsConnected(false);
    setIsConnecting(false);
  }, [reconnectAttempts, stopHeartbeat]);

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
    lastMessage,
    send,
    subscribe,
    unsubscribe,
    connect,
    disconnect,
  };
}

/**
 * Hook for subscribing to specific WebSocket events
 */
export function useWebSocketEvent(
  eventType: string,
  callback: (payload: any) => void,
  dependencies: any[] = []
) {
  const { lastMessage } = useWebSocket();

  useEffect(() => {
    if (lastMessage && lastMessage.type === eventType) {
      callback(lastMessage.payload);
    }
  }, [lastMessage, eventType, ...dependencies]);
}
