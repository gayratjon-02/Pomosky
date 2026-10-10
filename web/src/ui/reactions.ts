import { maybeById } from "../core/dom";

const EMOJI: Record<string, string> = { clap: "👏", coffee: "☕", fire: "🔥", heart: "❤️" };
const SEND_COOLDOWN_MS = 180;
const MAX_FLOATING = 40;

/** Wires the reaction buttons; returns a function that animates a reaction from anyone. */
export function initReactions(send: (kind: string) => void): (kind: string) => void {
  const bar = maybeById("reactions");
  const layer = maybeById("reaction-layer");
  const reduced = window.matchMedia("(prefers-reduced-motion: reduce)");
  let lastSent = 0;

  const float = (kind: string) => {
    if (!layer || reduced.matches || layer.childElementCount >= MAX_FLOATING) return;
    const el = document.createElement("span");
    el.className = "float-emoji";
    el.textContent = EMOJI[kind] ?? "✨";
    el.style.left = `${10 + Math.random() * 80}%`;
    el.style.setProperty("--drift", `${(Math.random() - 0.5) * 120}px`);
    el.style.animationDuration = `${2.4 + Math.random() * 1.2}s`;
    el.addEventListener("animationend", () => el.remove());
    layer.appendChild(el);
  };

  bar?.addEventListener("click", (e) => {
    const button = (e.target as HTMLElement).closest<HTMLButtonElement>("button[data-reaction]");
    const kind = button?.dataset["reaction"];
    if (!button || !kind) return;
    const now = performance.now();
    if (now - lastSent < SEND_COOLDOWN_MS) return;
    lastSent = now;
    send(kind);
    float(kind);
    button.classList.remove("pop");
    void button.offsetWidth; // restart the animation
    button.classList.add("pop");
  });

  return float;
}

/** Updates the small counters next to each reaction button. */
export function setReactionCounts(counts: Record<string, number>): void {
  for (const [kind, n] of Object.entries(counts)) {
    const el = maybeById(`count-${kind}`);
    if (el) el.textContent = n > 0 ? String(n) : "";
  }
}
