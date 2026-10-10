// Notes for the ambient music. Pure, so the tuning is easy to test.

const A4 = 440;

/** Frequency of a MIDI note number. */
export function midiToFreq(note: number): number {
  return A4 * Math.pow(2, (note - 69) / 12);
}

/** A slow, calm loop: Am9, Fmaj7, Cmaj9, G6 as MIDI note numbers (low pad voicing). */
export const CHORDS: readonly (readonly number[])[] = [
  [45, 52, 57, 60, 64],
  [41, 48, 55, 57, 64],
  [48, 55, 59, 62, 67],
  [43, 50, 59, 62, 64],
];

/** A minor pentatonic across two octaves for the soft plucks. */
export const PENTATONIC: readonly number[] = [69, 72, 74, 76, 79, 81, 84];

/** Picks a pluck note for chord `index` using `rand` in [0, 1). Never the same note twice in a row. */
export function pluckNote(rand: number, previous: number | null): number {
  let i = Math.floor(rand * PENTATONIC.length) % PENTATONIC.length;
  if (PENTATONIC[i] === previous) i = (i + 1) % PENTATONIC.length;
  return PENTATONIC[i]!;
}
