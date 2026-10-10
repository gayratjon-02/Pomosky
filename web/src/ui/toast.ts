// Small non-blocking messages ("connection lost", "link copied").

const VISIBLE_MS = 3200;

function region(): HTMLElement {
  let el = document.getElementById("toasts");
  if (!el) {
    el = document.createElement("div");
    el.id = "toasts";
    el.className = "toast-region";
    el.setAttribute("role", "status");
    el.setAttribute("aria-live", "polite");
    document.body.appendChild(el);
  }
  return el;
}

export function toast(message: string): void {
  const el = document.createElement("div");
  el.className = "toast";
  el.textContent = message;
  region().appendChild(el);
  window.setTimeout(() => {
    el.dataset["leaving"] = "";
    window.setTimeout(() => el.remove(), 300);
  }, VISIBLE_MS);
}
