import { maybeById } from "../core/dom";
import { RoomClient } from "../core/room-client";
import { startLiveSky } from "../sky/live-sky";
import { createConnectionChip } from "../ui/connection";
import { setViewers } from "../ui/presence";
import { initReactions } from "../ui/reactions";
import { renderQr } from "../ui/qr";
import { TimerRenderer } from "../ui/timer-view";


/** Record mode: a clean full-screen view with the clock, the live sky, and a QR code to join. */
export function initRecordPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const client = new RoomClient(roomId);

  startLiveSky(maybeById<HTMLCanvasElement>("sky"), client, roomId);

  const url = maybeById("join-url")?.dataset["url"];
  const qr = maybeById("qr");
  if (url && qr) {
    renderQr(qr, url);
    const text = maybeById("join-url");
    if (text) text.textContent = url.replace(/^https?:\/\//, "");
  }

  const renderer = new TimerRenderer();
  const floatReaction = initReactions(() => {});
  client.on("status", createConnectionChip());
  client.on("state", (s) => {
    setViewers(s.viewers);
  });
  client.on("tick", (t) => floatReaction(t.kind));
  window.setInterval(() => {
    if (client.state) renderer.render(client.state.timer, client.clock.serverNow());
  }, 250);
  client.connect();
}
