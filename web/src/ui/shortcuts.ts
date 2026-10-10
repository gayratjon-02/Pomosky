import type { Controls } from "./host-controls";

export type ShortcutAction = "toggle-pause" | "start-focus" | "start-break" | "reset";

export interface KeyInfo {
  key: string;
  shiftKey: boolean;
  ctrlKey: boolean;
  metaKey: boolean;
  altKey: boolean;
  /** true when the event came from a text field, a select or an editable element */
  typing: boolean;
}

/**
 * The host action a key press means, or null. Space or K pauses and resumes, F starts a
 * focus phase, B a break, Shift+R resets. A key only counts when the matching button is
 * visible, and never while typing or with Ctrl, Cmd or Alt held.
 */
export function shortcutFor(k: KeyInfo, c: Controls): ShortcutAction | null {
  if (k.typing || k.ctrlKey || k.metaKey || k.altKey) return null;
  const key = k.key.toLowerCase();
  if ((k.key === " " || key === "k") && !k.shiftKey) return c.showPause ? "toggle-pause" : null;
  if (key === "f" && !k.shiftKey) return c.showStart ? "start-focus" : null;
  if (key === "b" && !k.shiftKey) return c.showStart ? "start-break" : null;
  if (key === "r" && k.shiftKey) return c.showReset ? "reset" : null; // Shift needed, so it is never an accident
  return null;
}
