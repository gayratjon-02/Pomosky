import { maybeById } from "../core/dom";
import type { TimerView } from "../core/protocol";
import type { RoomClient } from "../core/room-client";
import { load, save } from "../core/storage";
import { pushActive, pushSupported, subscribePush, unsubscribePush } from "./push";
import { toast } from "./toast";

const KEY = "pomosky.notify";
const ASKED_KEY = "pomosky.notify-asked";
const ASK_DELAY_MS = 4000;

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
  toast(remote ? "You will get an alert when a phase ends, even if this page is closed" : "You will get an alert when a phase ends while this page is open");
  return true;
}

/** A small card, shown once, that offers to turn alerts on. The button click is the permission request's gesture. */
function offerAlerts(roomId: string, onChange: () => void): void {
  const card = document.createElement("div");
  card.className = "alert-ask glass";
  card.setAttribute("role", "dialog");
  card.setAttribute("aria-label", "Phase alerts");
  const text = document.createElement("p");
  text.textContent = "Get an alert when a focus or break ends, even if you close this page.";
  const yes = document.createElement("button");
  yes.type = "button";
  yes.className = "btn btn-primary btn-sm";
  yes.textContent = "Turn on";
  const no = document.createElement("button");
  no.type = "button";
  no.className = "btn btn-ghost btn-sm";
  no.textContent = "Not now";
  const row = document.createElement("div");
  row.className = "row";
  row.append(yes, no);
  card.append(text, row);
  const close = () => {
    save(ASKED_KEY, "1"); // asked once; the Alerts button is still there
    card.remove();
  };
  yes.addEventListener("click", async () => {
    close();
    await turnOn(roomId);
    onChange();
  });
  no.addEventListener("click", close);
  document.body.appendChild(card);
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

  if (Notification.permission === "default" && load(ASKED_KEY) === null) {
    window.setTimeout(() => {
      if (Notification.permission === "default" && load(ASKED_KEY) === null) offerAlerts(roomId, paint);
    }, ASK_DELAY_MS);
  }
}
