import { maybeById } from "../core/dom";

export function setViewers(count: number): void {
  const el = maybeById("viewers");
  if (el) el.textContent = String(count);
  const label = maybeById("viewers-label");
  if (label) label.textContent = count === 1 ? "person here" : "people here";
}
