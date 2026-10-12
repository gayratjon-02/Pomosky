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
