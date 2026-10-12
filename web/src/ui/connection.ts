import { maybeById } from "../core/dom";
import type { SocketStatus } from "../core/ws";
import { toast } from "./toast";

const TEXT: Record<SocketStatus, string> = {
  connecting: "Connecting",
  open: "Live",
  reconnecting: "Reconnecting",
  closed: "Disconnected",
};

/** A short drop (a server restart, a network blip) heals itself; only a longer outage is worth telling the user about. */
const OUTAGE_TOAST_MS = 8000;

/** Returns a status handler that paints the connection chip and toasts on connection loss. */
export function createConnectionChip(): (status: SocketStatus) => void {
  const dot = maybeById("conn-dot");
  const text = maybeById("conn-text");
  let wasOpen = false;
  let outageTimer: number | undefined;
  let told = false;
  return (status) => {
    if (text) text.textContent = TEXT[status];
    if (dot) dot.className = `dot ${status === "open" ? "dot-live" : status === "reconnecting" ? "dot-warn" : "dot-off"}`;
    if (status === "reconnecting" && wasOpen && outageTimer === undefined && !told) {
      outageTimer = window.setTimeout(() => {
        outageTimer = undefined;
        told = true;
        toast("Connection lost, reconnecting");
      }, OUTAGE_TOAST_MS);
    }
    if (status === "open") {
      window.clearTimeout(outageTimer);
      outageTimer = undefined;
      if (told) toast("Back online");
      told = false;
    }
    if (status === "open") wasOpen = true;
  };
}
