import { byId, maybeById } from "../core/dom";
import {
  busiest,
  compareText,
  currentMonth,
  dayLabel,
  formatDuration,
  heatLevel,
  leadingBlanks,
  monthLabel,
  shiftMonth,
  type MonthStats,
} from "../stats/format";
import { initAmbientSky } from "./ambient-sky";

/** Account page: monthly focus statistics with a calendar heatmap and daily bars. */
export function initAccountPage(): void {
  initAmbientSky();

  const root = document.querySelector<HTMLElement>(".acct");
  const url = root?.dataset["statsUrl"];
  if (!url) return;

  const nowMonth = currentMonth();
  let month = nowMonth;
  let request = 0;

  const load = async () => {
    const mine = ++request;
    byId("month-label").textContent = monthLabel(month);
    byId<HTMLButtonElement>("month-next").disabled = month >= nowMonth;
    try {
      const res = await fetch(`${url}?month=${month}`, { headers: { Accept: "application/json" } });
      if (!res.ok) throw new Error(String(res.status));
      const data = (await res.json()) as MonthStats;
      if (mine !== request) return; // a newer month was requested meanwhile
      byId("stats-error").hidden = true;
      render(data, month === nowMonth);
    } catch {
      if (mine === request) byId("stats-error").hidden = false;
    }
  };

  byId("month-prev").addEventListener("click", () => {
    month = shiftMonth(month, -1);
    void load();
  });
  byId("month-next").addEventListener("click", () => {
    if (month >= nowMonth) return;
    month = shiftMonth(month, 1);
    void load();
  });

  void load();
}

function render(stats: MonthStats, isCurrent: boolean): void {
  byId("s-total").textContent = formatDuration(stats.totalSeconds);
  byId("s-compare").textContent = compareText(stats.totalSeconds, stats.prevTotalSeconds);
  byId("s-sessions").textContent = String(stats.sessions);
  byId("s-avg").textContent = stats.sessions > 0 ? `${formatDuration(stats.totalSeconds / stats.sessions)} on average` : "";
  byId("s-days").textContent = `${stats.activeDays} / ${stats.days.length}`;
  const best = busiest(stats.days);
  byId("s-best").textContent = best ? `Best: ${dayLabel(best.date)}, ${formatDuration(best.seconds)}` : "";
  byId("s-streak").textContent = String(stats.streakDays);

  const max = Math.max(0, ...stats.days.map((d) => d.seconds));
  const today = new Date();
  const todayKey = `${currentMonth(today)}-${String(today.getDate()).padStart(2, "0")}`;
  const detail = byId("day-detail");
  const show = (d: MonthStats["days"][number]) => {
    detail.textContent =
      d.seconds > 0
        ? `${dayLabel(d.date)}: ${formatDuration(d.seconds)} in ${d.sessions} session${d.sessions === 1 ? "" : "s"}`
        : `${dayLabel(d.date)}: no focus time`;
  };

  // Calendar heatmap, weeks start on Monday.
  const heat = byId("heatmap");
  heat.replaceChildren();
  for (let i = 0; i < leadingBlanks(stats.month); i++) {
    const blank = document.createElement("span");
    blank.className = "heat blank";
    heat.appendChild(blank);
  }
  for (const d of stats.days) {
    const cell = document.createElement("button");
    cell.type = "button";
    cell.setAttribute("role", "listitem");
    const future = isCurrent && d.date > todayKey;
    cell.className = `heat h${heatLevel(d.seconds, max)}${future ? " future" : ""}${d.date === todayKey ? " today" : ""}`;
    cell.textContent = String(Number(d.date.slice(8)));
    cell.setAttribute("aria-label", `${dayLabel(d.date)}: ${formatDuration(d.seconds)}`);
    cell.addEventListener("click", () => show(d));
    cell.addEventListener("mouseenter", () => show(d));
    heat.appendChild(cell);
  }

  // Daily bars.
  const bars = byId("bars");
  bars.replaceChildren();
  for (const d of stats.days) {
    const bar = document.createElement("button");
    bar.type = "button";
    bar.className = `bar${d.seconds === 0 ? " zero" : ""}`;
    bar.setAttribute("aria-label", `${dayLabel(d.date)}: ${formatDuration(d.seconds)}`);
    const fill = document.createElement("span");
    fill.style.height = max > 0 ? `${Math.max(2, (d.seconds / max) * 100)}%` : "2%";
    bar.appendChild(fill);
    bar.addEventListener("mouseenter", () => show(d));
    bar.addEventListener("focus", () => show(d));
    bar.addEventListener("click", () => show(d));
    bars.appendChild(bar);
  }
  maybeById("day-detail")!.textContent = "Hover or tap a day to see its focus time.";
}
