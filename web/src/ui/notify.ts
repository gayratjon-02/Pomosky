import { maybeById } from "../core/dom";
import type { TimerView } from "../core/protocol";
import { load, save } from "../core/storage";
import { toast } from "./toast";

const KEY = "pomosky.notify";

/** The text of the "phase ended" notification. */
export function endMessage(t: TimerView): { title: string; body: string } {
  const label = t.label.trim();
  if (t.phase === "break") return { title: "Break is over", body: "Ready for the next focus session?" };
  return { title: "Focus session finished", body: label ? `${label}. Time for a break.` : "Time for a break." };
}

function supported(): boolean {
  return typeof Notification !== "undefined";
}

function enabled(): boolean {
  return supported() && Notification.permission === "granted" && load(KEY) === "on";
}

/** Shows a notification when a phase ends and the tab is in the background. */
export function notifyEnd(t: TimerView): void {
  if (!enabled() || !document.hidden) return;
  const { title, body } = endMessage(t);
  try {
    new Notification(title, { body, icon: "/static/favicon.svg", tag: "pomosky-end" });
  } catch {
    // some mobile browsers only allow notifications from a service worker: ignore
  }
}

/** Wires the #notify-toggle button: asks for permission once and remembers the choice. */
export function initNotifyToggle(): void {
  const btn = maybeById<HTMLButtonElement>("notify-toggle");
  if (!btn) return;
  if (!supported()) {
    btn.hidden = true;
    return;
  }
  const paint = () => {
    if (Notification.permission === "denied") {
      btn.textContent = "Alerts blocked";
      btn.disabled = true;
      btn.title = "Allow notifications for this site in your browser settings";
      return;
    }
    btn.textContent = enabled() ? "Alerts on" : "Alerts off";
    btn.setAttribute("aria-pressed", String(enabled()));
  };
  btn.addEventListener("click", async () => {
    if (enabled()) {
      save(KEY, "off");
    } else {
      const result = Notification.permission === "granted" ? "granted" : await Notification.requestPermission();
      if (result === "granted") {
        save(KEY, "on");
        toast("You will get an alert when a phase ends");
      }
    }
    paint();
  });
  paint();
}
