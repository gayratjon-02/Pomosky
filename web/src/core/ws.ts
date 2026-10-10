import type { Envelope } from "./protocol";

export type SocketStatus = "connecting" | "open" | "reconnecting" | "closed";

/**
 * WebSocket with automatic reconnect (exponential backoff + jitter) and a
 * heartbeat. The heartbeat keeps idle connections alive behind cloud networks
 * that drop silent TCP connections.
 */
export class ReconnectingSocket {
  onMessage: (env: Envelope) => void = () => {};
  onStatus: (status: SocketStatus) => void = () => {};

  private ws: WebSocket | null = null;
  private attempt = 0;
  private reconnectTimer: number | undefined;
  private heartbeatTimer: number | undefined;
  private closedByUser = false;

  constructor(
    private readonly url: string,
    private readonly heartbeatMs = 25_000,
  ) {}

  connect(): void {
    this.closedByUser = false;
    this.onStatus(this.attempt === 0 ? "connecting" : "reconnecting");

    const ws = new WebSocket(this.url);
    this.ws = ws;

    ws.onopen = () => {
      this.attempt = 0;
      this.onStatus("open");
      this.startHeartbeat();
    };
    ws.onmessage = (ev: MessageEvent) => {
      try {
        this.onMessage(JSON.parse(String(ev.data)) as Envelope);
      } catch {
        // ignore malformed frames
      }
    };
    ws.onclose = () => {
      this.stopHeartbeat();
      if (this.closedByUser) {
        this.onStatus("closed");
        return;
      }
      this.scheduleReconnect();
    };
    ws.onerror = () => ws.close();
  }

  send(type: string, data?: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }

  close(): void {
    this.closedByUser = true;
    window.clearTimeout(this.reconnectTimer);
    this.stopHeartbeat();
    this.ws?.close();
  }

  private scheduleReconnect(): void {
    this.onStatus("reconnecting");
    const base = Math.min(1000 * 2 ** this.attempt, 15_000);
    const delay = base / 2 + Math.random() * (base / 2);
    this.attempt++;
    this.reconnectTimer = window.setTimeout(() => this.connect(), delay);
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = window.setInterval(() => this.send("ping"), this.heartbeatMs);
  }

  private stopHeartbeat(): void {
    window.clearInterval(this.heartbeatTimer);
  }
}
