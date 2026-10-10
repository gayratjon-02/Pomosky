import type { TimerView } from "./protocol";

/** Remaining time at the given server time (ms). Mirrors room.Timer.Remaining in Go. */
export function remainingMs(t: TimerView, serverNowMs: number): number {
  switch (t.phase) {
    case "paused":
      return t.pausedRemainingMs;
    case "focus":
    case "break":
      return Math.max(0, t.startedAtMs + t.durationMs - serverNowMs);
    default:
      return 0;
  }
}

/** Formats milliseconds as MM:SS (rounded up to the next second). */
export function formatMMSS(ms: number): string {
  const total = Math.ceil(Math.max(0, ms) / 1000);
  const m = Math.floor(total / 60);
  const s = total % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}
