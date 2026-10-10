/** Closes the account menu when you click elsewhere or press Escape. */
export function initUserMenu(): void {
  const menu = document.getElementById("user-menu") as HTMLDetailsElement | null;
  if (!menu) return;
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
