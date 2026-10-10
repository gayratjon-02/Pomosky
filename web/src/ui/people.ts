import { maybeById } from "../core/dom";
import type { PeopleData } from "../core/protocol";
import { avatarElement } from "./avatar";

/** Shows the named people of the room as a row of small pictures with a "+N" for the rest. */
export function setPeople(data: PeopleData): void {
  const box = maybeById("people");
  if (!box) return;
  box.hidden = data.people.length === 0;
  box.replaceChildren();
  for (const p of data.people) {
    const a = avatarElement(p.name, p.avatarUrl);
    a.classList.add("people-item");
    a.title = p.name;
    box.appendChild(a);
  }
  if (data.more > 0) {
    const more = document.createElement("span");
    more.className = "people-more";
    more.textContent = `+${data.more}`;
    box.appendChild(more);
  }
  const names = data.people.map((p) => p.name).join(", ");
  box.setAttribute("aria-label", names ? `Working now: ${names}${data.more > 0 ? ` and ${data.more} more` : ""}` : "");
}
