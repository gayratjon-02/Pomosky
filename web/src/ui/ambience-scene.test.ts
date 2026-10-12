import { describe, expect, it } from "vitest";
import { MOODS } from "../sky/moods";
import { DEFAULT_INPUTS } from "../sky/types";
import { sceneFor } from "./ambience-scene";

const of = (id: string) => sceneFor({ ...DEFAULT_INPUTS, ...MOODS.find((m) => m.id === id)!.sky! });

describe("sceneFor", () => {
  it("has birds at dawn and crickets after sunset", () => {
    expect(of("dawn").birds).toBeGreaterThan(0.5);
    expect(of("dawn").birds).toBeGreaterThan(of("sunset").birds);
    expect(of("sunset").crickets).toBeGreaterThan(of("dawn").crickets);
    expect(of("night").crickets).toBeGreaterThan(0.4);
    expect(of("night").birds).toBe(0);
  });

  it("is quiet and airy on a clear day", () => {
    const s = of("clear");
    expect(s.rain).toBe(0);
    expect(s.birds).toBeGreaterThan(0.2);
    expect(s.crickets).toBe(0);
  });

  it("rains in the rain, snows in the snow, thunders in the storm", () => {
    expect(of("rain").rain).toBeGreaterThan(0.5);
    expect(of("rain").birds).toBe(0);
    expect(of("snow").hush).toBeGreaterThan(0.3);
    expect(of("snow").rain).toBe(0);
    expect(of("storm").thunder).toBe(1);
    expect(of("storm").rain).toBeGreaterThan(of("rain").rain);
    expect(of("storm").wind).toBeGreaterThan(of("rain").wind);
  });

  it("gives the moon night an owl and a deep drone", () => {
    expect(of("moon").owl).toBeGreaterThan(of("blue").owl);
    expect(of("moon").drone).toBeGreaterThan(0.4);
  });

  it("keeps every layer within 0..1", () => {
    for (const m of MOODS) {
      const s = sceneFor({ ...DEFAULT_INPUTS, ...(m.sky ?? {}) });
      for (const [k, v] of Object.entries(s)) {
        expect(v, `${m.id}.${k}`).toBeGreaterThanOrEqual(0);
        expect(v, `${m.id}.${k}`).toBeLessThanOrEqual(1);
      }
    }
  });
});
