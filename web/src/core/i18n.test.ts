import { describe, expect, it } from "vitest";
import { pluralIndex, t, tn } from "./i18n";

describe("i18n", () => {
  it("returns English text and fills placeholders", () => {
    expect(t("Link copied")).toBe("Link copied");
    expect(t("Hello, {name}", { name: "Ana" })).toBe("Hello, Ana");
    expect(t("Hello, {name}")).toBe("Hello, {name}"); // an unknown placeholder stays visible
  });

  it("picks the plural form", () => {
    expect(tn("{n} session|{n} sessions", 1)).toBe("1 session");
    expect(tn("{n} session|{n} sessions", 0)).toBe("0 sessions");
    expect(tn("{n} session|{n} sessions", 5)).toBe("5 sessions");
  });

  it("knows the Russian plural rules", () => {
    const idx = (n: number) => pluralIndex("ru", n, 3);
    expect([1, 21, 101].map(idx)).toEqual([0, 0, 0]);
    expect([2, 3, 4, 22, 24].map(idx)).toEqual([1, 1, 1, 1, 1]);
    expect([0, 5, 11, 12, 14, 20, 111].map(idx)).toEqual([2, 2, 2, 2, 2, 2, 2]);
    expect(pluralIndex("ko", 7, 1)).toBe(0);
    expect(pluralIndex("uz", 1, 2)).toBe(0);
    expect(pluralIndex("uz", 2, 2)).toBe(1);
    expect(pluralIndex("uz", 2, 1)).toBe(0); // a language that lists one form
  });
});
