import { maybeById } from "../core/dom";
import type { RoomClient } from "../core/room-client";
import type { TimerView } from "../core/protocol";
import { cueFor } from "./sound-cues";
import {
  playClick,
  playPause,
  playReaction,
  playReset,
  playResume,
  playStart,
  setSoundEnabled,
  soundEnabled,
  unlockAudio,
} from "./sound";

const CLICKABLE = ".btn, button.chip, .reactions button, .tabs button, .user-menu-btn";

/**
 * Connects the sounds to the page: a tick on every button, the cues of the room
 * timer (the same for the host and every viewer), a blip for reactions, and the
 * one switch that mutes everything. Needs a #sound-toggle button for the switch.
 */
export function initSounds(client: RoomClient): void {
  // Browsers allow audio only after a gesture: the first touch or click unlocks it.
  document.addEventListener("pointerdown", unlockAudio, { once: true, capture: true });

  const toggle = maybeById<HTMLButtonElement>("sound-toggle");
  const paint = () => {
    if (!toggle) return;
    toggle.textContent = soundEnabled() ? "Sound on" : "Sound off";
    toggle.setAttribute("aria-pressed", String(soundEnabled()));
  };
  toggle?.addEventListener("click", () => {
    setSoundEnabled(!soundEnabled());
    paint();
  });
  paint();

  document.addEventListener("click", (e) => {
    const el = (e.target as HTMLElement).closest(CLICKABLE);
    if (el && el !== toggle) playClick();
  });

  let prev: TimerView | null = null;
  client.on("state", (s) => {
    switch (cueFor(prev, s.timer)) {
      case "start":
        playStart();
        break;
      case "pause":
        playPause();
        break;
      case "resume":
        playResume();
        break;
      case "reset":
        playReset();
        break;
    }
    prev = s.timer;
  });
  client.on("tick", () => playReaction());
}
