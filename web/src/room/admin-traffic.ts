import { byId, maybeById } from "../core/dom";
import { dayLabel } from "../stats/format";

interface TrafficDay {
  date: string;
  pageviews: number;
  visitors: number;
  roomsCreated: number;
  logins: number;
}
interface Count {
  label: string;
  visitors: number;
}
interface Traffic {
  days: TrafficDay[];
  today: TrafficDay;
  pages: { path: string; pageviews: number; visitors: number }[];
  sources: { source: string; medium: string; visitors: number; roomsCreated: number; logins: number }[];
  devices: Count[];
  systems: Count[];
  browsers: Count[];
  countries: Count[];
  funnel: { visitors: number; roomViewers: number; roomsCreated: number; loginStarted: number; loginDone: number };
}

const REFRESH_MS = 60_000;
const n = (v: number): string => v.toLocaleString("en-US");
const set = (id: string, text: string) => {
  const el = maybeById(id);
  if (el) el.textContent = text;
};
const sum = (days: TrafficDay[], f: (d: TrafficDay) => number, last = days.length): number =>
  days.slice(-last).reduce((a, d) => a + f(d), 0);

/** One bar per day; tapping a bar shows its numbers under the chart. */
function bars(id: string, detailId: string, days: TrafficDay[], value: (d: TrafficDay) => number, describe: (d: TrafficDay) => string): void {
  const box = byId(id);
  const detail = byId(detailId);
  box.replaceChildren();
  const max = Math.max(0, ...days.map(value));
  for (const d of days) {
    const bar = document.createElement("button");
    bar.type = "button";
    bar.className = `bar${value(d) === 0 ? " zero" : ""}`;
    bar.setAttribute("aria-label", `${dayLabel(d.date)}: ${describe(d)}`);
    const fill = document.createElement("span");
    fill.className = "seg seg-0";
    fill.style.height = max > 0 ? `${(value(d) / max) * 100}%` : "0%";
    bar.appendChild(fill);
    const show = () => {
      detail.textContent = `${dayLabel(d.date)}: ${describe(d)}`;
    };
    bar.addEventListener("click", show);
    bar.addEventListener("focus", show);
    bar.addEventListener("mouseenter", show);
    box.appendChild(bar);
  }
}

interface Row {
  label: string;
  value: number;
  text: string;
  sub?: string;
}

/** A list of rows, each with a bar sized against the biggest value. */
function rows(id: string, items: Row[]): void {
  const box = byId(id);
  box.replaceChildren();
  const max = Math.max(1, ...items.map((i) => i.value));
  if (items.length === 0) {
    const li = document.createElement("li");
    li.className = "tt-empty muted";
    li.textContent = "No data yet.";
    box.appendChild(li);
    return;
  }
  for (const it of items) {
    const li = document.createElement("li");
    li.className = "tt-row";
    const label = document.createElement("span");
    label.className = "tt-label";
    label.textContent = it.label;
    if (it.sub) {
      const sub = document.createElement("small");
      sub.textContent = it.sub;
      label.appendChild(sub);
    }
    const bar = document.createElement("span");
    bar.className = "tt-bar";
    const fill = document.createElement("i");
    fill.style.width = `${Math.max(2, (it.value / max) * 100)}%`;
    bar.appendChild(fill);
    const val = document.createElement("strong");
    val.textContent = it.text;
    li.append(label, bar, val);
    box.appendChild(li);
  }
}

const pct = (part: number, whole: number): string => (whole > 0 ? `${Math.round((part / whole) * 100)}%` : "-");

function render(t: Traffic): void {
  const d = t.days;
  set("t-visitors", n(t.today.visitors));
  set("t-visitors-sub", `today · ${n(sum(d, (x) => x.visitors, 7))} in 7 days · ${n(sum(d, (x) => x.visitors))} in 30`);
  set("t-views", n(t.today.pageviews));
  set("t-views-sub", `today · ${n(sum(d, (x) => x.pageviews, 7))} in 7 days · ${n(sum(d, (x) => x.pageviews))} in 30`);
  set("t-rooms", n(t.today.roomsCreated));
  set("t-rooms-sub", `today · ${n(sum(d, (x) => x.roomsCreated, 7))} in 7 days · ${n(sum(d, (x) => x.roomsCreated))} in 30`);
  set("t-conv", pct(t.funnel.roomsCreated, t.funnel.visitors));
  set("t-conv-sub", "of visitors created a room (30 days)");

  bars("t-chart-visitors", "t-visitors-detail", d, (x) => x.visitors, (x) => `${x.visitors} visitors, ${x.pageviews} page views`);
  bars("t-chart-rooms", "t-rooms-detail", d, (x) => x.roomsCreated, (x) => `${x.roomsCreated} rooms created, ${x.logins} logins`);

  const f = t.funnel;
  rows("t-funnel", [
    { label: "Visited the site", value: f.visitors, text: n(f.visitors) },
    { label: "Opened a room", value: f.roomViewers, text: `${n(f.roomViewers)} · ${pct(f.roomViewers, f.visitors)}` },
    { label: "Created a room", value: f.roomsCreated, text: `${n(f.roomsCreated)} · ${pct(f.roomsCreated, f.visitors)}` },
    { label: "Started a login", value: f.loginStarted, text: `${n(f.loginStarted)} · ${pct(f.loginStarted, f.visitors)}` },
    { label: "Finished a login", value: f.loginDone, text: `${n(f.loginDone)} · ${pct(f.loginDone, f.loginStarted)} of starts` },
  ]);

  rows(
    "t-sources",
    t.sources.map((s) => ({
      label: s.source,
      sub: s.medium || undefined,
      value: s.visitors,
      text: `${n(s.visitors)} · ${s.roomsCreated} room${s.roomsCreated === 1 ? "" : "s"} · ${s.logins} login${s.logins === 1 ? "" : "s"}`,
    })),
  );
  rows("t-pages", t.pages.map((p) => ({ label: p.path, value: p.pageviews, text: `${n(p.pageviews)} views` })));
  const total = (c: Count[]) => c.reduce((a, x) => a + x.visitors, 0);
  const share = (c: Count[]) => c.map((x) => ({ label: x.label, value: x.visitors, text: `${n(x.visitors)} · ${pct(x.visitors, total(c))}` }));
  rows("t-devices", share(t.devices));
  rows("t-systems", share(t.systems));
  rows("t-browsers", share(t.browsers));
  rows("t-countries", share(t.countries));
}

/** The admin traffic section: anonymous visitors, sources, funnel and devices of the last 30 days. */
export function initTraffic(): void {
  const load = async () => {
    try {
      const res = await fetch("/api/admin/traffic", { headers: { Accept: "application/json" }, cache: "no-store" });
      if (!res.ok) return;
      render((await res.json()) as Traffic);
    } catch {
      // keep what is shown; the next refresh tries again
    }
  };
  void load();
  window.setInterval(() => {
    if (!document.hidden) void load();
  }, REFRESH_MS);
  maybeById("adm-refresh")?.addEventListener("click", () => void load());
}
