import { maybeById } from "../core/dom";

export function setViewers(count: number): void {
  const el = maybeById("viewers");
  if (el) el.textContent = String(count);
}
