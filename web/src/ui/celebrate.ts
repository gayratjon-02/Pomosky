// A short burst of sparkles when a phase ends. Plain DOM and CSS, no canvas, and
// nothing at all for people who asked for reduced motion.

const COLORS = ["#8ab4ff", "#b79bff", "#7be0b0", "#ffb36b", "#ffe08a", "#ff8fb1"];
const COUNT = 36;
const LIFETIME_MS = 2200;

export function celebrate(): void {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  const layer = document.createElement("div");
  layer.className = "celebration";
  layer.setAttribute("aria-hidden", "true");

  for (let i = 0; i < COUNT; i++) {
    const p = document.createElement("span");
    p.className = i % 3 === 0 ? "spark spark-star" : "spark";
    const angle = (Math.PI * 2 * i) / COUNT + Math.random() * 0.4;
    const distance = 120 + Math.random() * 280;
    p.style.setProperty("--dx", `${Math.cos(angle) * distance}px`);
    p.style.setProperty("--dy", `${Math.sin(angle) * distance - 60}px`); // a little upward bias
    p.style.setProperty("--rot", `${Math.round((Math.random() - 0.5) * 540)}deg`);
    p.style.setProperty("--size", `${6 + Math.random() * 8}px`);
    p.style.setProperty("--delay", `${Math.round(Math.random() * 120)}ms`);
    p.style.background = COLORS[i % COLORS.length]!;
    layer.appendChild(p);
  }
  document.body.appendChild(layer);
  window.setTimeout(() => layer.remove(), LIFETIME_MS);
}
