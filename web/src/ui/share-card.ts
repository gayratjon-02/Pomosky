import { longDate, t, tn } from "../core/i18n";
import { createSky } from "../sky";
import { latestSky } from "../sky/sky-feed";
import { DEFAULT_INPUTS } from "../sky/types";
import { formatDuration, formatTotal } from "../stats/format";
import { toast } from "./toast";

// A picture of what you did, for stories and chats: the current sky (live or the chosen
// mood) behind a big number. It is drawn in the browser; nothing is sent anywhere.

export interface CardContent {
  big: string; // "25", "1h 25m"
  line: string; // "minutes of focus"
  label?: string; // what the person worked on
  sub?: string; // a smaller line, e.g. "5 sessions · 12 active days"
}

/** The big number and its caption for a focus session of the given length. */
export function sessionContent(durationMs: number, label: string): CardContent {
  const minutes = Math.max(1, Math.round(durationMs / 60_000));
  const trimmed = label.trim();
  const base: CardContent = trimmed ? { big: "", line: "", label: trimmed } : { big: "", line: "" };
  if (minutes >= 60) return { ...base, big: formatDuration(minutes * 60), line: t("of focus") };
  return { ...base, big: String(minutes), line: tn("minute of focus|minutes of focus", minutes) };
}

/** The card of a whole month. */
export function monthContent(totalSeconds: number, sessions: number, activeDays: number, monthLabel: string): CardContent {
  return {
    big: formatTotal(totalSeconds),
    line: t("focused in {month}", { month: monthLabel }),
    sub: `${tn("{n} session|{n} sessions", sessions)} · ${tn("{n} active day|{n} active days", activeDays)}`,
  };
}

const W = 1080;
const H = 1920;
const FONT = `ui-rounded, "SF Pro Rounded", "SF Pro Display", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif`;

/** Breaks text into at most `maxLines` lines that fit `maxWidth`; the last line gets an ellipsis if cut. */
function wrap(ctx: CanvasRenderingContext2D, text: string, maxWidth: number, maxLines: number): string[] {
  const words = text.split(/\s+/);
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    const next = line ? `${line} ${word}` : word;
    if (ctx.measureText(next).width <= maxWidth || !line) line = next;
    else {
      lines.push(line);
      line = word;
    }
  }
  if (line) lines.push(line);
  if (lines.length > maxLines) {
    const kept = lines.slice(0, maxLines);
    kept[maxLines - 1] = `${kept[maxLines - 1]!.replace(/[\s.,;:]+$/, "")}…`;
    return kept;
  }
  return lines;
}

/** Draws the sky the visitor currently sees onto a fresh canvas of the card size. */
function drawSky(): HTMLCanvasElement {
  const host = document.createElement("canvas");
  host.style.cssText = `position:fixed;left:-99999px;top:0;width:${W}px;height:${H}px;pointer-events:none`;
  host.setAttribute("aria-hidden", "true");
  document.body.appendChild(host);
  const sky = createSky(host, { still: true, maxDpr: 1 });
  sky.update(latestSky() ?? DEFAULT_INPUTS);
  const out = document.createElement("canvas");
  out.width = W;
  out.height = H;
  out.getContext("2d")!.drawImage(host, 0, 0, W, H);
  sky.destroy();
  host.remove();
  return out;
}

/** Renders the card as a PNG (1080 x 1920, the size of a story). */
export async function renderCard(content: CardContent): Promise<Blob> {
  const canvas = drawSky();
  const ctx = canvas.getContext("2d")!;
  ctx.textAlign = "center";
  ctx.textBaseline = "alphabetic";

  // A soft dark veil keeps white text readable on any sky, brighter or darker.
  const veil = ctx.createLinearGradient(0, 0, 0, H);
  veil.addColorStop(0, "rgba(6,14,40,0.38)");
  veil.addColorStop(0.45, "rgba(6,14,40,0.18)");
  veil.addColorStop(0.62, "rgba(6,14,40,0.34)");
  veil.addColorStop(1, "rgba(6,14,40,0.62)");
  ctx.fillStyle = veil;
  ctx.fillRect(0, 0, W, H);

  const text = (value: string, y: number, size: number, weight: number, alpha = 1) => {
    ctx.font = `${weight} ${size}px ${FONT}`;
    ctx.fillStyle = `rgba(255,255,255,${alpha})`;
    ctx.shadowColor = "rgba(6,14,40,0.45)";
    ctx.shadowBlur = size * 0.18;
    ctx.shadowOffsetY = size * 0.04;
    ctx.fillText(value, W / 2, y);
    ctx.shadowColor = "transparent";
  };

  // The brand, small, at the top.
  const mark = ctx.createLinearGradient(0, 150, 0, 230);
  mark.addColorStop(0, "#6a9bff");
  mark.addColorStop(1, "#ffb36b");
  ctx.fillStyle = mark;
  ctx.beginPath();
  ctx.arc(W / 2 - 150, 190, 38, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = "rgba(255,246,227,0.95)";
  ctx.beginPath();
  ctx.arc(W / 2 - 150, 202, 15, 0, Math.PI * 2);
  ctx.fill();
  ctx.textAlign = "left";
  ctx.font = `700 72px ${FONT}`;
  ctx.fillStyle = "#fff";
  ctx.shadowColor = "rgba(6,14,40,0.4)";
  ctx.shadowBlur = 12;
  ctx.fillText("Pomosky", W / 2 - 92, 214);
  ctx.shadowColor = "transparent";
  ctx.textAlign = "center";

  // The big number, shrunk until it fits the card ("1 ч 25 мин" is much wider than "25").
  ctx.font = `800 520px ${FONT}`;
  const bigSize = Math.min(520, Math.floor((520 * 900) / Math.max(ctx.measureText(content.big).width, 1)));
  text(content.big, 1000, bigSize, 800);
  text(content.line, 1000 + 120, 74, 600, 0.95);

  let y = 1000 + 120;
  if (content.label) {
    ctx.font = `600 60px ${FONT}`;
    const lines = wrap(ctx, `“${content.label}”`, 860, 2);
    y += 110;
    for (const line of lines) {
      text(line, y, 60, 600, 0.92);
      y += 76;
    }
  }
  if (content.sub) text(content.sub, y + 56, 54, 600, 0.9);

  text(t("Focus together under a living sky"), 1680, 52, 600, 0.92);
  text("pomosky.pro", 1770, 66, 800);
  text(longDate(new Date()), 1840, 40, 500, 0.7);

  return await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((b) => (b ? resolve(b) : reject(new Error("could not make the image"))), "image/png");
  });
}

/** Hands the picture to the phone's share sheet, or saves it where there is none. */
export async function shareImage(blob: Blob, filename: string): Promise<void> {
  const file = new File([blob], filename, { type: "image/png" });
  const caption = `${t("Focus together under a living sky")}\n${location.origin}`;
  if (navigator.canShare?.({ files: [file] })) {
    try {
      await navigator.share({ files: [file], text: caption });
      return;
    } catch (e) {
      if ((e as DOMException).name === "AbortError") return; // closed without sharing
    }
  }
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 10_000);
  toast(t("Image saved"));
}

/** Makes and shares a card, keeping the button from being pressed twice meanwhile. */
export async function shareCard(button: HTMLButtonElement | null, content: CardContent, filename: string): Promise<void> {
  if (button) button.disabled = true;
  try {
    await shareImage(await renderCard(content), filename);
  } catch {
    toast(t("Something went wrong"));
  } finally {
    if (button) button.disabled = false;
  }
}
