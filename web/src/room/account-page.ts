import { byId, maybeById } from "../core/dom";
import {
  busiest,
  compareText,
  dayShare,
  defaultDay,
  currentMonth,
  dayLabel,
  formatDuration,
  heatLevel,
  leadingBlanks,
  monthLabel,
  shiftMonth,
  type DayStat,
  type MonthStats,
} from "../stats/format";
import { initCommunity } from "../stats/community";
import { initLeaderboard } from "../stats/leaderboard";
import { initAmbientSky } from "./ambient-sky";
import { t, tn } from "../core/i18n";
import { GOAL_CHOICES_MIN, goalProgress, loadGoal, saveGoal } from "../stats/goal";
import { monthContent, shareCard } from "../ui/share-card";

/** Account page: monthly focus statistics with a calendar heatmap and daily bars. */
export function initAccountPage(): void {
  initAmbientSky();
  void initCommunity();
  initLeaderboard();

  const root = document.querySelector<HTMLElement>(".acct");
  const url = root?.dataset["statsUrl"];
  if (!url) return;

  const nowMonth = currentMonth();
  let month = nowMonth;
  let request = 0;
  let shown: MonthStats | null = null; // the month on screen, for the share picture
  const shareBtn = maybeById<HTMLButtonElement>("share-month");
  shareBtn?.addEventListener("click", () => {
    if (shown) void shareCard(shareBtn, monthContent(shown.totalSeconds, shown.sessions, shown.activeDays, monthLabel(shown.month)), "pomosky-month.png");
  });

  const load = async () => {
    const mine = ++request;
    byId("month-label").textContent = monthLabel(month);
    byId<HTMLButtonElement>("month-next").disabled = month >= nowMonth;
    try {
      const res = await fetch(`${url}?month=${month}&tz=${encodeURIComponent(timeZone())}`, { headers: { Accept: "application/json" } });
      if (!res.ok) throw new Error(String(res.status));
      const data = (await res.json()) as MonthStats;
      if (mine !== request) return; // a newer month was requested meanwhile
      byId("stats-error").hidden = true;
      render(data, month === nowMonth);
      shown = data;
      if (shareBtn) shareBtn.hidden = data.totalSeconds <= 0;
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

function timeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

let selectedDate = "";
let todaySeconds = 0;
let goalMinutes = loadGoal();

function todayKeyNow(): string {
  const d = new Date();
  return `${currentMonth(d)}-${String(d.getDate()).padStart(2, "0")}`;
}

/** The daily goal card: a ring with today's progress and the goal choices. */
function showGoal(seconds: number): void {
  todaySeconds = seconds;
  const share = goalProgress(seconds, goalMinutes);
  const left = goalMinutes * 60 - seconds;
  byId("goal-fill").style.strokeDashoffset = String(100 - share * 100);
  byId("goal-pct").textContent = `${Math.floor(share * 100)}%`;
  byId("goal-line").textContent = `${formatDuration(seconds)} / ${formatDuration(goalMinutes * 60)}`;
  byId("goal-sub").textContent = left > 0 ? t("{time} to go", { time: formatDuration(left) }) : t("Goal reached! Well done.");
  byId("goal-card").classList.toggle("done", left <= 0);
  const box = byId("goal-choices");
  if (box.childElementCount === 0) {
    for (const m of GOAL_CHOICES_MIN) {
      const b = document.createElement("button");
      b.type = "button";
      b.className = "goal-chip";
      b.dataset["min"] = String(m);
      b.textContent = formatDuration(m * 60);
      b.addEventListener("click", () => {
        goalMinutes = m;
        saveGoal(m);
        showGoal(todaySeconds);
      });
      box.appendChild(b);
    }
  }
  for (const b of box.children) b.setAttribute("aria-pressed", String((b as HTMLElement).dataset["min"] === String(goalMinutes)));
}

function render(stats: MonthStats, isCurrent: boolean): void {
  byId("s-total").textContent = formatDuration(stats.totalSeconds);
  byId("s-compare").textContent = compareText(stats.totalSeconds, stats.prevTotalSeconds);
  byId("s-sessions").textContent = String(stats.sessions);
  byId("s-avg").textContent = stats.sessions > 0 ? t("{time} on average", { time: formatDuration(stats.totalSeconds / stats.sessions) }) : "";
  byId("s-days").textContent = `${stats.activeDays} / ${stats.days.length}`;
  const best = busiest(stats.days);
  byId("s-best").textContent = best ? t("Best: {day}, {time}", { day: dayLabel(best.date), time: formatDuration(best.seconds) }) : "";
  byId("s-streak").textContent = String(stats.streakDays);
  if (isCurrent) showGoal(stats.days.find((d) => d.date === todayKeyNow())?.seconds ?? 0);

  const max = Math.max(0, ...stats.days.map((d) => d.seconds));
  const today = new Date();
  const todayKey = `${currentMonth(today)}-${String(today.getDate()).padStart(2, "0")}`;
  selectedDate = defaultDay(stats.days, isCurrent ? todayKey : "");

  const cells = new Map<string, HTMLElement>();
  const bars = new Map<string, HTMLElement>();

  // One place paints the selection: both charts and the day card follow it.
  const select = (d: DayStat) => {
    selectedDate = d.date;
    for (const [date, el] of cells) el.classList.toggle("selected", date === d.date);
    for (const [date, el] of bars) el.classList.toggle("selected", date === d.date);
    showDay(d, stats.totalSeconds);
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
    cell.addEventListener("click", () => select(d));
    cells.set(d.date, cell);
    heat.appendChild(cell);
  }

  // Daily bars.
  const barsEl = byId("bars");
  barsEl.replaceChildren();
  for (const d of stats.days) {
    const bar = document.createElement("button");
    bar.type = "button";
    bar.className = `bar${d.seconds === 0 ? " zero" : ""}`;
    bar.setAttribute("aria-label", `${dayLabel(d.date)}: ${formatDuration(d.seconds)}`);
    const fill = document.createElement("span");
    fill.style.height = max > 0 ? `${Math.max(2, (d.seconds / max) * 100)}%` : "2%";
    bar.appendChild(fill);
    bar.addEventListener("click", () => select(d));
    bar.addEventListener("focus", () => select(d));
    bars.set(d.date, bar);
    barsEl.appendChild(bar);
  }

  const initial = stats.days.find((d) => d.date === selectedDate);
  if (initial) select(initial);
}

function showDay(d: DayStat, monthSeconds: number): void {
  byId("day-title").textContent = dayLabel(d.date);
  byId("day-time").textContent = formatDuration(d.seconds);
  byId("day-sessions").textContent =
    d.seconds > 0 ? tn("{n} session|{n} sessions", d.sessions) : t("No focus time this day");
  byId("day-share").textContent = d.seconds > 0 ? t("{pct}% of the month", { pct: dayShare(d.seconds, monthSeconds) }) : "";
}
