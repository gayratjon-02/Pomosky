// Pure helpers for the monthly statistics view.

export interface DayStat {
  date: string; // YYYY-MM-DD
  seconds: number;
  sessions: number;
}

export interface MonthStats {
  month: string; // YYYY-MM
  totalSeconds: number;
  sessions: number;
  activeDays: number;
  streakDays: number;
  prevTotalSeconds: number;
  days: DayStat[];
}

const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

/** "12h 05m", "45m", "0m". Seconds are rounded down to whole minutes. */
export function formatDuration(seconds: number): string {
  const minutes = Math.floor(Math.max(0, seconds) / 60);
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  return `${h}h ${String(m).padStart(2, "0")}m`;
}

/** Exact time: "3h 20m 15s", "12m 04s", "45s", "0s". */
export function formatExact(seconds: number): string {
  const total = Math.floor(Math.max(0, seconds));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, "0")}m ${String(s).padStart(2, "0")}s`;
  if (m > 0) return `${m}m ${String(s).padStart(2, "0")}s`;
  return `${s}s`;
}

/** "2026-10" -> "October 2026". */
export function monthLabel(month: string): string {
  const [y, m] = month.split("-").map(Number) as [number, number];
  return `${MONTHS[m - 1] ?? "?"} ${y}`;
}

/** Moves a YYYY-MM string by whole months. */
export function shiftMonth(month: string, delta: number): string {
  const [y, m] = month.split("-").map(Number) as [number, number];
  const index = y * 12 + (m - 1) + delta;
  return `${Math.floor(index / 12)}-${String((index % 12) + 1).padStart(2, "0")}`;
}

export function currentMonth(now: Date = new Date()): string {
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

/** Heat level 0..4 of a day relative to the busiest day of the month. */
export function heatLevel(seconds: number, maxSeconds: number): 0 | 1 | 2 | 3 | 4 {
  if (seconds <= 0 || maxSeconds <= 0) return 0;
  const r = seconds / maxSeconds;
  if (r > 0.75) return 4;
  if (r > 0.5) return 3;
  if (r > 0.25) return 2;
  return 1;
}

/** Number of empty cells before the 1st when weeks start on Monday (0..6). */
export function leadingBlanks(month: string): number {
  const [y, m] = month.split("-").map(Number) as [number, number];
  const weekday = new Date(y, m - 1, 1).getDay(); // 0 = Sunday
  return (weekday + 6) % 7;
}

/** Short sentence comparing this month with the previous one. */
export function compareText(current: number, previous: number): string {
  if (previous <= 0) return current > 0 ? "First focused month on record" : "No focus time yet";
  const pct = Math.round(((current - previous) / previous) * 100);
  if (pct === 0) return "Same as last month";
  return `${pct > 0 ? "+" : ""}${pct}% vs last month`;
}

export function busiest(days: DayStat[]): DayStat | null {
  let best: DayStat | null = null;
  for (const d of days) if (d.seconds > 0 && (!best || d.seconds > best.seconds)) best = d;
  return best;
}

/** "Mon 6 Oct". */
export function dayLabel(date: string): string {
  const [y, m, d] = date.split("-").map(Number) as [number, number, number];
  const dt = new Date(y, m - 1, d);
  const wd = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"][dt.getDay()];
  return `${wd} ${d} ${MONTHS[m - 1]?.slice(0, 3)}`;
}

/** Share of the month's focus time that one day makes up, as a whole percent. */
export function dayShare(daySeconds: number, monthSeconds: number): number {
  if (monthSeconds <= 0 || daySeconds <= 0) return 0;
  return Math.round((daySeconds / monthSeconds) * 100);
}

/** The day to select when a month is shown: today in the current month, otherwise the busiest day (or the 1st). */
export function defaultDay(days: DayStat[], todayKey: string): string {
  if (days.some((d) => d.date === todayKey)) return todayKey;
  return busiest(days)?.date ?? days[0]?.date ?? "";
}

/** Big totals read better in hours: "42m", "7h 05m", "1,234 h". */
export function formatTotal(seconds: number): string {
  const hours = Math.floor(Math.max(0, seconds) / 3600);
  if (hours >= 100) return `${hours.toLocaleString("en-US")} h`;
  return formatDuration(seconds);
}

export interface Period {
  seconds: number;
  sessions: number;
}

/** The JSON of GET /api/stats/community (see docs/api.md). */
export interface CommunityStats {
  today: Period;
  week: Period;
  month: Period;
  all: Period;
  days: DayStat[];
}

export type BoardPeriod = "today" | "week" | "month" | "all";

export interface LeaderRow {
  rank: number;
  name: string;
  avatarUrl?: string;
  seconds: number;
  sessions: number;
  you?: boolean;
}

/** First letter of a name for the avatar fallback, "?" when empty. */
export function initialOf(name: string): string {
  const first = [...name.trim()][0];
  return first ? first.toUpperCase() : "?";
}

/** "12.4 MB", "830 KB", "1.2 GB". */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${Math.max(0, Math.round(bytes))} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = bytes / 1024;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i++;
  }
  return `${value >= 100 ? Math.round(value) : value.toFixed(1)} ${units[i]}`;
}

/** "3d 4h", "5h 12m", "42s". */
export function formatUptime(ms: number): string {
  const s = Math.floor(Math.max(0, ms) / 1000);
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${String(m).padStart(2, "0")}m`;
  if (m > 0) return `${m}m ${s % 60}s`;
  return `${s}s`;
}

/** A whole percent, or "-" when there is nothing to divide by. */
export function percent(part: number, whole: number): string {
  return whole > 0 ? `${Math.round((part / whole) * 100)}%` : "-";
}
