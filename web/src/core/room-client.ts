import { ClockSync, estimateOffset } from "./clock-sync";
import { Emitter } from "./events";
import {
  ClientMsg,
  ServerMsg,
  type Envelope,
  type ErrorData,
  type PongData,
  type ReactionTickData,
  type StateData,
  type TimerStartData,
} from "./protocol";
import { ReconnectingSocket, type SocketStatus } from "./ws";

export interface RoomEvents extends Record<string, unknown> {
  status: SocketStatus;
  state: StateData;
  tick: ReactionTickData;
  hostOk: undefined;
  error: ErrorData;
}

const PING_INTERVAL_MS = 25_000;
const BURST_PING_DELAYS_MS = [0, 400, 1200];

/**
 * One room connection: reconnects, keeps the server clock offset up to date and
 * turns the wire protocol into typed events. Pages never touch raw frames.
 */
export class RoomClient extends Emitter<RoomEvents> {
  readonly clock = new ClockSync();
  state: StateData | null = null;
  isHost = false;

  private readonly socket: ReconnectingSocket;
  private pendingPings: number[] = [];
  private pingTimer: number | undefined;
  private burstTimers: number[] = [];

  constructor(
    roomId: string,
    private readonly opts: { hostToken?: string; nickname?: string } = {},
  ) {
    super();
    const proto = location.protocol === "https:" ? "wss" : "ws";
    this.socket = new ReconnectingSocket(`${proto}://${location.host}/ws/${roomId}`, 0);
    this.socket.onStatus = (status) => this.handleStatus(status);
    this.socket.onMessage = (env) => this.handleMessage(env);
  }

  connect(): void {
    this.socket.connect();
  }

  close(): void {
    this.stopPings();
    this.socket.close();
  }

  setNickname(nickname: string): void {
    this.socket.send(ClientMsg.Join, { nickname });
  }

  react(kind: string): void {
    this.socket.send(ClientMsg.Reaction, { kind });
  }

  start(data: TimerStartData): void {
    this.socket.send(ClientMsg.TimerStart, data);
  }

  pause(): void {
    this.socket.send(ClientMsg.TimerPause);
  }

  resume(): void {
    this.socket.send(ClientMsg.TimerResume);
  }

  reset(): void {
    this.socket.send(ClientMsg.TimerReset);
  }

  private handleStatus(status: SocketStatus): void {
    if (status === "open") {
      if (this.opts.hostToken) this.socket.send(ClientMsg.HostAuth, { token: this.opts.hostToken });
      if (this.opts.nickname) this.setNickname(this.opts.nickname);
      this.startPings();
    } else {
      this.isHost = false;
      this.stopPings();
    }
    this.emit("status", status);
  }

  private handleMessage(env: Envelope): void {
    switch (env.type) {
      case ServerMsg.State: {
        const data = env.data as StateData;
        // One-way sample: only good enough until the first ping round trip lands.
        if (this.clock.sampleCount === 0) this.clock.add(data.serverNowMs - Date.now());
        this.state = data;
        this.emit("state", data);
        break;
      }
      case ServerMsg.Pong: {
        const sentAt = this.pendingPings.shift();
        if (sentAt !== undefined) {
          const { serverNowMs } = env.data as PongData;
          this.clock.add(estimateOffset(serverNowMs, sentAt, Date.now()));
        }
        break;
      }
      case ServerMsg.ReactionTick:
        this.emit("tick", env.data as ReactionTickData);
        break;
      case ServerMsg.HostOk:
        this.isHost = true;
        this.emit("hostOk", undefined);
        break;
      case ServerMsg.Error:
        this.emit("error", env.data as ErrorData);
        break;
    }
  }

  private ping(): void {
    this.pendingPings.push(Date.now());
    if (this.pendingPings.length > 8) this.pendingPings.shift();
    this.socket.send(ClientMsg.Ping);
  }

  private startPings(): void {
    this.stopPings();
    this.pendingPings = [];
    this.burstTimers = BURST_PING_DELAYS_MS.map((d) => window.setTimeout(() => this.ping(), d));
    this.pingTimer = window.setInterval(() => this.ping(), PING_INTERVAL_MS);
  }

  private stopPings(): void {
    window.clearInterval(this.pingTimer);
    this.burstTimers.forEach((t) => window.clearTimeout(t));
    this.burstTimers = [];
  }
}
