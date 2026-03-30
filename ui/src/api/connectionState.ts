/**
 * Connection state types and observable
 */

import { logger } from '../utils/logger';

export type ConnectionState =
  | 'connected'
  | 'connecting'
  | 'disconnected'
  | 'reconnecting';

interface ConnectionStateListener {
  (state: ConnectionState): void;
}

class ConnectionStateObservable {
  private state: ConnectionState = 'disconnected';
  private listeners: Set<ConnectionStateListener> = new Set();

  /**
   * Get current connection state
   */
  getState(): ConnectionState {
    return this.state;
  }

  /**
   * Update connection state
   */
  setState(newState: ConnectionState): void {
    if (this.state !== newState) {
      this.state = newState;
      this.notifyListeners();
    }
  }

  /**
   * Subscribe to connection state changes
   */
  subscribe(listener: ConnectionStateListener): () => void {
    this.listeners.add(listener);

    // Return unsubscribe function
    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Notify all listeners of state change
   */
  private notifyListeners(): void {
    this.listeners.forEach((listener) => {
      try {
        listener(this.state);
      } catch (err) {
        // Prevent one listener error from affecting others
        logger.error('Error in connection state listener:', err);
      }
    });
  }

  /**
   * Get number of active listeners (useful for testing)
   */
  getListenerCount(): number {
    return this.listeners.size;
  }
}

export const connectionStateObservable = new ConnectionStateObservable();
