import { css } from "../sky/color";
import { overcastOf } from "../sky/engine";
import { MOODS, currentMood, onMoodChange, setMood, type Mood } from "../sky/moods";
import { skyGradient } from "../sky/palette";
import { DEFAULT_INPUTS } from "../sky/types";

/** The preview of a mood: its own sky colours from top to horizon, with the sun or moon where it sits. */
function swatchBackground(mood: Mood): string {
  if (!mood.sky) return "conic-gradient(from 200deg, #3f7fd0, #ffb36b, #c4607a, #17236a, #060b2a, #3f7fd0)";
  const inputs = { ...DEFAULT_INPUTS, ...mood.sky };
  const sky = skyGradient(inputs.sunElevationDeg, overcastOf(inputs));
  const gradient = `linear-gradient(180deg, ${css(sky.top)} 0%, ${css(sky.mid)} 55%, ${css(sky.bottom)} 100%)`;
  if (!mood.orb) return gradient;
  const { x, y, color } = mood.orb;
  return `radial-gradient(circle at ${x * 100}% ${y * 100}%, ${color} 0 9%, transparent 22%), ${gradient}`;
}

/** Fills the mood panel in the header and keeps it in step with the chosen mood. */
export function initMoodPicker(): void {
  const menu = document.getElementById("mood-menu") as HTMLDetailsElement | null;
  const panel = document.getElementById("mood-panel");
  if (!menu || !panel) return;
  const label = document.getElementById("mood-btn-name");

  const buttons = new Map<string, HTMLButtonElement>();
  for (const mood of MOODS) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "mood-option";
    btn.dataset["mood"] = mood.id;

    const swatch = document.createElement("span");
    swatch.className = "mood-swatch";
    swatch.style.background = swatchBackground(mood);
    const text = document.createElement("span");
    text.className = "mood-text";
    const name = document.createElement("strong");
    name.textContent = mood.name;
    const hint = document.createElement("small");
    hint.textContent = mood.hint;
    text.append(name, hint);

    btn.append(swatch, text);
    btn.addEventListener("click", () => setMood(mood.id));
    panel.appendChild(btn);
    buttons.set(mood.id, btn);
  }

  const paint = () => {
    const active = currentMood();
    for (const [id, btn] of buttons) btn.setAttribute("aria-pressed", String(id === active.id));
    if (label) label.textContent = active.id === "live" ? "Mood" : active.name;
  };
  paint();
  onMoodChange(paint);

  document.addEventListener("click", (e) => {
    if (menu.open && !menu.contains(e.target as Node)) menu.open = false;
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && menu.open) {
      menu.open = false;
      menu.querySelector("summary")?.focus();
    }
  });
}
