// Mirror of internal/realtime/protocol.go. Keep both in sync (see docs/ws-protocol.md).

export type Phase = "idle" | "focus" | "break" | "paused";

export interface Envelope<T = unknown> {
  type: string;
  data?: T;
}

export interface TimerView {
  phase: Phase;
  startedAtMs: number;
  durationMs: number;
  pausedRemainingMs: number;
  label: string;
}

export interface StateData {
  roomId: string;
  timer: TimerView;
  viewers: number;
  reactions: Record<string, number>;
  city: string;
  serverNowMs: number;
}

export interface PresenceData {
  viewers: number;
}

export interface ReactionTickData {
  kind: string;
  count: number;
}

export interface ErrorData {
  code: string;
  message: string;
}

export const ClientMsg = {
  Join: "join",
  HostAuth: "host.auth",
  TimerStart: "timer.start",
  TimerPause: "timer.pause",
  TimerResume: "timer.resume",
  TimerReset: "timer.reset",
  Reaction: "reaction",
  Ping: "ping",
} as const;

export const ServerMsg = {
  State: "state",
  TimerUpdated: "timer.updated",
  Presence: "presence",
  ReactionTick: "reaction.tick",
  Weather: "weather",
  Pong: "pong",
  Error: "error",
} as const;
