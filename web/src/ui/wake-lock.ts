import type { RoomClient } from "../core/room-client";

interface WakeLockSentinelLike {
  release(): Promise<void>;
  addEventListener(type: "release", cb: () => void): void;
}

/**
 * Keeps the screen on while a phase is running (focus or break), so a phone on a desk
 * does not go dark in the middle of a session. Where the browser has no Wake Lock
 * API, or refuses, nothing happens. The lock is given back when the timer stops.
 */
export function initWakeLock(client: RoomClient): void {
  const api = (navigator as Navigator & { wakeLock?: { request(type: "screen"): Promise<WakeLockSentinelLike> } }).wakeLock;
  if (!api) return;

  let sentinel: WakeLockSentinelLike | null = null;
  let wanted = false;

  const acquire = async () => {
    if (!wanted || sentinel || document.visibilityState !== "visible") return;
    try {
      const lock = await api.request("screen");
      sentinel = lock;
      lock.addEventListener("release", () => {
        if (sentinel === lock) sentinel = null;
      });
      if (!wanted) await release(); // the timer stopped while we were asking
    } catch {
      // refused (low battery, no permission): fine, the screen just follows its normal timeout
    }
  };

  const release = async () => {
    const lock = sentinel;
    sentinel = null;
    try {
      await lock?.release();
    } catch {
      // already released
    }
  };

  client.on("state", (s) => {
    wanted = s.timer.phase === "focus" || s.timer.phase === "break";
    if (wanted) void acquire();
    else void release();
  });
  // The browser drops the lock when the tab is hidden; take it again when it is back.
  document.addEventListener("visibilitychange", () => void acquire());
}
