import { maybeById } from "../core/dom";
import { musicPlaying, musicWanted, setMusic } from "./music";

/**
 * Wires the #music-toggle button. Browsers need a gesture before audio, so a
 * remembered "on" starts with the first click or key press anywhere on the page.
 */
export function initMusic(): void {
  const toggle = maybeById<HTMLButtonElement>("music-toggle");
  const paint = () => {
    if (!toggle) return;
    toggle.textContent = musicPlaying() ? "Music on" : "Music off";
    toggle.setAttribute("aria-pressed", String(musicPlaying()));
  };
  toggle?.addEventListener("click", () => {
    setMusic(!musicPlaying());
    paint();
  });
  if (musicWanted()) {
    const resume = (e: Event) => {
      if (toggle && e.target instanceof Node && toggle.contains(e.target)) return; // the button decides itself
      if (!musicPlaying() && musicWanted()) setMusic(true);
      paint();
    };
    document.addEventListener("pointerdown", resume, { capture: true });
    document.addEventListener("keydown", resume, { capture: true });
  }
  paint();
}
