// Pomosky service worker: shows the "phase ended" alert sent by the server
// (Web Push), also when no Pomosky page is open or the phone is locked.

self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (event) => event.waitUntil(self.clients.claim()));

self.addEventListener("push", (event) => {
  let msg = {};
  try {
    msg = event.data ? event.data.json() : {};
  } catch {
    msg = {};
  }
  const title = msg.title || "Pomosky";
  event.waitUntil(
    (async () => {
      // The open page already rings and shows the phase change itself.
      const pages = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
      if (pages.some((p) => p.visibilityState === "visible" && p.focused)) return;
      await self.registration.showNotification(title, {
        body: msg.body || "",
        icon: "/static/icon-192.png",
        badge: "/static/icon-192.png",
        tag: "pomosky-end",
        renotify: true,
        data: { room: msg.room || "" },
      });
    })(),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const room = (event.notification.data && event.notification.data.room) || "";
  const path = room ? `/r/${room}` : "/";
  event.waitUntil(
    (async () => {
      const pages = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
      const same = pages.find((p) => new URL(p.url).pathname.startsWith(path));
      if (same) return same.focus();
      if (pages.length > 0 && "navigate" in pages[0]) {
        const moved = await pages[0].navigate(path);
        return (moved || pages[0]).focus();
      }
      return self.clients.openWindow(path);
    })(),
  );
});
