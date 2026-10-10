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
