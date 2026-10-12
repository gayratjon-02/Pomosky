import { maybeById } from "../core/dom";
import type { TimerView } from "../core/protocol";
import type { RoomClient } from "../core/room-client";
import { sessionContent, shareCard } from "./share-card";

/**
 * The "Share this session" button: it appears when a focus phase ends and goes away
 * when the next one starts. Returns the function to call when a phase ends.
 */
export function initSessionCard(client: RoomClient): (t: TimerView) => void {
  const btn = maybeById<HTMLButtonElement>("share-session");
  let last: { startedAtMs: number; durationMs: number; label: string } | null = null;
  if (!btn) return () => {};

  client.on("state", (s) => {
    if (last && s.timer.startedAtMs !== last.startedAtMs) {
      last = null;
      btn.hidden = true;
    }
  });
  btn.addEventListener("click", () => {
    if (last) void shareCard(btn, sessionContent(last.durationMs, last.label), "pomosky-session.png");
  });

  return (t) => {
    if (t.phase !== "focus") return;
    last = { startedAtMs: t.startedAtMs, durationMs: t.durationMs, label: t.label };
    btn.hidden = false;
  };
}
