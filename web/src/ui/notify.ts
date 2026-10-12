import { maybeById } from "../core/dom";
import type { TimerView } from "../core/protocol";
import type { RoomClient } from "../core/room-client";
import { load, save } from "../core/storage";
import { pushActive, pushSupported, subscribePush, unsubscribePush } from "./push";
import { toast } from "./toast";
import { t as tr } from "../core/i18n";

const KEY = "pomosky.notify";
const ASKED_KEY = "pomosky.notify-asked";
const ASK_DELAY_MS = 1500;

/** The text of the "phase ended" notification. */
export function endMessage(t: TimerView): { title: string; body: string } {
  const label = t.label.trim();
  if (t.phase === "break") return { title: tr("Break is over"), body: tr("Ready for the next focus session?") };
  return { title: tr("Focus session finished"), body: label ? tr("{label}. Time for a break.", { label }) : tr("Time for a break.") };
}

function supported(): boolean {
  return typeof Notification !== "undefined";
}

function enabled(): boolean {
  return supported() && Notification.permission === "granted" && load(KEY) === "on";
}

/** Shows a notification when a phase ends and the tab is in the background. */
export function notifyEnd(t: TimerView): void {
  // With Web Push the service worker shows the alert (also when the page is closed).
  if (!enabled() || !document.hidden || pushActive()) return;
  const { title, body } = endMessage(t);
  try {
    new Notification(title, { body, icon: "/static/favicon.svg", tag: "pomosky-end" });
  } catch {
    // some mobile browsers only allow notifications from a service worker: ignore
  }
}

/** Asks for permission and subscribes; returns whether alerts are on now. */
async function turnOn(roomId: string): Promise<boolean> {
  const result = Notification.permission === "granted" ? "granted" : await Notification.requestPermission();
  if (result !== "granted") return false;
  save(KEY, "on");
  const remote = await subscribePush(roomId);
  toast(remote ? tr("You will get an alert when a phase ends, even if this page is closed") : tr("You will get an alert when a phase ends while this page is open"));
  return true;
}

/**
 * Shows the browser's own permission prompt on arrival, once. Chrome allows that
 * without a click; Firefox and Safari only show it after one, so there it waits for
 * the first click or key press. A prompt that was answered or dismissed is not repeated.
 */
function askOnArrival(roomId: string, onChange: () => void): void {
  const pending = () => Notification.permission === "default" && load(ASKED_KEY) === null;
  const ask = async () => {
    if (!pending()) return;
    const started = performance.now();
    const result = await Notification.requestPermission().catch(() => "default" as NotificationPermission);
    if (result === "granted") {
      save(ASKED_KEY, "1");
      await turnOn(roomId);
      onChange();
    } else if (performance.now() - started > 400 || result === "denied") {
      save(ASKED_KEY, "1"); // the person saw the prompt and chose; do not ask again
      onChange();
    } else {
      waitForGesture(); // no prompt appeared without a gesture
    }
  };
  const waitForGesture = () => {
    const once = () => {
      window.removeEventListener("pointerdown", once, true);
      window.removeEventListener("keydown", once, true);
      void ask();
    };
    window.addEventListener("pointerdown", once, true);
    window.addEventListener("keydown", once, true);
  };
  window.setTimeout(() => void ask(), ASK_DELAY_MS);
}

/**
 * Wires the #notify-toggle button and the first-visit offer. Alerts use Web Push
 * when the server has it (they arrive with the page closed) and fall back to
 * in-page notifications. The subscription is refreshed whenever the room connects,
 * because the server forgets it after a restart.
 */
export function initNotifyToggle(client: RoomClient, roomId: string): void {
  const btn = maybeById<HTMLButtonElement>("notify-toggle");
  if (!btn) return;
  if (!supported()) {
    btn.hidden = true;
    return;
  }
  const paint = () => {
    if (Notification.permission === "denied") {
      btn.textContent = tr("Alerts blocked");
      btn.disabled = true;
      btn.title = tr("Allow notifications for this site in your browser settings");
      return;
    }
    btn.textContent = enabled() ? tr("Alerts on") : tr("Alerts off");
    btn.setAttribute("aria-pressed", String(enabled()));
  };
  btn.addEventListener("click", async () => {
    if (enabled()) {
      save(KEY, "off");
      await unsubscribePush();
    } else {
      await turnOn(roomId);
    }
    paint();
  });
  client.on("status", (status) => {
    if (status === "open" && enabled() && pushSupported()) void subscribePush(roomId);
  });
  paint();

  askOnArrival(roomId, paint);
}
