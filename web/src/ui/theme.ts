import { load, save } from "../core/storage";

// The light theme only applies to the plain text pages (privacy, admin, errors).
// Pages that show the live sky keep their own night and day look.

const KEY = "pomosky.theme";
export type Theme = "dark" | "light";

export function currentTheme(): Theme {
  return load(KEY) === "light" ? "light" : "dark";
}

export function applyTheme(theme: Theme): void {
  document.documentElement.dataset["theme"] = theme;
}

/** Applies the saved theme and wires the #theme-toggle button. */
export function initTheme(): void {
  applyTheme(currentTheme());
  const toggle = document.getElementById("theme-toggle");
  if (!toggle) return;
  const paint = () => {
    toggle.textContent = currentTheme() === "light" ? "Light theme" : "Dark theme";
    toggle.setAttribute("aria-pressed", String(currentTheme() === "light"));
  };
  toggle.addEventListener("click", () => {
    const next: Theme = currentTheme() === "light" ? "dark" : "light";
    save(KEY, next);
    applyTheme(next);
    paint();
  });
  paint();
}
