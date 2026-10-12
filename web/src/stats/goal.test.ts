import { describe, expect, it } from "vitest";
import { goalProgress } from "./goal";

describe("goalProgress", () => {
  it("is the share of the goal, capped at 1", () => {
    expect(goalProgress(0, 60)).toBe(0);
    expect(goalProgress(1800, 60)).toBe(0.5);
    expect(goalProgress(9000, 60)).toBe(1);
  });
  it("ignores bad input", () => {
    expect(goalProgress(-5, 60)).toBe(0);
    expect(goalProgress(100, 0)).toBe(0);
  });
});
