import { clamp01, css, smoothstep } from "../color";
import { skyPoint } from "../geometry";
import { sunColor } from "../palette";
import type { Layer } from "../types";

/** Sun disc with a soft bloom. It grows, reddens and fades near the horizon and is veiled by clouds. */
export function sunLayer(): Layer {
  return {
    draw(f) {
      const { ctx, w, h, inputs, overcast } = f;
      const e = inputs.sunElevationDeg;
      const visible = smoothstep(-8, 1, e);
      if (visible < 0.01) return;

      const { x, y } = skyPoint(f, inputs.sunAzimuthDeg, e);
      const base = Math.min(w, h) * 0.04;
      const r = base * (1 + 0.4 * (1 - smoothstep(0, 25, e)));
      const color = sunColor(e);
      const clear = clamp01(1 - overcast * 0.85);

      ctx.save();
      ctx.globalCompositeOperation = "screen";

      const bloomR = r * 16;
      const bloom = ctx.createRadialGradient(x, y, 0, x, y, bloomR);
      bloom.addColorStop(0, css(color, 0.55 * visible * clear));
      bloom.addColorStop(0.25, css(color, 0.22 * visible * clear));
      bloom.addColorStop(1, css(color, 0));
      ctx.fillStyle = bloom;
      ctx.fillRect(x - bloomR, y - bloomR, bloomR * 2, bloomR * 2);

      const halo = ctx.createRadialGradient(x, y, r * 0.6, x, y, r * 3.2);
      halo.addColorStop(0, css(color, 0.8 * visible * clear));
      halo.addColorStop(1, css(color, 0));
      ctx.fillStyle = halo;
      ctx.beginPath();
      ctx.arc(x, y, r * 3.2, 0, Math.PI * 2);
      ctx.fill();

      // Soft rays fanning out from the sun by day, turning very slowly.
      const rays = smoothstep(4, 22, e) * clear;
      if (rays > 0.01) {
        const reach = Math.max(w, h) * 0.7;
        const spin = f.timeMs * 0.00002;
        const count = 14;
        for (let i = 0; i < count; i++) {
          const a = spin + (i / count) * Math.PI * 2 + Math.sin(i * 12.9898) * 0.2;
          const half = 0.025 + 0.025 * Math.abs(Math.sin(i * 4.1));
          const g = ctx.createRadialGradient(x, y, r, x, y, reach);
          g.addColorStop(0, css(color, 0.05 * rays));
          g.addColorStop(0.6, css(color, 0.015 * rays));
          g.addColorStop(1, css(color, 0));
          ctx.fillStyle = g;
          ctx.beginPath();
          ctx.moveTo(x, y);
          ctx.arc(x, y, reach, a - half, a + half);
          ctx.closePath();
          ctx.fill();
        }
      }

      const disc = ctx.createRadialGradient(x, y, 0, x, y, r);
      disc.addColorStop(0, css([255, 255, 255], visible * clear));
      disc.addColorStop(0.7, css(color, visible * clear));
      disc.addColorStop(1, css(color, 0.9 * visible * clear));
      ctx.fillStyle = disc;
      ctx.beginPath();
      ctx.arc(x, y, r, 0, Math.PI * 2);
      ctx.fill();

      ctx.restore();
    },
  };
}
