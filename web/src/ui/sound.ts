import { load, save } from "../core/storage";

// Every sound is synthesised with the Web Audio API, so there are no audio files to
// download or license. One switch turns all of them off; the choice is remembered.

const KEY = "pomosky.sound";
let ctx: AudioContext | null = null;
let enabled = load(KEY) !== "off";
let lastClick = 0;
let lastReaction = 0;

export function soundEnabled(): boolean {
  return enabled;
}

export function setSoundEnabled(on: boolean): void {
  enabled = on;
  save(KEY, on ? "on" : "off");
  if (on) {
    unlockAudio(); // created inside the click that enabled it, so browsers allow playback later
    playToggleOn();
  }
}

/** Browsers only allow audio after a user gesture. Call this from one. */
export function unlockAudio(): void {
  audio();
}

function audio(): AudioContext | null {
  try {
    ctx ??= new AudioContext();
    if (ctx.state === "suspended") void ctx.resume();
    return ctx;
  } catch {
    return null;
  }
}

interface Note {
  freq: number;
  at: number; // seconds from now
  dur: number;
  gain?: number;
  type?: OscillatorType;
}

function play(notes: Note[]): void {
  if (!enabled) return;
  const a = audio();
  if (!a) return;
  const now = a.currentTime;
  for (const n of notes) {
    const osc = a.createOscillator();
    const gain = a.createGain();
    osc.type = n.type ?? "sine";
    osc.frequency.value = n.freq;
    const t = now + n.at;
    const peak = n.gain ?? 0.16;
    gain.gain.setValueAtTime(0.0001, t);
    gain.gain.exponentialRampToValueAtTime(peak, t + 0.012);
    gain.gain.exponentialRampToValueAtTime(0.0001, t + n.dur);
    osc.connect(gain).connect(a.destination);
    osc.start(t);
    osc.stop(t + n.dur + 0.05);
  }
}

/** A soft tick for any button. */
export function playClick(): void {
  const now = performance.now();
  if (now - lastClick < 40) return;
  lastClick = now;
  play([{ freq: 720, at: 0, dur: 0.05, gain: 0.07, type: "triangle" }]);
}

/** A rising three-note figure when a phase starts. */
export function playStart(): void {
  play([
    { freq: 523.25, at: 0, dur: 0.22 },
    { freq: 659.25, at: 0.11, dur: 0.22 },
    { freq: 783.99, at: 0.22, dur: 0.5, gain: 0.18 },
  ]);
}

export function playPause(): void {
  play([
    { freq: 587.33, at: 0, dur: 0.16 },
    { freq: 440, at: 0.1, dur: 0.3 },
  ]);
}

export function playResume(): void {
  play([
    { freq: 440, at: 0, dur: 0.16 },
    { freq: 587.33, at: 0.1, dur: 0.3 },
  ]);
}

export function playReset(): void {
  play([{ freq: 330, at: 0, dur: 0.18, type: "triangle" }, { freq: 247, at: 0.09, dur: 0.26, type: "triangle" }]);
}

/** A tiny blip for a reaction, limited so a burst of reactions stays pleasant. */
export function playReaction(): void {
  const now = performance.now();
  if (now - lastReaction < 350) return;
  lastReaction = now;
  play([{ freq: 988, at: 0, dur: 0.09, gain: 0.07 }]);
}

function playToggleOn(): void {
  play([{ freq: 660, at: 0, dur: 0.12, gain: 0.1 }, { freq: 880, at: 0.07, dur: 0.18, gain: 0.1 }]);
}

/** A soft two-note chime when a phase ends. */
export function chime(): void {
  play([
    { freq: 660, at: 0, dur: 1.1, gain: 0.18 },
    { freq: 880, at: 0.22, dur: 1.1, gain: 0.18 },
  ]);
}
