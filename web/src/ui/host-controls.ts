import type { Phase, TimerView } from "../core/protocol";
import { remainingMs } from "../core/timer";

export type ControlState = "idle" | "running" | "paused" | "ended";

export interface Controls {
  state: ControlState;
  /** The start buttons are shown only when no phase is running or paused. */
  showStart: boolean;
  /** Pause while running, Resume while paused. */
  showPause: boolean;
  pauseLabel: "Pause" | "Resume";
  showReset: boolean;
  /** The settings (minutes, label, presets) can be edited only while nothing is running or paused. */
  editable: boolean;
  /** After a phase ends the natural next one is highlighted. */
  suggest: "focus" | "break" | null;
}

/** Which controls the host sees for a timer at the given server time. Mirrors the server rule in room.canStart. */
export function controlsFor(t: TimerView, serverNowMs: number): Controls {
  const running = t.phase === "focus" || t.phase === "break";
  const ended = running && remainingMs(t, serverNowMs) === 0;

  let state: ControlState = "idle";
  if (t.phase === "paused") state = "paused";
  else if (ended) state = "ended";
  else if (running) state = "running";

  const lastPhase: Phase = t.phase;
  return {
    state,
    showStart: state === "idle" || state === "ended",
    showPause: state === "running" || state === "paused",
    pauseLabel: state === "paused" ? "Resume" : "Pause",
    showReset: state !== "idle",
    editable: state === "idle" || state === "ended",
    suggest: state === "ended" ? (lastPhase === "focus" ? "break" : "focus") : null,
  };
}
