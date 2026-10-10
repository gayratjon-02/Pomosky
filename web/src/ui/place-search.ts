import { maybeById } from "../core/dom";

export interface FoundPlace {
  name: string;
  region?: string;
  country: string;
  lat: number;
  lon: number;
  tz: string;
}

const DEBOUNCE_MS = 350;

/** Search box for any city: asks our server (which asks the geocoding service), shows a short list. */
export function initPlaceSearch(onPick: (p: FoundPlace) => void): void {
  const input = maybeById<HTMLInputElement>("place-q");
  const list = maybeById("place-results");
  if (!input || !list) return;

  let timer: number | undefined;
  let request = 0;

  const hide = () => {
    list.hidden = true;
    list.replaceChildren();
  };

  const show = (places: FoundPlace[]) => {
    list.replaceChildren();
    if (places.length === 0) {
      const empty = document.createElement("li");
      empty.className = "place-empty";
      empty.textContent = "No city found";
      list.appendChild(empty);
    }
    for (const p of places) {
      const li = document.createElement("li");
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = p.name;
      const sub = document.createElement("small");
      sub.textContent = [p.region, p.country].filter(Boolean).join(", ");
      b.appendChild(sub);
      b.addEventListener("click", () => {
        onPick(p);
        input.value = "";
        hide();
      });
      li.appendChild(b);
      list.appendChild(li);
    }
    list.hidden = false;
  };

  const search = async (q: string) => {
    const mine = ++request;
    try {
      const res = await fetch(`/api/places?q=${encodeURIComponent(q)}`, { headers: { Accept: "application/json" } });
      if (mine !== request || !res.ok) return;
      show(((await res.json()) as { places: FoundPlace[] }).places);
    } catch {
      // offline: leave the list as it is
    }
  };

  input.addEventListener("input", () => {
    window.clearTimeout(timer);
    const q = input.value.trim();
    if (q.length < 2) return hide();
    timer = window.setTimeout(() => void search(q), DEBOUNCE_MS);
  });
  input.addEventListener("keydown", (e) => {
    if (e.key === "Escape") hide();
  });
  document.addEventListener("click", (e) => {
    if (!(e.target as HTMLElement).closest(".place-search")) hide();
  });
}
