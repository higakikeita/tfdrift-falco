/**
 * Connection Status Component
 * Shows WebSocket and SSE connection status with visual indicators
 */

import { memo } from 'react';
import { Wifi, WifiOff, Activity, Radio } from 'lucide-react';
import { useWebSocket } from '../api/websocket';
import { useSSE } from '../api/sse';
import { Card } from './ui/card';

interface ConnectionStatusProps {
  showDetails?: boolean;
  compact?: boolean;
}

export const ConnectionStatus = memo(({ showDetails = false, compact = false }: ConnectionStatusProps) => {
  const {
    isConnected: wsConnected,
    isConnecting: wsConnecting,
    error: wsError
  } = useWebSocket({ autoConnect: false }); // Don't auto-connect in status component

  const {
    isConnected: sseConnected,
    isConnecting: sseConnecting,
    error: sseError
  } = useSSE({ autoConnect: false }); // Don't auto-connect in status component

  // Compact view
  if (compact) {
    return (
      <div className="flex items-center gap-2">
        {/* WebSocket Status */}
        <div
          className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
            wsConnected
              ? 'bg-green-100 text-green-700'
              : wsConnecting
              ? 'bg-yellow-100 text-yellow-700'
              : 'bg-red-100 text-red-700'
          }`}
          title={`WebSocket: ${wsConnected ? 'Connected' : wsConnecting ? 'Connecting...' : 'Disconnected'}`}
        >
          {wsConnected ? (
            <Wifi className="w-3 h-3" />
          ) : (
            <WifiOff className="w-3 h-3" />
          )}
          <span>WS</span>
        </div>

        {/* SSE Status */}
        <div
          className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
            sseConnected
              ? 'bg-blue-100 text-blue-700'
              : sseConnecting
              ? 'bg-yellow-100 text-yellow-700'
              : 'bg-red-100 text-red-700'
          }`}
          title={`SSE: ${sseConnected ? 'Connected' : sseConnecting ? 'Connecting...' : 'Disconnected'}`}
        >
          {sseConnected ? (
            <Radio className="w-3 h-3" />
          ) : (
            <Activity className="w-3 h-3" />
          )}
          <span>SSE</span>
        </div>
      </div>
    );
  }

  // Detailed view
  return (
    <Card className="p-3">
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-gray-700">Real-time Status</span>
          <div className="flex items-center gap-1">
            {(wsConnected || sseConnected) && (
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
              </span>
            )}
          </div>
        </div>

        {/* WebSocket Status */}
        <div className="flex items-start gap-2">
          <div className={`
            mt-0.5 p-1 rounded
            ${wsConnected ? 'bg-green-100' : wsConnecting ? 'bg-yellow-100' : 'bg-red-100'}
          `}>
            {wsConnected ? (
              <Wifi className="w-4 h-4 text-green-600" />
            ) : (
              <WifiOff className="w-4 h-4 text-red-600" />
            )}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">WebSocket</span>
              <span className={`
                text-xs font-medium px-2 py-0.5 rounded-full
                ${wsConnected ? 'bg-green-100 text-green-700' : wsConnecting ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}
              `}>
                {wsConnected ? 'Connected' : wsConnecting ? 'Connecting' : 'Disconnected'}
              </span>
            </div>
            {showDetails && (
              <p className="text-xs text-gray-500 mt-0.5">
                {wsConnected
                  ? 'Bidirectional communication active'
                  : wsConnecting
                  ? 'Establishing connection...'
                  : wsError
                  ? wsError.message
                  : 'Not connected'}
              </p>
            )}
          </div>
        </div>

        {/* SSE Status */}
        <div className="flex items-start gap-2">
          <div className={`
            mt-0.5 p-1 rounded
            ${sseConnected ? 'bg-blue-100' : sseConnecting ? 'bg-yellow-100' : 'bg-red-100'}
          `}>
            {sseConnected ? (
              <Radio className="w-4 h-4 text-blue-600" />
            ) : (
              <Activity className="w-4 h-4 text-red-600" />
            )}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">Server-Sent Events</span>
              <span className={`
                text-xs font-medium px-2 py-0.5 rounded-full
                ${sseConnected ? 'bg-blue-100 text-blue-700' : sseConnecting ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}
              `}>
                {sseConnected ? 'Streaming' : sseConnecting ? 'Connecting' : 'Offline'}
              </span>
            </div>
            {showDetails && (
              <p className="text-xs text-gray-500 mt-0.5">
                {sseConnected
                  ? 'Receiving live events'
                  : sseConnecting
                  ? 'Opening event stream...'
                  : sseError
                  ? sseError.message
                  : 'Not connected'}
              </p>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
});

ConnectionStatus.displayName = 'ConnectionStatus';

/**
 * Minimal connection indicator for header/toolbar
 */
export const ConnectionIndicator = memo(() => {
  const { isConnected: wsConnected } = useWebSocket({ autoConnect: false });
  const { isConnected: sseConnected } = useSSE({ autoConnect: false });

  const isConnected = wsConnected || sseConnected;

  return (
    <div
      className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
        isConnected
          ? 'bg-green-50 text-green-700 border border-green-200'
          : 'bg-gray-50 text-gray-500 border border-gray-200'
      }`}
      title={`Real-time: ${isConnected ? 'Active' : 'Offline'}`}
    >
      <span className="relative flex h-2 w-2">
        {isConnected && (
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
        )}
        <span className={`relative inline-flex rounded-full h-2 w-2 ${
          isConnected ? 'bg-green-500' : 'bg-gray-400'
        }`}></span>
      </span>
      <span>{isConnected ? 'Live' : 'Offline'}</span>
    </div>
  );
});

ConnectionIndicator.displayName = 'ConnectionIndicator';
