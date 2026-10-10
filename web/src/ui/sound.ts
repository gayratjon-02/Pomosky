import { load, save } from "../core/storage";

const KEY = "pomosky.sound";
let ctx: AudioContext | null = null;
let enabled = load(KEY) !== "off";

export function soundEnabled(): boolean {
  return enabled;
}

export function setSoundEnabled(on: boolean): void {
  enabled = on;
  save(KEY, on ? "on" : "off");
  if (on) audio(); // created inside the click that enabled it, so browsers allow playback later
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

/** A soft two-note chime for the end of a phase. */
export function chime(): void {
  if (!enabled) return;
  const a = audio();
  if (!a) return;
  const now = a.currentTime;
  [660, 880].forEach((freq, i) => {
    const osc = a.createOscillator();
    const gain = a.createGain();
    osc.type = "sine";
    osc.frequency.value = freq;
    const t = now + i * 0.22;
    gain.gain.setValueAtTime(0.0001, t);
    gain.gain.exponentialRampToValueAtTime(0.18, t + 0.03);
    gain.gain.exponentialRampToValueAtTime(0.0001, t + 1.1);
    osc.connect(gain).connect(a.destination);
    osc.start(t);
    osc.stop(t + 1.2);
  });
}
