import { css, mix } from "../color";
import { dayAmount, skyGradient } from "../palette";
import type { Layer } from "../types";

/** The sky gradient: three colour stops from the zenith down to the horizon. */
export function backgroundLayer(): Layer {
  return {
    draw({ ctx, w, h, horizonY, inputs, overcast }) {
      const sky = skyGradient(inputs.sunElevationDeg, overcast);
      const horizon = horizonY / h;
      const grad = ctx.createLinearGradient(0, 0, 0, h);
      grad.addColorStop(0, css(sky.top));
      grad.addColorStop(horizon * 0.55, css(sky.mid));
      grad.addColorStop(horizon, css(sky.bottom));
      grad.addColorStop(1, css(sky.bottom));
      ctx.fillStyle = grad;
      ctx.fillRect(0, 0, w, h);

      // By day the air near the horizon is bright and hazy.
      const day = dayAmount(inputs.sunElevationDeg) * (1 - 0.7 * overcast);
      if (day > 0.01) {
        const haze = mix(sky.bottom, [255, 255, 255], 0.55);
        const band = ctx.createLinearGradient(0, horizonY - h * 0.38, 0, horizonY);
        band.addColorStop(0, css(haze, 0));
        band.addColorStop(1, css(haze, 0.55 * day));
        ctx.fillStyle = band;
        ctx.fillRect(0, horizonY - h * 0.38, w, h * 0.38);
      }
    },
  };
}
