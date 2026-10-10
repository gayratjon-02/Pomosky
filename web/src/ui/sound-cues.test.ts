import { describe, expect, it } from "vitest";
import type { TimerView } from "../core/protocol";
import { cueFor } from "./sound-cues";

const base: TimerView = { phase: "idle", startedAtMs: 0, durationMs: 0, pausedRemainingMs: 0, label: "" };
const focus = (startedAtMs: number): TimerView => ({ ...base, phase: "focus", startedAtMs, durationMs: 25 * 60_000 });
const paused: TimerView = { ...base, phase: "paused", pausedFrom: "focus", durationMs: 25 * 60_000, pausedRemainingMs: 600_000 };

describe("cueFor", () => {
  it("is silent for the first state and for no change", () => {
    expect(cueFor(null, focus(1000))).toBeNull();
    expect(cueFor(focus(1000), focus(1000))).toBeNull();
    expect(cueFor(base, base)).toBeNull();
  });

  it("plays start when a phase begins from idle or after the previous one ended", () => {
    expect(cueFor(base, focus(1000))).toBe("start");
    expect(cueFor(focus(1000), { ...focus(2000), phase: "break" })).toBe("start"); // a new phase after the end
  });

  it("plays pause and resume", () => {
    expect(cueFor(focus(1000), paused)).toBe("pause");
    expect(cueFor(paused, focus(5000))).toBe("resume");
  });

  it("plays reset when the timer returns to idle", () => {
    expect(cueFor(focus(1000), base)).toBe("reset");
    expect(cueFor(paused, base)).toBe("reset");
  });
});
