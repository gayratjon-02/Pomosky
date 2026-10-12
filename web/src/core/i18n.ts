import ko from "../../../internal/i18n/locales/ko.json";
import ru from "../../../internal/i18n/locales/ru.json";
import uz from "../../../internal/i18n/locales/uz.json";

// Translations: the key of every text is its English wording, so a missing translation
// shows English. The same JSON files are used by the server templates. Placeholders are
// written {name}. A text that depends on a count is a key with the forms joined by "|":
// t("{n} session|{n} sessions") and the translation lists as many forms as its language needs.

export type Lang = "en" | "uz" | "ru" | "ko";

const DICTS: Record<string, Record<string, string>> = { uz, ru, ko };

/** The language of the page (the server writes it into <html lang>). */
export function currentLang(): Lang {
  const lang = typeof document !== "undefined" ? document.documentElement.lang : "en";
  return lang in DICTS ? (lang as Lang) : "en";
}

/** Locale for Intl formatting of dates and numbers. */
export const intlLocale = (): string => (currentLang() === "en" ? "en-US" : currentLang());

function fill(text: string, vars?: Record<string, string | number>): string {
  if (!vars) return text;
  return text.replace(/\{(\w+)\}/g, (whole, name: string) => (name in vars ? String(vars[name]) : whole));
}

/** Translates an English text and fills its {placeholders}. */
export function t(key: string, vars?: Record<string, string | number>): string {
  const lang = currentLang();
  const text = lang === "en" ? key : (DICTS[lang]?.[key] ?? key);
  return fill(text, vars);
}

// Uzbek month and weekday names are missing from the browsers' Intl data (it prints "M10"), so they are listed here.
const UZ_MONTHS = ["Yanvar", "Fevral", "Mart", "Aprel", "May", "Iyun", "Iyul", "Avgust", "Sentabr", "Oktabr", "Noyabr", "Dekabr"];
const UZ_MONTHS_SHORT = ["Yan", "Fev", "Mar", "Apr", "May", "Iyn", "Iyl", "Avg", "Sen", "Okt", "Noy", "Dek"];
const UZ_WEEKDAYS_SHORT = ["Yak", "Dush", "Sesh", "Chor", "Pay", "Jum", "Shan"]; // Sunday first

/** Name of a month (0 = January) in the language of the page. */
export function monthName(month: number, form: "long" | "short" = "long"): string {
  if (currentLang() === "uz") return (form === "long" ? UZ_MONTHS : UZ_MONTHS_SHORT)[month] ?? "";
  return new Intl.DateTimeFormat(intlLocale(), { month: form }).format(new Date(2024, month, 1));
}

/** Short name of a weekday (0 = Sunday) in the language of the page. */
export function weekdayName(day: number): string {
  if (currentLang() === "uz") return UZ_WEEKDAYS_SHORT[day] ?? "";
  return new Intl.DateTimeFormat(intlLocale(), { weekday: "short" }).format(new Date(2024, 0, 7 + day)); // 7 Jan 2024 was a Sunday
}

/** "12 October 2026" in the language of the page. */
export function longDate(date: Date): string {
  if (currentLang() === "uz") return `${date.getDate()}-${monthName(date.getMonth())} ${date.getFullYear()}`;
  return new Intl.DateTimeFormat(intlLocale(), { day: "numeric", month: "long", year: "numeric" }).format(date);
}

/** Which of the plural forms a language uses for n. */
export function pluralIndex(lang: Lang, n: number, forms: number): number {
  let index: number;
  if (lang === "ru") {
    const mod10 = n % 10;
    const mod100 = n % 100;
    if (mod10 === 1 && mod100 !== 11) index = 0;
    else if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) index = 1;
    else index = 2;
  } else if (lang === "ko") {
    index = 0; // no plural forms
  } else {
    index = n === 1 ? 0 : 1; // English and Uzbek: one, other
  }
  return Math.min(index, forms - 1);
}

/** Translates a text with plural forms; {n} is filled with the count. */
export function tn(key: string, n: number, vars?: Record<string, string | number>): string {
  const lang = currentLang();
  const text = lang === "en" ? key : (DICTS[lang]?.[key] ?? key);
  const forms = text.split("|");
  return fill(forms[pluralIndex(lang, n, forms.length)] ?? text, { n, ...vars });
}
