import { css } from "../color";
import { skyGradient } from "../palette";
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
    },
  };
}
