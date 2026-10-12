import { css } from "../sky/color";
import { ambienceUnset, ambienceWanted, onAmbienceChange, setAmbience } from "./ambience";
import { toast } from "./toast";
import { overcastOf } from "../sky/engine";
import { MOODS, currentMood, onMoodChange, setMood, type Mood } from "../sky/moods";
import { skyGradient } from "../sky/palette";
import { DEFAULT_INPUTS } from "../sky/types";
import { t } from "../core/i18n";

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

  // A switch for the mood's sound sits at the top of the panel.
  const sound = document.createElement("button");
  sound.type = "button";
  sound.className = "mood-sound";
  sound.addEventListener("click", () => setAmbience(!ambienceWanted()));
  panel.appendChild(sound);
  const paintSound = () => {
    sound.textContent = ambienceWanted() ? t("Mood sound: on") : t("Mood sound: off");
    sound.setAttribute("aria-pressed", String(ambienceWanted()));
  };
  paintSound();
  onAmbienceChange(paintSound);

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
    name.textContent = t(mood.name);
    const hint = document.createElement("small");
    hint.textContent = t(mood.hint);
    text.append(name, hint);

    btn.append(swatch, text);
    btn.addEventListener("click", () => {
      setMood(mood.id);
      menu.open = false; // the choice is made: close the panel
      // The first time someone picks a mood, its sound starts too (this click lets the browser play it).
      if (mood.id !== "live" && ambienceUnset()) {
        setAmbience(true);
        toast(t("Mood sound is on. You can switch it off in the mood menu."));
      }
    });
    panel.appendChild(btn);
    buttons.set(mood.id, btn);
  }

  const paint = () => {
    const active = currentMood();
    for (const [id, btn] of buttons) btn.setAttribute("aria-pressed", String(id === active.id));
    if (label) label.textContent = active.id === "live" ? t("Mood") : t(active.name);
  };
  paint();
  onMoodChange(paint);

  document.addEventListener("click", (e) => {
    // A tap outside the panel or on the dimmed backdrop (which belongs to the menu itself) closes it.
    if (menu.open && (!menu.contains(e.target as Node) || e.target === menu)) menu.open = false;
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && menu.open) {
      menu.open = false;
      menu.querySelector("summary")?.focus();
    }
  });
}
