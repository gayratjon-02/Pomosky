import { smoothstep } from "../color";
import type { Layer } from "../types";

/** A soft dark edge that keeps text readable over any sky; a bit stronger at night. */
export function vignetteLayer(): Layer {
  return {
    draw({ ctx, w, h, inputs }) {
      const night = 1 - smoothstep(-8, 10, inputs.sunElevationDeg);
      const strength = 0.28 + 0.14 * night;
      const g = ctx.createRadialGradient(w / 2, h * 0.45, Math.min(w, h) * 0.3, w / 2, h * 0.5, Math.max(w, h) * 0.75);
      g.addColorStop(0, "rgba(0,0,0,0)");
      g.addColorStop(1, `rgba(2,4,14,${strength})`);
      ctx.fillStyle = g;
      ctx.fillRect(0, 0, w, h);
    },
  };
}
