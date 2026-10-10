import { maybeById } from "../core/dom";
import type { Phase, TimerView } from "../core/protocol";
import { formatMMSS, remainingMs } from "../core/timer";
import { celebrate } from "./celebrate";
import { chime } from "./sound";

const PHASE_TEXT: Record<Phase, string> = {
  idle: "Waiting for the host",
  focus: "Focus",
  break: "Break",
  paused: "Paused",
};

const RING_LENGTH = 2 * Math.PI * 120; // matches r=120 in the timer partial

/** Renders the timer (digits, ring, phase, label) and rings the chime when a phase ends. */
export class TimerRenderer {
  private readonly digits = maybeById("timer-digits");
  private readonly phase = maybeById("timer-phase");
  private readonly label = maybeById("timer-label");
  private readonly ring = maybeById("ring-progress");
  private readonly knob = maybeById("ring-knob");
  private chimedFor = -1;

  render(t: TimerView, serverNowMs: number): void {
    const remaining = remainingMs(t, serverNowMs);
    const text = t.phase === "idle" ? formatMMSS(0) : formatMMSS(remaining);
    const ended = (t.phase === "focus" || t.phase === "break") && remaining === 0;

    if (this.digits) this.digits.textContent = text;
    if (this.phase) this.phase.textContent = ended ? "Time's up" : PHASE_TEXT[t.phase];
    if (this.label) this.label.textContent = t.label;
    document.body.dataset["phase"] = ended ? "ended" : t.phase;
    document.body.dataset["running"] = t.phase === "focus" || t.phase === "break" ? "1" : "0";

    if (this.ring) {
      const total = t.durationMs > 0 ? t.durationMs : 1;
      const progress = t.phase === "idle" ? 0 : 1 - Math.min(1, remaining / total);
      this.ring.style.strokeDasharray = String(RING_LENGTH);
      this.ring.style.strokeDashoffset = String(RING_LENGTH * (1 - progress));
      if (this.knob) this.knob.style.transform = `rotate(${progress * 360}deg)`; // the knob rides the end of the arc
    }

    document.title = t.phase === "idle" ? "Pomosky" : `${text} · ${ended ? "Done" : PHASE_TEXT[t.phase]} · Pomosky`;

    if (ended && this.chimedFor !== t.startedAtMs) {
      this.chimedFor = t.startedAtMs;
      chime();
      celebrate();
    }
  }
}
