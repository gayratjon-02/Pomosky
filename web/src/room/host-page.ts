import { maybeById } from "../core/dom";
import { RoomClient } from "../core/room-client";
import { load, save } from "../core/storage";
import { createSky } from "../sky";
import { astroInputs } from "../sky/inputs";
import { skyPlace } from "../sky/place";
import { createConnectionChip } from "../ui/connection";
import { setViewers } from "../ui/presence";
import { initShare } from "../ui/share";
import { TimerRenderer } from "../ui/timer-view";
import { toast } from "../ui/toast";
import { PLACES } from "../core/places";


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

  const canvas = maybeById<HTMLCanvasElement>("sky");
  const sky = canvas ? createSky(canvas) : null;
  const refreshSky = () => {
    const p = skyPlace(client?.state ?? null);
    sky?.update(astroInputs(new Date(), p.lat, p.lon));
  };
  refreshSky();
  sky?.start();
  window.setInterval(refreshSky, 30_000);

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
    refreshSky();
    setViewers(s.viewers);
    if (pause) pause.textContent = s.timer.phase === "paused" ? "Resume" : "Pause";
  });

  window.setInterval(() => {
    if (client.state) renderer.render(client.state.timer, client.clock.serverNow());
  }, 250);

  const label = () => maybeById<HTMLInputElement>("label")?.value.trim() ?? "";
  maybeById("btn-focus")?.addEventListener("click", () =>
    client.start({ phase: "focus", durationMs: num("focus-min", 25) * 60_000, label: label() }),
  );
  maybeById("btn-break")?.addEventListener("click", () =>
    client.start({ phase: "break", durationMs: num("break-min", 5) * 60_000, label: label() }),
  );
  pause?.addEventListener("click", () => (client.state?.timer.phase === "paused" ? client.resume() : client.pause()));
  maybeById("btn-reset")?.addEventListener("click", () => client.reset());

  maybeById("presets")?.addEventListener("click", (e) => {
    const b = (e.target as HTMLElement).closest<HTMLButtonElement>("button[data-focus]");
    if (!b) return;
    maybeById<HTMLInputElement>("focus-min")!.value = b.dataset["focus"] ?? "25";
    maybeById<HTMLInputElement>("break-min")!.value = b.dataset["break"] ?? "5";
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

  client.connect();
}
