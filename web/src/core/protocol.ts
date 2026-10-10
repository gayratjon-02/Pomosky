// Mirror of internal/realtime/protocol.go. Keep both in sync (see docs/ws-protocol.md).

export type Phase = "idle" | "focus" | "break" | "paused";

export interface Envelope<T = unknown> {
  type: string;
  data?: T;
}

export interface TimerView {
  phase: Phase;
  pausedFrom?: Phase;
  startedAtMs: number;
  durationMs: number;
  pausedRemainingMs: number;
  label: string;
}

export interface LocationView {
  lat: number;
  lon: number;
  tz: string;
}

export interface LocationSetData {
  city: string;
  lat: number;
  lon: number;
  tz: string;
}

export interface StateData {
  roomId: string;
  timer: TimerView;
  viewers: number;
  reactions: Record<string, number>;
  city: string;
  location?: LocationView;
  serverNowMs: number;
}

export interface ReactionTickData {
  kind: string;
  count: number;
}

export interface PongData {
  serverNowMs: number;
}

export interface ErrorData {
  code: string;
  message: string;
}

export interface TimerStartData {
  phase: "focus" | "break";
  durationMs: number;
  label: string;
}

export const ClientMsg = {
  Join: "join",
  HostAuth: "host.auth",
  TimerStart: "timer.start",
  TimerPause: "timer.pause",
  TimerResume: "timer.resume",
  TimerReset: "timer.reset",
  TimerRecover: "timer.recover",
  Reaction: "reaction",
  LocationSet: "location.set",
  Ping: "ping",
} as const;

export const ServerMsg = {
  State: "state",
  HostOk: "host.ok",
  ReactionTick: "reaction.tick",
  Weather: "weather", // reserved for stage 6
  Pong: "pong",
  Error: "error",
} as const;

export const ErrorCode = {
  BadRequest: "bad_request",
  Forbidden: "forbidden",
  Invalid: "invalid",
  InvalidTransition: "invalid_transition",
  RateLimited: "rate_limited",
} as const;
