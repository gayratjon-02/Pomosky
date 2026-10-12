// Daily focus goal: kept in this browser only, so it needs no account or server.

export const GOAL_CHOICES_MIN = [30, 60, 120, 180, 240] as const;
const DEFAULT_GOAL_MIN = 120;
const KEY = "pomosky_goal";

/** The saved goal in minutes; anything unexpected falls back to the default. */
export function loadGoal(): number {
  try {
    const n = Number(localStorage.getItem(KEY));
    if ((GOAL_CHOICES_MIN as readonly number[]).includes(n)) return n;
  } catch {
    /* storage can be blocked */
  }
  return DEFAULT_GOAL_MIN;
}

export function saveGoal(minutes: number): void {
  try {
    localStorage.setItem(KEY, String(minutes));
  } catch {
    /* storage can be blocked */
  }
}

/** Share of the goal done, from 0 to 1. */
export function goalProgress(seconds: number, goalMinutes: number): number {
  if (goalMinutes <= 0) return 0;
  return Math.min(1, Math.max(0, seconds) / (goalMinutes * 60));
}
