/**
 * Centralized logger utility
 * In development: logs to console
 * In production: can integrate with external logging service
 */

interface LoggerInterface {
  debug(message: string, data?: unknown): void;
  info(message: string, data?: unknown): void;
  warn(message: string, data?: unknown): void;
  error(message: string, data?: unknown): void;
}

class Logger implements LoggerInterface {
  private getIsDevelopment(): boolean {
    return process.env.NODE_ENV === 'development';
  }

  debug(message: string, data?: unknown): void {
    if (this.getIsDevelopment()) {
      console.debug(`[DEBUG] ${message}`, data);
    }
  }

  info(message: string, data?: unknown): void {
    if (this.getIsDevelopment()) {
      console.info(`[INFO] ${message}`, data);
    }
  }

  warn(message: string, data?: unknown): void {
    if (this.getIsDevelopment()) {
      console.warn(`[WARN] ${message}`, data);
    }
    // In production, you could send warnings to external logging service
    // await this.sendToLoggingService('warn', message, data);
  }

  error(message: string, data?: unknown): void {
    if (this.getIsDevelopment()) {
      console.error(`[ERROR] ${message}`, data);
    }
    // In production, you could send errors to external logging service
    // await this.sendToLoggingService('error', message, data);
  }

  // Placeholder for sending logs to external service
  // private async sendToLoggingService(
  //   level: 'warn' | 'error',
  //   message: string,
  //   data?: unknown
  // ): Promise<void> {
  //   try {
  //     await fetch('/api/logs', {
  //       method: 'POST',
  //       headers: { 'Content-Type': 'application/json' },
  //       body: JSON.stringify({ level, message, data, timestamp: new Date() }),
  //     });
  //   } catch (err) {
  //     // Fail silently to avoid recursive logging issues
  //   }
  // }
}

export const logger = new Logger();
