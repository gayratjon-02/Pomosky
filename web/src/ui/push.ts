// Web Push: the server sends an alert when a phase ends, also while the page is
// closed or the phone is locked. The page only subscribes and keeps the server's
// memory of it fresh (the server forgets subscriptions when a room closes or restarts).

let active = false;

/** True while this browser holds a working push subscription for the room. */
export const pushActive = (): boolean => active;

export function pushSupported(): boolean {
  return "serviceWorker" in navigator && "PushManager" in window && typeof Notification !== "undefined";
}

const csrf = (): string => document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')?.content ?? "";

async function post(path: string, body: unknown): Promise<boolean> {
  try {
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf() },
      body: JSON.stringify(body),
    });
    return res.ok;
  } catch {
    return false;
  }
}

function keyBytes(base64url: string): Uint8Array<ArrayBuffer> {
  const pad = "=".repeat((4 - (base64url.length % 4)) % 4);
  const raw = atob((base64url + pad).replace(/-/g, "+").replace(/_/g, "/"));
  const out = new Uint8Array(new ArrayBuffer(raw.length));
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

async function serverKey(): Promise<string | null> {
  try {
    const res = await fetch("/api/push/key", { headers: { Accept: "application/json" } });
    if (!res.ok) return null; // 404: the server has no push keys
    return ((await res.json()) as { publicKey?: string }).publicKey ?? null;
  } catch {
    return null;
  }
}

/** Subscribes (or refreshes the subscription) for a room. Needs granted notification permission. */
export async function subscribePush(roomId: string): Promise<boolean> {
  if (!pushSupported() || Notification.permission !== "granted") return false;
  try {
    const reg = await navigator.serviceWorker.register("/sw.js");
    await navigator.serviceWorker.ready;
    let sub = await reg.pushManager.getSubscription();
    if (!sub) {
      const key = await serverKey();
      if (!key) return false;
      sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: keyBytes(key) });
    }
    active = await post("/api/push/subscribe", { room: roomId, subscription: sub.toJSON() });
    return active;
  } catch {
    active = false;
    return false;
  }
}

/** Removes the subscription of this browser. */
export async function unsubscribePush(): Promise<void> {
  active = false;
  if (!pushSupported()) return;
  try {
    const reg = await navigator.serviceWorker.getRegistration("/sw.js");
    const sub = await reg?.pushManager.getSubscription();
    if (!sub) return;
    await post("/api/push/unsubscribe", { endpoint: sub.endpoint });
    await sub.unsubscribe();
  } catch {
    // nothing to clean up
  }
}
