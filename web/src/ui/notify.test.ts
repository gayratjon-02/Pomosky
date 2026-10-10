import { describe, expect, it } from "vitest";
import type { TimerView } from "../core/protocol";
import { endMessage } from "./notify";

const t = (phase: TimerView["phase"], label = ""): TimerView => ({ phase, startedAtMs: 0, durationMs: 0, pausedRemainingMs: 0, label });

describe("endMessage", () => {
  it("says a focus session finished and uses the label", () => {
    expect(endMessage(t("focus", "Go backend"))).toEqual({ title: "Focus session finished", body: "Go backend. Time for a break." });
    expect(endMessage(t("focus"))).toEqual({ title: "Focus session finished", body: "Time for a break." });
  });
  it("says the break is over", () => {
    expect(endMessage(t("break")).title).toBe("Break is over");
  });
});
