import { clamp01, css, smoothstep } from "../color";
import { skyPoint } from "../geometry";
import type { Layer } from "../types";

/**
 * The moon with its real phase. The lit part is a half disc on one side plus or
 * minus half of an ellipse (the terminator). Waxing moons are lit on the right.
 */
export function moonLayer(): Layer {
  return {
    draw(f) {
      const { ctx, w, h, inputs, overcast } = f;
      const e = inputs.moonElevationDeg;
      const visible = smoothstep(-6, 3, e);
      // Faint in daylight, bright at night, hidden behind thick clouds.
      const dim = 1 - 0.85 * smoothstep(-8, 12, inputs.sunElevationDeg);
      const alpha = clamp01(visible * dim * (1 - 0.9 * overcast));
      if (alpha < 0.01) return;

      const { x, y } = skyPoint(f, inputs.moonAzimuthDeg, e);
      const r = Math.min(w, h) * 0.03;
      const waxing = inputs.moonPhase < 0.5;
      const k = clamp01(inputs.moonFraction);

      ctx.save();

      const glow = ctx.createRadialGradient(x, y, r * 0.8, x, y, r * 7);
      glow.addColorStop(0, `rgba(190,205,255,${0.28 * alpha * (0.4 + 0.6 * k)})`);
      glow.addColorStop(1, "rgba(190,205,255,0)");
      ctx.fillStyle = glow;
      ctx.beginPath();
      ctx.arc(x, y, r * 7, 0, Math.PI * 2);
      ctx.fill();

      // The dark side stays faintly visible (earthshine).
      ctx.fillStyle = css([88, 100, 140], 0.5 * alpha);
      ctx.beginPath();
      ctx.arc(x, y, r, 0, Math.PI * 2);
      ctx.fill();

      // Lit part. Mirrored for a waning moon.
      ctx.translate(x, y);
      if (!waxing) ctx.scale(-1, 1);
      const terminator = r * Math.abs(1 - 2 * k);
      ctx.beginPath();
      ctx.arc(0, 0, r, -Math.PI / 2, Math.PI / 2, false); // right half of the limb
      if (k < 0.5) {
        ctx.ellipse(0, 0, terminator, r, 0, Math.PI / 2, -Math.PI / 2, true); // crescent
      } else {
        ctx.ellipse(0, 0, terminator, r, 0, Math.PI / 2, (3 * Math.PI) / 2, false); // gibbous
      }
      ctx.closePath();
      const lit = ctx.createLinearGradient(-r, -r, r, r);
      lit.addColorStop(0, css([250, 250, 255], alpha));
      lit.addColorStop(1, css([214, 222, 244], alpha));
      ctx.fillStyle = lit;
      ctx.fill();

      ctx.restore();
    },
  };
}
