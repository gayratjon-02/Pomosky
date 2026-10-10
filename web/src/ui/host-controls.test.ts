import { describe, expect, it } from "vitest";
import type { TimerView } from "../core/protocol";
import { controlsFor } from "./host-controls";

const T0 = 1_700_000_000_000;
const base: TimerView = { phase: "idle", startedAtMs: 0, durationMs: 0, pausedRemainingMs: 0, label: "" };
const running = (phase: "focus" | "break"): TimerView => ({ ...base, phase, startedAtMs: T0, durationMs: 25 * 60_000 });

describe("controlsFor", () => {
  it("idle: only the start buttons, settings editable", () => {
    const c = controlsFor(base, T0);
    expect(c).toMatchObject({ state: "idle", showStart: true, showPause: false, showReset: false, editable: true, suggest: null });
  });

  it("running: Pause and Reset, no start buttons, settings locked", () => {
    const c = controlsFor(running("focus"), T0 + 60_000);
    expect(c).toMatchObject({ state: "running", showStart: false, showPause: true, pauseLabel: "Pause", showReset: true, editable: false });
  });

  it("paused: Resume and Reset, no start buttons", () => {
    const c = controlsFor({ ...base, phase: "paused", pausedFrom: "focus", durationMs: 25 * 60_000, pausedRemainingMs: 10 * 60_000 }, T0);
    expect(c).toMatchObject({ state: "paused", showStart: false, showPause: true, pauseLabel: "Resume", showReset: true, editable: false });
  });

  it("ended focus: start buttons are back and break is suggested", () => {
    const c = controlsFor(running("focus"), T0 + 25 * 60_000);
    expect(c).toMatchObject({ state: "ended", showStart: true, showPause: false, showReset: true, editable: true, suggest: "break" });
  });

  it("ended break suggests focus", () => {
    expect(controlsFor(running("break"), T0 + 26 * 60_000).suggest).toBe("focus");
  });

  it("one millisecond before the end it is still running", () => {
    expect(controlsFor(running("focus"), T0 + 25 * 60_000 - 1).state).toBe("running");
  });
});
