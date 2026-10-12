const COOKIE = "pomosky_lang"; // the server reads the same cookie (internal/i18n)
const YEAR = 60 * 60 * 24 * 365;

/** Wires the language switcher: the choice is kept in a cookie and the page is rendered again in that language. */
export function initLangMenu(): void {
  const menu = document.getElementById("lang-menu") as HTMLDetailsElement | null;
  if (!menu) return;
  const label = menu.querySelector(".mood-btn-name");
  if (label) label.textContent = document.documentElement.lang.toUpperCase();

  menu.querySelectorAll<HTMLButtonElement>(".lang-option").forEach((btn) => {
    btn.addEventListener("click", () => {
      const lang = btn.dataset["lang"];
      if (!lang || lang === document.documentElement.lang) {
        menu.open = false;
        return;
      }
      try {
        const secure = location.protocol === "https:" ? "; Secure" : "";
        document.cookie = `${COOKIE}=${encodeURIComponent(lang)}; Max-Age=${YEAR}; Path=/; SameSite=Lax${secure}`;
      } catch {
        // cookies blocked: the language cannot be remembered
      }
      location.reload();
    });
  });

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
