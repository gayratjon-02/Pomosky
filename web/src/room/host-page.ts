import { maybeById } from "../core/dom";
import { RoomClient } from "../core/room-client";
import { load, save } from "../core/storage";
import { startLiveSky } from "../sky/live-sky";
import { createConnectionChip } from "../ui/connection";
import { setViewers } from "../ui/presence";
import { controlsFor } from "../ui/host-controls";
import { initNotifyToggle, notifyEnd } from "../ui/notify";
import { initSounds } from "../ui/sound-events";
import { shortcutFor } from "../ui/shortcuts";
import { initShare } from "../ui/share";
import { initWakeLock } from "../ui/wake-lock";
import { TimerRenderer } from "../ui/timer-view";
import { toast } from "../ui/toast";
import { PLACES } from "../core/places";
import { initPlaceSearch } from "../ui/place-search";


const tokenKey = (roomId: string) => `pomosky.host.${roomId}`;

/**
 * The host token arrives in the URL fragment right after the room is created.
 * It is moved into localStorage and removed from the address bar, so it is not
 * left in history, screenshots or copied links.
 */
function takeHostToken(roomId: string): string | undefined {
  const match = /(?:^#|&)token=([A-Za-z0-9_-]+)/.exec(location.hash);
  if (match?.[1]) {
    save(tokenKey(roomId), match[1]);
    history.replaceState(null, "", location.pathname + location.search);
    return match[1];
  }
  return load(tokenKey(roomId)) ?? undefined;
}

const num = (id: string, fallback: number): number => {
  const v = Number(maybeById<HTMLInputElement>(id)?.value);
  return Number.isFinite(v) && v >= 1 ? Math.min(720, Math.round(v)) : fallback;
};

/** Host panel: authenticates with the host token and sends timer commands. */
export function initHostPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const token = takeHostToken(roomId);
  const client = new RoomClient(roomId, token ? { hostToken: token } : {});

  startLiveSky(maybeById<HTMLCanvasElement>("sky"), client, roomId);

  const renderer = new TimerRenderer();
  const badge = maybeById("host-badge");
  const pause = maybeById<HTMLButtonElement>("btn-pause");
  const controls = Array.from(document.querySelectorAll<HTMLButtonElement>(".host-controls button"));
  initShare();

  const setBadge = (state: "ok" | "denied" | "pending", text: string) => {
    if (!badge) return;
    badge.dataset["state"] = state;
    badge.textContent = text;
    controls.forEach((b) => (b.disabled = state !== "ok"));
  };
  setBadge(token ? "pending" : "denied", token ? "Checking host access" : "No host key in this browser");

  client.on("status", createConnectionChip());
  client.on("hostOk", () => setBadge("ok", "You are the host"));
  client.on("error", (e) => {
    if (e.code === "forbidden") setBadge("denied", "Host key rejected");
    else if (e.code === "invalid_transition") toast("That is not possible right now");
    else if (e.code === "invalid") toast("Check the minutes and the label");
    else if (e.code === "rate_limited") toast("Slow down a little");
  });
  client.on("state", (s) => {
    setViewers(s.viewers);
    applyControls();
  });

  // Which buttons are visible follows the timer: start only when nothing runs, pause or
  // resume while it does, reset once something has started. The server enforces the same rule.
  const focusBtn = maybeById<HTMLButtonElement>("btn-focus");
  const breakBtn = maybeById<HTMLButtonElement>("btn-break");
  const resetBtn = maybeById<HTMLButtonElement>("btn-reset");
  const settings = Array.from(document.querySelectorAll<HTMLInputElement | HTMLButtonElement>("#focus-min, #break-min, #label, #presets button"));
  const applyControls = () => {
    if (!client.state) return;
    const c = controlsFor(client.state.timer, client.clock.serverNow());
    if (focusBtn) focusBtn.hidden = !c.showStart;
    if (breakBtn) breakBtn.hidden = !c.showStart;
    if (pause) {
      pause.hidden = !c.showPause;
      pause.textContent = c.pauseLabel;
    }
    if (resetBtn) resetBtn.hidden = !c.showReset;
    settings.forEach((el) => (el.disabled = !c.editable));
    // After a phase ends, the natural next one is the main button.
    focusBtn?.classList.toggle("btn-primary", c.suggest !== "break");
    focusBtn?.classList.toggle("btn-ghost", c.suggest === "break");
    breakBtn?.classList.toggle("btn-warm", c.suggest !== "focus");
    breakBtn?.classList.toggle("btn-ghost", c.suggest === "focus");
    // An odd number of visible buttons: the last one takes the full row.
    const visible = controls.filter((b) => !b.hidden);
    controls.forEach((b) => b.classList.remove("span-2"));
    if (visible.length % 2 === 1) visible[visible.length - 1]?.classList.add("span-2");
  };

  window.setInterval(() => {
    if (!client.state) return;
    renderer.render(client.state.timer, client.clock.serverNow());
    applyControls();
  }, 250);

  const label = () => maybeById<HTMLInputElement>("label")?.value.trim() ?? "";
  // A line under the buttons says what "start" will do, from the minutes below.
  const planLine = maybeById("plan-line");
  const showPlan = () => {
    if (planLine) planLine.textContent = `${num("focus-min", 25)} min focus, ${num("break-min", 5)} min break`;
  };
  showPlan();
  maybeById("focus-min")?.addEventListener("input", showPlan);
  maybeById("break-min")?.addEventListener("input", showPlan);

  const canStart = () => !!client.state && controlsFor(client.state.timer, client.clock.serverNow()).showStart;
  focusBtn?.addEventListener("click", () => {
    if (canStart()) client.start({ phase: "focus", durationMs: num("focus-min", 25) * 60_000, label: label() });
  });
  breakBtn?.addEventListener("click", () => {
    if (canStart()) client.start({ phase: "break", durationMs: num("break-min", 5) * 60_000, label: label() });
  });
  pause?.addEventListener("click", () => (client.state?.timer.phase === "paused" ? client.resume() : client.pause()));
  maybeById("btn-reset")?.addEventListener("click", () => client.reset());

  maybeById("presets")?.addEventListener("click", (e) => {
    const b = (e.target as HTMLElement).closest<HTMLButtonElement>("button[data-focus]");
    if (!b) return;
    maybeById<HTMLInputElement>("focus-min")!.value = b.dataset["focus"] ?? "25";
    maybeById<HTMLInputElement>("break-min")!.value = b.dataset["break"] ?? "5";
    showPlan();
  });

  // Place of the sky. Picking one sends it to the server, which rounds it to about 1 km.
  const placeSelect = maybeById<HTMLSelectElement>("place");
  if (placeSelect) {
    for (const p of PLACES) placeSelect.add(new Option(p.name, p.name));
    placeSelect.addEventListener("change", () => {
      const p = PLACES.find((x) => x.name === placeSelect.value);
      if (p) client.setLocation({ city: p.name, lat: p.lat, lon: p.lon, tz: p.tz });
    });
    client.on("state", (s) => {
      if (s.city && document.activeElement !== placeSelect) placeSelect.value = s.city;
    });
  }
  initPlaceSearch((p) => {
    client.setLocation({ city: p.name, lat: p.lat, lon: p.lon, tz: p.tz });
    toast(`Sky moved to ${p.name}`);
  });
  maybeById("use-device-place")?.addEventListener("click", () => {
    if (!navigator.geolocation) return toast("Your browser cannot share a location");
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        // Rounded by the server; the exact point is never stored.
        const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
        client.setLocation({ city: "My place", lat: pos.coords.latitude, lon: pos.coords.longitude, tz });
        toast("Sky moved to your place");
      },
      () => toast("Could not read your location"),
      { maximumAge: 600_000, timeout: 10_000 },
    );
  });

  // Keyboard: Space or K pause and resume, F starts focus, B a break, Shift+R resets.
  document.addEventListener("keydown", (e) => {
    if (!client.state || e.repeat) return;
    const t = e.target as HTMLElement;
    const typing = t.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(t.tagName);
    const action = shortcutFor(
      { key: e.key, shiftKey: e.shiftKey, ctrlKey: e.ctrlKey, metaKey: e.metaKey, altKey: e.altKey, typing },
      controlsFor(client.state.timer, client.clock.serverNow()),
    );
    if (!action) return;
    e.preventDefault(); // Space would otherwise scroll the page
    const button = { "toggle-pause": pause, "start-focus": focusBtn, "start-break": breakBtn, reset: resetBtn }[action];
    button?.click(); // the same code path as a mouse click, including the sounds
  });

  initSounds(client);
  renderer.onEnd = notifyEnd;
  initNotifyToggle();
  initWakeLock(client);

  client.connect();
}
