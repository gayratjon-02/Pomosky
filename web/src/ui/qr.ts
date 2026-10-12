import qrcode from "qrcode-generator";
import { t } from "../core/i18n";

/** Draws the QR code of a link as an inline SVG. Generated locally: no data leaves the page. */
export function renderQr(target: HTMLElement, url: string): void {
  const qr = qrcode(0, "M");
  qr.addData(url);
  qr.make();
  const n = qr.getModuleCount();
  const quiet = 2;
  const size = n + quiet * 2;
  let path = "";
  for (let y = 0; y < n; y++) {
    for (let x = 0; x < n; x++) {
      if (qr.isDark(y, x)) path += `M${x + quiet} ${y + quiet}h1v1h-1z`;
    }
  }
  target.innerHTML =
    `<svg viewBox="0 0 ${size} ${size}" role="img" aria-label="${t("QR code of the room link")}" shape-rendering="crispEdges">` +
    `<rect width="${size}" height="${size}" fill="#fff"/><path d="${path}" fill="#0b1020"/></svg>`;
}
