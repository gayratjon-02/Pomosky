import { maybeById } from "../core/dom";
import { t } from "../core/i18n";
import { toast } from "./toast";

// The host's settings panel. On wide screens it sits beside the timer. On phones it is
// a bottom sheet that opens from the "Settings" button, so the timer stays the hero of the screen.

const PHONE = "(max-width: 900px)";
const MIN = 1;
const MAX = 720;

export function initSettingsSheet(): void {
  const panel = maybeById("host-panel");
  const openBtn = maybeById<HTMLButtonElement>("open-settings");
  const closeBtn = maybeById<HTMLButtonElement>("close-settings");
  if (!panel) return;
  const phone = window.matchMedia(PHONE);

  // The backdrop lives next to the panel (see host.html) so both share one stacking context.
  const backdrop = maybeById("sheet-backdrop");
  if (!backdrop) return;

  const isOpen = () => panel.hasAttribute("data-open");
  const open = () => {
    if (!phone.matches || isOpen()) return;
    panel.setAttribute("data-open", "");
    backdrop.setAttribute("data-open", "");
    document.body.classList.add("sheet-open");
    openBtn?.setAttribute("aria-expanded", "true");
    closeBtn?.focus({ preventScroll: true });
  };
  const close = () => {
    if (!isOpen()) return;
    panel.removeAttribute("data-open");
    backdrop.removeAttribute("data-open");
    document.body.classList.remove("sheet-open");
    openBtn?.setAttribute("aria-expanded", "false");
    openBtn?.focus({ preventScroll: true });
  };

  openBtn?.addEventListener("click", open);
  closeBtn?.addEventListener("click", close);
  backdrop.addEventListener("click", close);
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && isOpen()) close();
  });
  phone.addEventListener("change", () => {
    if (!phone.matches) close();
  });

  // − and + next to the minutes: big targets for a thumb.
  document.querySelectorAll<HTMLButtonElement>(".stepper button").forEach((btn) => {
    btn.addEventListener("click", () => {
      const input = maybeById<HTMLInputElement>(btn.dataset["for"] ?? "");
      if (!input || input.disabled) return;
      const next = Math.min(MAX, Math.max(MIN, (Number(input.value) || MIN) + Number(btn.dataset["step"] ?? 0)));
      input.value = String(next);
      input.dispatchEvent(new Event("input", { bubbles: true }));
    });
  });

  // Share straight from the main screen: the phone's share sheet when there is one, else copy the link.
  maybeById("share-quick")?.addEventListener("click", async () => {
    const url = maybeById("copy-link")?.dataset["url"] ?? location.href.split("#")[0]!;
    if (navigator.share) {
      try {
        await navigator.share({ title: "Pomosky", text: t("Work with me"), url });
      } catch {
        // closed without sharing
      }
      return;
    }
    try {
      await navigator.clipboard.writeText(url);
      toast(t("Link copied"));
    } catch {
      toast(t("Could not copy. Select the link and copy it by hand."));
    }
  });
}
