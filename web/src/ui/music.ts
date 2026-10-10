import { load, save } from "../core/storage";
import { CHORDS, midiToFreq, pluckNote } from "./music-theory";

// Calm background music, made live with the Web Audio API (no audio files, nothing
// to license): slow pad chords, a few soft plucks and a breath of filtered noise.
// It has its own switch, off by default, and the choice is remembered.

const KEY = "pomosky.music";
const CHORD_SECONDS = 9;
const LEVEL = 0.18;

let ctx: AudioContext | null = null;
let master: GainNode | null = null;
let timer: number | undefined;
let on = false;
let chordIndex = 0;
let lastPluck: number | null = null;
let airStarted = false;

export function musicWanted(): boolean {
  return load(KEY) === "on";
}

export function musicPlaying(): boolean {
  return on;
}

/** Starts or stops the music. Call it from a click so the browser allows audio. */
export function setMusic(wanted: boolean): void {
  save(KEY, wanted ? "on" : "off");
  if (wanted) start();
  else stop();
}

function start(): void {
  if (on) return;
  try {
    ctx ??= new AudioContext();
    void ctx.resume();
  } catch {
    return;
  }
  if (!master) {
    master = ctx.createGain();
    master.gain.value = 0;
    master.connect(ctx.destination);
  }
  on = true;
  master.gain.cancelScheduledValues(ctx.currentTime);
  master.gain.setTargetAtTime(LEVEL, ctx.currentTime, 1.5); // fade in
  startAir();
  nextChord();
}

function stop(): void {
  if (!on || !ctx || !master) return;
  on = false;
  window.clearTimeout(timer);
  master.gain.cancelScheduledValues(ctx.currentTime);
  master.gain.setTargetAtTime(0, ctx.currentTime, 0.6); // fade out
}

function nextChord(): void {
  if (!on || !ctx || !master) return;
  const chord = CHORDS[chordIndex % CHORDS.length]!;
  chordIndex++;
  const t = ctx.currentTime;
  for (const note of chord) pad(midiToFreq(note), t);
  const plucks = 2 + Math.floor(Math.random() * 3);
  for (let i = 0; i < plucks; i++) {
    const n = pluckNote(Math.random(), lastPluck);
    lastPluck = n;
    pluck(midiToFreq(n), t + 1 + Math.random() * (CHORD_SECONDS - 3));
  }
  timer = window.setTimeout(nextChord, CHORD_SECONDS * 1000 - 600); // chords overlap a little
}

/** One pad voice: two slightly detuned sines with a slow swell, through a soft low-pass. */
function pad(freq: number, t: number): void {
  if (!ctx || !master) return;
  const filter = ctx.createBiquadFilter();
  filter.type = "lowpass";
  filter.frequency.value = 900;
  const gain = ctx.createGain();
  gain.gain.setValueAtTime(0.0001, t);
  gain.gain.linearRampToValueAtTime(0.05, t + 3);
  gain.gain.linearRampToValueAtTime(0.0001, t + CHORD_SECONDS + 1.5);
  filter.connect(gain).connect(master);
  for (const detune of [-6, 6]) {
    const osc = ctx.createOscillator();
    osc.type = "sine";
    osc.frequency.value = freq;
    osc.detune.value = detune;
    osc.connect(filter);
    osc.start(t);
    osc.stop(t + CHORD_SECONDS + 2);
  }
}

/** A soft bell-like note that fades away. */
function pluck(freq: number, t: number): void {
  if (!ctx || !master) return;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = "triangle";
  osc.frequency.value = freq;
  gain.gain.setValueAtTime(0.0001, t);
  gain.gain.exponentialRampToValueAtTime(0.07, t + 0.02);
  gain.gain.exponentialRampToValueAtTime(0.0001, t + 2.8);
  osc.connect(gain).connect(master);
  osc.start(t);
  osc.stop(t + 3);
}

/** Looping filtered noise, like a quiet breeze under the chords. */
function startAir(): void {
  if (airStarted || !ctx || !master) return;
  airStarted = true;
  const length = ctx.sampleRate * 4;
  const buffer = ctx.createBuffer(1, length, ctx.sampleRate);
  const data = buffer.getChannelData(0);
  for (let i = 0; i < length; i++) data[i] = Math.random() * 2 - 1;
  const src = ctx.createBufferSource();
  src.buffer = buffer;
  src.loop = true;
  const filter = ctx.createBiquadFilter();
  filter.type = "lowpass";
  filter.frequency.value = 380;
  const gain = ctx.createGain();
  gain.gain.value = 0.035;
  src.connect(filter).connect(gain).connect(master);
  src.start();
}
