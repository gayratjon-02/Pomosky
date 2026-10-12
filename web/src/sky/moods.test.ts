import { describe, expect, it } from "vitest";
import { MOODS } from "./moods";

describe("MOODS", () => {
  it("has unique ids and starts with the live sky", () => {
    const ids = MOODS.map((m) => m.id);
    expect(new Set(ids).size).toBe(ids.length);
    expect(MOODS[0]?.id).toBe("live");
    expect(MOODS[0]?.sky).toBeNull();
  });

  it("fixes the sun for every other mood", () => {
    for (const m of MOODS.slice(1)) {
      expect(m.sky, m.id).not.toBeNull();
      expect(typeof m.sky?.sunElevationDeg, m.id).toBe("number");
    }
  });
});
