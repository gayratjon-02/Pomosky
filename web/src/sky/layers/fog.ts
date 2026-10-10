import { clamp01, css, mix, hex } from "../color";
import { skyGradient } from "../palette";
import type { Layer } from "../types";

/** Fog: a pale band that thickens towards the horizon, plus a light veil over everything. */
export function fogLayer(): Layer {
  return {
    draw({ ctx, w, h, horizonY, inputs, overcast }) {
      const density = clamp01(inputs.fogDensity);
      if (density < 0.01) return;

      const sky = skyGradient(inputs.sunElevationDeg, Math.max(overcast, 0.6));
      const fog = mix(sky.bottom, hex("#cfd6e4"), 0.55);

      const top = horizonY - h * 0.5 * density;
      const band = ctx.createLinearGradient(0, top, 0, h);
      band.addColorStop(0, css(fog, 0));
      band.addColorStop(0.55, css(fog, 0.78 * density));
      band.addColorStop(1, css(fog, 0.92 * density));
      ctx.fillStyle = band;
      ctx.fillRect(0, top, w, h - top);

      ctx.fillStyle = css(fog, 0.22 * density);
      ctx.fillRect(0, 0, w, h);
    },
  };
}
