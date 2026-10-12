import { maybeById } from "../core/dom";
import { tn } from "../core/i18n";

export function setViewers(count: number): void {
  const el = maybeById("viewers");
  if (el) el.textContent = String(count);
  const label = maybeById("viewers-label");
  if (label) label.textContent = tn("person here|people here", count);
}
