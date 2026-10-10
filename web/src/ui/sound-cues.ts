import type { TimerView } from "../core/protocol";

export type Cue = "start" | "pause" | "resume" | "reset";

/**
 * Which sound a change of the room timer deserves, if any. It is driven by the
 * server's state, so the host and every viewer hear the same cue. The first state
 * after joining (prev is null) never makes a sound.
 */
export function cueFor(prev: TimerView | null, next: TimerView): Cue | null {
  if (!prev) return null;
  const nextRunning = next.phase === "focus" || next.phase === "break";
  const prevRunning = prev.phase === "focus" || prev.phase === "break";

  if (next.phase === "idle" && prev.phase !== "idle") return "reset";
  if (next.phase === "paused" && prevRunning) return "pause";
  if (nextRunning && prev.phase === "paused") return "resume";
  if (nextRunning && (prev.phase === "idle" || (prevRunning && prev.startedAtMs !== next.startedAtMs))) return "start";
  return null;
}
