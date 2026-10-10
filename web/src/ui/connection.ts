import { maybeById } from "../core/dom";
import type { SocketStatus } from "../core/ws";
import { toast } from "./toast";

const TEXT: Record<SocketStatus, string> = {
  connecting: "Connecting",
  open: "Live",
  reconnecting: "Reconnecting",
  closed: "Disconnected",
};

/** Returns a status handler that paints the connection chip and toasts on connection loss. */
export function createConnectionChip(): (status: SocketStatus) => void {
  const dot = maybeById("conn-dot");
  const text = maybeById("conn-text");
  let wasOpen = false;
  return (status) => {
    if (text) text.textContent = TEXT[status];
    if (dot) dot.className = `dot ${status === "open" ? "dot-live" : status === "reconnecting" ? "dot-warn" : "dot-off"}`;
    if (status === "reconnecting" && wasOpen) toast("Connection lost, reconnecting");
    if (status === "open" && wasOpen) toast("Back online");
    if (status === "open") wasOpen = true;
  };
}
