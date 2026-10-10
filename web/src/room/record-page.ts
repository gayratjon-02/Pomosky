import { maybeById } from "../core/dom";
import { RoomClient } from "../core/room-client";
import { createSky } from "../sky";
import { astroInputs } from "../sky/inputs";
import { createConnectionChip } from "../ui/connection";
import { setViewers } from "../ui/presence";
import { initReactions } from "../ui/reactions";
import { renderQr } from "../ui/qr";
import { TimerRenderer } from "../ui/timer-view";

const DEFAULT_LAT = 37.57; // TODO(stage 6): the host's rounded location
const DEFAULT_LON = 126.98;

/** Record mode: a clean full-screen view with the clock, the live sky, and a QR code to join. */
export function initRecordPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const canvas = maybeById<HTMLCanvasElement>("sky");
  const sky = canvas ? createSky(canvas) : null;
  const refreshSky = () => sky?.update(astroInputs(new Date(), DEFAULT_LAT, DEFAULT_LON));
  refreshSky();
  sky?.start();
  window.setInterval(refreshSky, 30_000);

  const url = maybeById("join-url")?.dataset["url"];
  const qr = maybeById("qr");
  if (url && qr) {
    renderQr(qr, url);
    const text = maybeById("join-url");
    if (text) text.textContent = url.replace(/^https?:\/\//, "");
  }

  const client = new RoomClient(roomId);
  const renderer = new TimerRenderer();
  const floatReaction = initReactions(() => {});
  client.on("status", createConnectionChip());
  client.on("state", (s) => setViewers(s.viewers));
  client.on("tick", (t) => floatReaction(t.kind));
  window.setInterval(() => {
    if (client.state) renderer.render(client.state.timer, client.clock.serverNow());
  }, 250);
  client.connect();
}
