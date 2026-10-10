import { byId, maybeById } from "../core/dom";
import { dayLabel, formatBytes, formatDuration, formatTotal, formatUptime, initialOf, percent } from "../stats/format";
import { initAmbientSky } from "./ambient-sky";
import { initTraffic } from "./admin-traffic";

interface AdminDay {
  date: string;
  newUsers: number;
  registeredSeconds: number;
  registeredSessions: number;
  anonymousSeconds: number;
  anonymousSessions: number;
  activeHosts: number;
}

interface Overview {
  live: { rooms: number; clients: number; focus: number; break: number; paused: number; idle: number };
  totals: {
    users: number;
    rooms: number;
    registeredSeconds: number;
    registeredSessions: number;
    anonymousSeconds: number;
    anonymousSessions: number;
    hiddenUsers: number;
  };
  newUsers: { today: number; week: number; month: number };
  activeHosts: { today: number; week: number; month: number };
  days: AdminDay[];
  system: {
    startedAtMs: number;
    goVersion: string;
    goroutines: number;
    heapBytes: number;
    databaseBytes: number;
    weatherCalls: number;
    weatherHits: number;
    weatherErrors: number;
    weatherBusy: number;
  };
}

interface UserRow {
  email: string;
  provider: string;
  name: string;
  avatarUrl?: string;
  joinedAtMs: number;
  showOnLeaderboard: boolean;
  focusSeconds: number;
  sessions: number;
  lastFocusAtMs?: number;
}

const REFRESH_MS = 10_000;

function timeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

const set = (id: string, text: string) => {
  const el = maybeById(id);
  if (el) el.textContent = text;
};

const dateText = (ms: number): string => new Date(ms).toLocaleDateString(undefined, { day: "numeric", month: "short", year: "numeric" });

async function getJSON<T>(url: string): Promise<T | null> {
  try {
    const res = await fetch(url, { headers: { Accept: "application/json" }, cache: "no-store" });
    return res.ok ? ((await res.json()) as T) : null;
  } catch {
    return null;
  }
}

function renderBars(
  id: string,
  detailId: string,
  days: AdminDay[],
  parts: (d: AdminDay) => number[],
  describe: (d: AdminDay) => string,
): void {
  const box = byId(id);
  const detail = byId(detailId);
  box.replaceChildren();
  const totals = days.map((d) => parts(d).reduce((a, b) => a + b, 0));
  const max = Math.max(0, ...totals);
  days.forEach((d, i) => {
    const bar = document.createElement("button");
    bar.type = "button";
    bar.className = `bar${totals[i] === 0 ? " zero" : ""}`;
    bar.setAttribute("aria-label", `${dayLabel(d.date)}: ${describe(d)}`);
    parts(d).forEach((value, k) => {
      const fill = document.createElement("span");
      fill.className = `seg seg-${k}`;
      fill.style.height = max > 0 ? `${(value / max) * 100}%` : "0%";
      bar.appendChild(fill);
    });
    const show = () => {
      detail.textContent = `${dayLabel(d.date)}: ${describe(d)}`;
    };
    bar.addEventListener("click", show);
    bar.addEventListener("focus", show);
    bar.addEventListener("mouseenter", show);
    box.appendChild(bar);
  });
}

function renderOverview(o: Overview): void {
  set("l-rooms", String(o.live.rooms));
  set("l-clients", String(o.live.clients));
  set("l-focus", String(o.live.focus));
  set("l-other", `${o.live.break} break · ${o.live.paused} paused · ${o.live.idle} idle`);

  const t = o.totals;
  const allSeconds = t.registeredSeconds + t.anonymousSeconds;
  const allSessions = t.registeredSessions + t.anonymousSessions;
  set("o-users", t.users.toLocaleString("en-US"));
  set("o-users-sub", `${t.hiddenUsers} hidden from the leaderboard`);
  set("o-new-today", String(o.newUsers.today));
  set("o-new-sub", `today · ${o.newUsers.week} in 7 days · ${o.newUsers.month} in 30`);
  set("o-hosts-today", String(o.activeHosts.today));
  set("o-hosts-sub", `today · ${o.activeHosts.week} in 7 days · ${o.activeHosts.month} in 30`);
  set("o-focus", formatTotal(allSeconds));
  set("o-focus-sub", `${formatTotal(t.registeredSeconds)} logged in · ${formatTotal(t.anonymousSeconds)} anonymous`);
  set("o-sessions", allSessions.toLocaleString("en-US"));
  set("o-sessions-sub", `${t.registeredSessions.toLocaleString("en-US")} logged in · ${t.anonymousSessions.toLocaleString("en-US")} anonymous`);
  set("o-anon", percent(t.anonymousSeconds, allSeconds));

  renderBars(
    "a-focus",
    "a-focus-detail",
    o.days,
    (d) => [d.registeredSeconds, d.anonymousSeconds],
    (d) => `${formatDuration(d.registeredSeconds)} logged in, ${formatDuration(d.anonymousSeconds)} anonymous (${d.registeredSessions + d.anonymousSessions} sessions)`,
  );
  renderBars("a-users", "a-users-detail", o.days, (d) => [d.newUsers], (d) => `${d.newUsers} new user${d.newUsers === 1 ? "" : "s"}`);
  renderBars("a-hosts", "a-hosts-detail", o.days, (d) => [d.activeHosts], (d) => `${d.activeHosts} active host${d.activeHosts === 1 ? "" : "s"}`);

  const s = o.system;
  const lookups = s.weatherCalls + s.weatherHits;
  const rows: [string, string][] = [
    ["Running for", formatUptime(Date.now() - s.startedAtMs)],
    ["Started", new Date(s.startedAtMs).toLocaleString()],
    ["Go", s.goVersion],
    ["Goroutines", String(s.goroutines)],
    ["Memory in use", formatBytes(s.heapBytes)],
    ["Database size", s.databaseBytes > 0 ? formatBytes(s.databaseBytes) : "-"],
    ["Rooms stored", t.rooms.toLocaleString("en-US")],
    ["Weather lookups", lookups.toLocaleString("en-US")],
    ["Weather cache hits", percent(s.weatherHits, lookups)],
    ["Weather upstream calls", s.weatherCalls.toLocaleString("en-US")],
    ["Weather errors", s.weatherErrors.toLocaleString("en-US")],
    ["Weather refused (budget)", s.weatherBusy.toLocaleString("en-US")],
  ];
  const dl = byId("adm-sys");
  dl.replaceChildren();
  for (const [k, v] of rows) {
    const wrap = document.createElement("div");
    const dt = document.createElement("dt");
    dt.textContent = k;
    const dd = document.createElement("dd");
    dd.textContent = v;
    wrap.append(dt, dd);
    dl.appendChild(wrap);
  }
}

function userRow(u: UserRow): HTMLElement {
  const li = document.createElement("li");
  li.className = "user-row";

  const av = document.createElement("span");
  av.className = "avatar avatar-sm";
  if (u.avatarUrl) {
    const img = document.createElement("img");
    img.className = "avatar-img";
    img.src = u.avatarUrl;
    img.alt = "";
    img.width = 64;
    img.height = 64;
    img.referrerPolicy = "no-referrer";
    img.loading = "lazy";
    av.appendChild(img);
  } else {
    const i = document.createElement("span");
    i.className = "avatar-initial";
    i.textContent = initialOf(u.name);
    av.appendChild(i);
  }

  const who = document.createElement("span");
  who.className = "user-who";
  const name = document.createElement("strong");
  name.textContent = u.name;
  const mail = document.createElement("small");
  mail.textContent = `${u.email || "no email"} · ${u.provider}`;
  who.append(name, mail);

  const focus = document.createElement("span");
  focus.className = "user-focus";
  const total = document.createElement("strong");
  total.textContent = formatTotal(u.focusSeconds);
  const sessions = document.createElement("small");
  sessions.textContent = `${u.sessions} session${u.sessions === 1 ? "" : "s"}`;
  focus.append(total, sessions);

  const meta = document.createElement("span");
  meta.className = "user-meta";
  const joined = document.createElement("small");
  joined.textContent = `Joined ${dateText(u.joinedAtMs)}`;
  const last = document.createElement("small");
  last.textContent = u.lastFocusAtMs ? `Last focus ${dateText(u.lastFocusAtMs)}` : "Never focused";
  const board = document.createElement("small");
  board.textContent = u.showOnLeaderboard ? "On the leaderboard" : "Hidden from leaderboard";
  meta.append(joined, last, board);

  li.append(av, who, focus, meta);
  return li;
}

/** The operator's admin page: live numbers, totals, 30-day charts, users and system health. */
export function initAdminPage(): void {
  initAmbientSky();
  initTraffic();

  let sort: "newest" | "focus" = "newest";
  let page = 1;
  let pageSize = 25;
  let total = 0;

  const loadOverview = async () => {
    const o = await getJSON<Overview>(`/api/admin/overview?tz=${encodeURIComponent(timeZone())}`);
    if (!o) {
      set("adm-updated", "failed, retrying");
      return;
    }
    renderOverview(o);
    set("adm-updated", new Date().toLocaleTimeString());
  };

  const loadUsers = async () => {
    const data = await getJSON<{ users: UserRow[]; total: number; pageSize: number }>(`/api/admin/users?sort=${sort}&page=${page}`);
    if (!data) return;
    total = data.total;
    pageSize = data.pageSize;
    const list = byId("us-list");
    list.replaceChildren(...data.users.map(userRow));
    byId("us-empty").hidden = data.users.length > 0;
    set("us-total", `(${total.toLocaleString("en-US")})`);
    const pages = Math.max(1, Math.ceil(total / pageSize));
    set("us-page", `Page ${page} of ${pages}`);
    byId<HTMLButtonElement>("us-prev").disabled = page <= 1;
    byId<HTMLButtonElement>("us-next").disabled = page >= pages;
  };

  byId("us-prev").addEventListener("click", () => {
    page = Math.max(1, page - 1);
    void loadUsers();
  });
  byId("us-next").addEventListener("click", () => {
    page += 1;
    void loadUsers();
  });
  const tabs = Array.from(document.querySelectorAll<HTMLButtonElement>("#us-sort button[data-sort]"));
  tabs.forEach((t) =>
    t.addEventListener("click", () => {
      sort = t.dataset["sort"] === "focus" ? "focus" : "newest";
      page = 1;
      tabs.forEach((x) => {
        x.setAttribute("aria-selected", String(x === t));
        x.tabIndex = x === t ? 0 : -1;
      });
      void loadUsers();
    }),
  );
  byId("adm-refresh").addEventListener("click", () => {
    void loadOverview();
    void loadUsers();
  });

  void loadOverview();
  void loadUsers();
  // Only the live numbers need to stay fresh; pause while the tab is hidden.
  window.setInterval(() => {
    if (!document.hidden) void loadOverview();
  }, REFRESH_MS);
}
