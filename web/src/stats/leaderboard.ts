import { maybeById } from "../core/dom";
import { avatarElement } from "../ui/avatar";
import { formatTotal, type BoardPeriod, type LeaderRow } from "./format";
import { t, tn } from "../core/i18n";

const PERIODS: BoardPeriod[] = ["today", "week", "month", "all"];

function timeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

async function fetchBoard(period: BoardPeriod): Promise<LeaderRow[] | null> {
  try {
    const res = await fetch(`/api/stats/leaderboard?period=${period}&tz=${encodeURIComponent(timeZone())}`, { headers: { Accept: "application/json" } });
    if (!res.ok) return null;
    return ((await res.json()) as { entries: LeaderRow[] }).entries;
  } catch {
    return null;
  }
}

function renderRows(list: HTMLElement, rows: LeaderRow[]): void {
  list.replaceChildren();
  const empty = maybeById("board-empty");
  if (empty) empty.hidden = rows.length > 0;
  rows.forEach((row, i) => {
    const li = document.createElement("li");
    li.className = `board-row${row.you ? " is-you" : ""}${row.rank <= 3 ? ` top top-${row.rank}` : ""}`;
    li.style.setProperty("--i", String(Math.min(i, 12)));

    const rank = document.createElement("span");
    rank.className = "board-rank";
    rank.textContent = String(row.rank);

    const who = document.createElement("span");
    who.className = "board-who";
    const name = document.createElement("strong");
    name.textContent = row.name;
    who.appendChild(name);
    if (row.you) {
      const you = document.createElement("span");
      you.className = "board-you";
      you.textContent = t("You");
      who.appendChild(you);
    }

    const time = document.createElement("span");
    time.className = "board-time";
    const total = document.createElement("strong");
    total.textContent = formatTotal(row.seconds);
    const sessions = document.createElement("small");
    sessions.textContent = tn("{n} session|{n} sessions", row.sessions);
    time.append(total, sessions);

    li.append(rank, avatarElement(row.name, row.avatarUrl), who, time);
    list.appendChild(li);
  });
}

/** The leaderboard of the statistics page: tabs for today, the week, the month and all time. */
export function initLeaderboard(): void {
  const section = maybeById("leaderboard");
  const list = maybeById("board-list");
  const tabs = Array.from(document.querySelectorAll<HTMLButtonElement>("#board-tabs button[data-period]"));
  if (!section || !list || tabs.length === 0) return;

  let request = 0;
  const show = async (period: BoardPeriod) => {
    const mine = ++request;
    tabs.forEach((t) => {
      const on = t.dataset["period"] === period;
      t.setAttribute("aria-selected", String(on));
      t.tabIndex = on ? 0 : -1;
    });
    list.setAttribute("aria-busy", "true");
    const rows = await fetchBoard(period);
    if (mine !== request) return; // a newer tab was chosen meanwhile
    list.removeAttribute("aria-busy");
    if (!rows) {
      list.replaceChildren();
      const empty = maybeById("board-empty");
      if (empty) {
        empty.hidden = false;
        empty.textContent = t("The leaderboard is not available right now.");
      }
      return;
    }
    const empty = maybeById("board-empty");
    if (empty) empty.textContent = t("Nobody here yet. Host a room while logged in and be the first.");
    renderRows(list, rows);
  };

  tabs.forEach((t, i) => {
    t.addEventListener("click", () => void show(t.dataset["period"] as BoardPeriod));
    t.addEventListener("keydown", (e) => {
      if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
      const next = tabs[(i + (e.key === "ArrowRight" ? 1 : tabs.length - 1)) % tabs.length]!;
      next.focus();
      void show(next.dataset["period"] as BoardPeriod);
    });
  });
  void show(PERIODS[0]!);
}
