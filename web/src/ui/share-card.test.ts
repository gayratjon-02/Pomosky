import { describe, expect, it } from "vitest";
import { monthContent, sessionContent } from "./share-card";

describe("card content", () => {
  it("shows a short session in minutes", () => {
    expect(sessionContent(25 * 60_000, "")).toEqual({ big: "25", line: "minutes of focus" });
    expect(sessionContent(60_000, "  ")).toEqual({ big: "1", line: "minute of focus" });
  });

  it("shows a long session in hours and minutes", () => {
    expect(sessionContent(85 * 60_000, "")).toEqual({ big: "1h 25m", line: "of focus" });
  });

  it("keeps what the person worked on", () => {
    expect(sessionContent(25 * 60_000, " essay ").label).toBe("essay");
  });

  it("sums up a month", () => {
    const c = monthContent(42 * 3600 + 600, 31, 12, "October 2026");
    expect(c.big).toBe("42h 10m");
    expect(c.line).toBe("focused in October 2026");
    expect(c.sub).toBe("31 sessions · 12 active days");
  });
});
