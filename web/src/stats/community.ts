import { maybeById } from "../core/dom";
import { dayLabel, formatDuration, formatTotal, type CommunityStats } from "./format";
import { t, tn } from "../core/i18n";

function timeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

export async function fetchCommunity(): Promise<CommunityStats | null> {
  try {
    const res = await fetch(`/api/stats/community?tz=${encodeURIComponent(timeZone())}`, { headers: { Accept: "application/json" } });
    return res.ok ? ((await res.json()) as CommunityStats) : null;
  } catch {
    return null;
  }
}

const sessionsText = (n: number): string => tn("{n} session|{n} sessions", n);

/** Fills the "Everyone on Pomosky" section of the statistics page. */
export async function initCommunity(): Promise<void> {
  const section = maybeById("community");
  if (!section) return;
  const stats = await fetchCommunity();
  if (!stats) {
    section.hidden = true;
    return;
  }

  const set = (id: string, text: string) => {
    const el = maybeById(id);
    if (el) el.textContent = text;
  };
  set("c-today", formatTotal(stats.today.seconds));
  set("c-today-n", sessionsText(stats.today.sessions));
  set("c-week", formatTotal(stats.week.seconds));
  set("c-week-n", sessionsText(stats.week.sessions));
  set("c-month", formatTotal(stats.month.seconds));
  set("c-month-n", sessionsText(stats.month.sessions));
  set("c-all", formatTotal(stats.all.seconds));
  set("c-all-n", sessionsText(stats.all.sessions));

  const bars = maybeById("c-bars");
  const detail = maybeById("c-detail");
  if (!bars) return;
  bars.replaceChildren();
  const max = Math.max(0, ...stats.days.map((d) => d.seconds));
  for (const d of stats.days) {
    const bar = document.createElement("button");
    bar.type = "button";
    bar.className = `bar${d.seconds === 0 ? " zero" : ""}`;
    bar.setAttribute("aria-label", `${dayLabel(d.date)}: ${formatDuration(d.seconds)}`);
    const fill = document.createElement("span");
    fill.style.height = max > 0 ? `${Math.max(2, (d.seconds / max) * 100)}%` : "2%";
    bar.appendChild(fill);
    const show = () => {
      if (detail) detail.textContent = t("{day}: {time} in {sessions}", { day: dayLabel(d.date), time: formatDuration(d.seconds), sessions: sessionsText(d.sessions) });
    };
    bar.addEventListener("click", show);
    bar.addEventListener("focus", show);
    bar.addEventListener("mouseenter", show);
    bars.appendChild(bar);
  }
}
