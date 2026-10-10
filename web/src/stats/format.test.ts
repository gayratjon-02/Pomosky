import { describe, expect, it } from "vitest";
import { initialOf, formatTotal, busiest, dayShare, defaultDay, compareText, formatDuration, heatLevel, leadingBlanks, monthLabel, shiftMonth } from "./format";

describe("formatDuration", () => {
  it("formats minutes and hours", () => {
    expect(formatDuration(0)).toBe("0m");
    expect(formatDuration(59)).toBe("0m");
    expect(formatDuration(45 * 60)).toBe("45m");
    expect(formatDuration(3600 + 5 * 60)).toBe("1h 05m");
    expect(formatDuration(-5)).toBe("0m");
  });
});

describe("months", () => {
  it("labels and shifts across years", () => {
    expect(monthLabel("2026-10")).toBe("October 2026");
    expect(shiftMonth("2026-01", -1)).toBe("2025-12");
    expect(shiftMonth("2026-12", 1)).toBe("2027-01");
    expect(shiftMonth("2026-10", 0)).toBe("2026-10");
  });

  it("counts blanks with Monday first", () => {
    expect(leadingBlanks("2026-10")).toBe(3); // 1 Oct 2026 is a Thursday
    expect(leadingBlanks("2026-06")).toBe(0); // 1 Jun 2026 is a Monday
    expect(leadingBlanks("2026-02")).toBe(6); // 1 Feb 2026 is a Sunday
  });
});

describe("heatLevel", () => {
  it("scales against the busiest day", () => {
    expect(heatLevel(0, 100)).toBe(0);
    expect(heatLevel(10, 100)).toBe(1);
    expect(heatLevel(40, 100)).toBe(2);
    expect(heatLevel(60, 100)).toBe(3);
    expect(heatLevel(100, 100)).toBe(4);
    expect(heatLevel(5, 0)).toBe(0);
  });
});

describe("compareText and busiest", () => {
  it("describes the change", () => {
    expect(compareText(120, 100)).toBe("+20% vs last month");
    expect(compareText(50, 100)).toBe("-50% vs last month");
    expect(compareText(100, 100)).toBe("Same as last month");
    expect(compareText(10, 0)).toBe("First focused month on record");
    expect(compareText(0, 0)).toBe("No focus time yet");
  });

  it("finds the busiest day", () => {
    expect(busiest([])).toBeNull();
    const days = [
      { date: "2026-10-01", seconds: 0, sessions: 0 },
      { date: "2026-10-02", seconds: 600, sessions: 1 },
      { date: "2026-10-03", seconds: 900, sessions: 2 },
    ];
    expect(busiest(days)?.date).toBe("2026-10-03");
  });
});

describe("dayShare and defaultDay", () => {
  it("computes the share of the month", () => {
    expect(dayShare(300, 1200)).toBe(25);
    expect(dayShare(0, 1200)).toBe(0);
    expect(dayShare(5, 0)).toBe(0);
  });

  it("selects today, else the busiest day, else the first", () => {
    const days = [
      { date: "2026-09-01", seconds: 0, sessions: 0 },
      { date: "2026-09-02", seconds: 900, sessions: 1 },
      { date: "2026-09-03", seconds: 300, sessions: 1 },
    ];
    expect(defaultDay(days, "2026-09-03")).toBe("2026-09-03");
    expect(defaultDay(days, "2026-10-10")).toBe("2026-09-02");
    expect(defaultDay([{ date: "2026-09-01", seconds: 0, sessions: 0 }], "2026-10-10")).toBe("2026-09-01");
    expect(defaultDay([], "2026-10-10")).toBe("");
  });
});

describe("formatTotal", () => {
  it("switches to plain hours for big numbers", () => {
    expect(formatTotal(0)).toBe("0m");
    expect(formatTotal(42 * 60)).toBe("42m");
    expect(formatTotal(7 * 3600 + 5 * 60)).toBe("7h 05m");
    expect(formatTotal(1234 * 3600 + 59)).toBe("1,234 h");
    expect(formatTotal(100 * 3600)).toBe("100 h");
  });
});

describe("initialOf", () => {
  it("takes the first letter, capitalised", () => {
    expect(initialOf("ayla")).toBe("A");
    expect(initialOf("  Tik Tok")).toBe("T");
    expect(initialOf("élan")).toBe("É");
    expect(initialOf("")).toBe("?");
    expect(initialOf("😀 fun")).toBe("😀");
  });
});
