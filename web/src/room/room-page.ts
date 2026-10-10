import { ClockSync } from "../core/clock-sync";
import { maybeById } from "../core/dom";
import { ReconnectingSocket } from "../core/ws";
import { formatMMSS, remainingMs } from "../core/timer";
import type { StateData } from "../core/protocol";
import { createSky } from "../sky";
import { setViewers } from "../ui/presence";

/** Viewer page: connects to the room, shows the timer and the live sky. */
export function initRoomPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const canvas = maybeById<HTMLCanvasElement>("sky");
  if (canvas) createSky(canvas).start();

  const clock = new ClockSync();
  const digits = maybeById("timer-digits");
  let state: StateData | null = null;

  const proto = location.protocol === "https:" ? "wss" : "ws";
  const socket = new ReconnectingSocket(`${proto}://${location.host}/ws/${roomId}`);

  socket.onMessage = (env) => {
    if (env.type === "state") {
      const data = env.data as StateData;
      state = data;
      clock.add(data.serverNowMs - Date.now());
      setViewers(data.viewers);
    }
  };

  window.setInterval(() => {
    if (state && digits) digits.textContent = formatMMSS(remainingMs(state.timer, clock.serverNow()));
  }, 250);

  socket.connect();
}
