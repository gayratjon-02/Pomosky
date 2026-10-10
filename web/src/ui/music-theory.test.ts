import { describe, expect, it } from "vitest";
import { CHORDS, PENTATONIC, midiToFreq, pluckNote } from "./music-theory";

describe("music theory", () => {
  it("tunes A4 to 440 Hz and octaves double", () => {
    expect(midiToFreq(69)).toBeCloseTo(440);
    expect(midiToFreq(81)).toBeCloseTo(880);
  });
  it("keeps the pad chords in a low range", () => {
    for (const chord of CHORDS) for (const n of chord) expect(n).toBeGreaterThanOrEqual(36), expect(n).toBeLessThan(72);
  });
  it("never repeats the previous pluck", () => {
    for (const prev of PENTATONIC) for (let r = 0; r < 1; r += 0.05) expect(pluckNote(r, prev)).not.toBe(prev);
  });
});
