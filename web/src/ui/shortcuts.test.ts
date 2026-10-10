import { describe, expect, it } from "vitest";
import type { TimerView } from "../core/protocol";
import { controlsFor } from "./host-controls";
import { shortcutFor, type KeyInfo } from "./shortcuts";

const T0 = 1_700_000_000_000;
const idle: TimerView = { phase: "idle", startedAtMs: 0, durationMs: 0, pausedRemainingMs: 0, label: "" };
const running: TimerView = { ...idle, phase: "focus", startedAtMs: T0, durationMs: 25 * 60_000 };

const key = (k: string, extra: Partial<KeyInfo> = {}): KeyInfo => ({ key: k, shiftKey: false, ctrlKey: false, metaKey: false, altKey: false, typing: false, ...extra });

describe("shortcutFor", () => {
  it("maps the keys to actions that fit the state", () => {
    const idleC = controlsFor(idle, T0);
    expect(shortcutFor(key("f"), idleC)).toBe("start-focus");
    expect(shortcutFor(key("b"), idleC)).toBe("start-break");
    expect(shortcutFor(key(" "), idleC)).toBeNull(); // nothing to pause yet
    expect(shortcutFor(key("R", { shiftKey: true }), idleC)).toBeNull(); // nothing to reset yet

    const runC = controlsFor(running, T0 + 60_000);
    expect(shortcutFor(key(" "), runC)).toBe("toggle-pause");
    expect(shortcutFor(key("k"), runC)).toBe("toggle-pause");
    expect(shortcutFor(key("R", { shiftKey: true }), runC)).toBe("reset");
    expect(shortcutFor(key("f"), runC)).toBeNull(); // start is hidden while running
  });

  it("needs Shift for reset, so a plain r does nothing", () => {
    expect(shortcutFor(key("r"), controlsFor(running, T0 + 1000))).toBeNull();
  });

  it("ignores typing and modifier combinations", () => {
    const c = controlsFor(idle, T0);
    expect(shortcutFor(key("f", { typing: true }), c)).toBeNull();
    expect(shortcutFor(key("f", { metaKey: true }), c)).toBeNull();
    expect(shortcutFor(key("f", { ctrlKey: true }), c)).toBeNull();
    expect(shortcutFor(key("b", { altKey: true }), c)).toBeNull();
  });
});
