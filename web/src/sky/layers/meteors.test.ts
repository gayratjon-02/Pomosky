import { describe, expect, it } from "vitest";
import { nextGapSeconds } from "./meteors";

describe("nextGapSeconds", () => {
  it("stays between 7 and 18 seconds", () => {
    expect(nextGapSeconds(() => 0)).toBe(7);
    expect(nextGapSeconds(() => 1)).toBe(18);
    expect(nextGapSeconds(() => 0.5)).toBeCloseTo(12.5);
  });
});
