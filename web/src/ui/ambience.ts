import { load, save } from "../core/storage";
import { onBolt, onSky } from "../sky/sky-feed";
import { AmbienceEngine } from "./ambience-engine";
import { sceneFor } from "./ambience-scene";

// The sound that goes with the sky: whatever mood is chosen (or the live sky) decides
// what you hear, see ambience-scene.ts. It has its own switch, off until the visitor
// turns it on, and the choice is remembered.

const KEY = "pomosky.ambience";
const LEVEL = 0.8;
const LOOKAHEAD = 12; // seconds of birds, crickets and so on planned ahead (timers are slow in background tabs)

let ctx: AudioContext | null = null;
let master: GainNode | null = null;
let engine: AmbienceEngine | null = null;
let timer: number | undefined;
let stopSky: (() => void) | undefined;
let stopBolt: (() => void) | undefined;
let suspendTimer: number | undefined;
let on = false;
const listeners = new Set<() => void>();

/** Whether the visitor wants ambience (it may still wait for the first click to be allowed to play). */
export const ambienceWanted = (): boolean => load(KEY) === "on";

/** True when the visitor has never made a choice. */
export const ambienceUnset = (): boolean => load(KEY) === null;

export function onAmbienceChange(fn: () => void): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

/** Turns the ambience on or off. Call it from a click so the browser allows audio. */
export function setAmbience(wanted: boolean): void {
  save(KEY, wanted ? "on" : "off");
  if (wanted) start();
  else stop();
  listeners.forEach((fn) => fn());
}

function start(): void {
  if (on) return;
  try {
    ctx ??= new AudioContext();
    void ctx.resume();
  } catch {
    return;
  }
  window.clearTimeout(suspendTimer);
  if (!master || !engine) {
    master = ctx.createGain();
    master.gain.value = 0;
    master.connect(ctx.destination);
    engine = new AmbienceEngine(ctx, master);
  }
  on = true;
  const c = ctx;
  const e = engine;
  master.gain.cancelScheduledValues(c.currentTime);
  master.gain.setTargetAtTime(LEVEL, c.currentTime, 1.2); // fade in
  stopSky = onSky((sky) => e.setScene(sceneFor(sky)));
  stopBolt = onBolt(() => e.thunderAt(c.currentTime + 0.3 + Math.random() * 2.2)); // thunder follows the flash
  const plan = () => e.schedule(c.currentTime + LOOKAHEAD);
  plan();
  timer = window.setInterval(plan, 1000);
}

function stop(): void {
  if (!on || !ctx || !master) return;
  on = false;
  window.clearInterval(timer);
  stopSky?.();
  stopBolt?.();
  const c = ctx;
  master.gain.cancelScheduledValues(c.currentTime);
  master.gain.setTargetAtTime(0, c.currentTime, 0.4); // fade out
  suspendTimer = window.setTimeout(() => {
    if (!on) void c.suspend(); // stop the audio thread while it is silent
  }, 2500);
}

/** Browsers keep audio locked until the page is used; the first click or key press unlocks it. */
function unlockOnFirstGesture(): void {
  const unlock = () => {
    window.removeEventListener("pointerdown", unlock, true);
    window.removeEventListener("keydown", unlock, true);
    if (ctx && on) void ctx.resume();
  };
  window.addEventListener("pointerdown", unlock, true);
  window.addEventListener("keydown", unlock, true);
}

/** Wires the #ambience-toggle button (room pages) and restores the saved choice. */
export function initAmbience(): void {
  const btn = document.getElementById("ambience-toggle") as HTMLButtonElement | null;
  const paint = () => {
    if (!btn) return;
    btn.textContent = ambienceWanted() ? "Ambience on" : "Ambience off";
    btn.setAttribute("aria-pressed", String(ambienceWanted()));
  };
  btn?.addEventListener("click", () => setAmbience(!ambienceWanted()));
  onAmbienceChange(paint);
  paint();
  if (ambienceWanted()) {
    start();
    unlockOnFirstGesture();
  }
}
