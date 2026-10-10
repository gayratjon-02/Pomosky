import { clamp, clamp01 } from "../color";
import { areaScale } from "../geometry";
import { createRandom } from "../random";
import type { Layer } from "../types";

interface Drop {
  x: number;
  y: number;
  len: number;
  speed: number;
  alpha: number;
}

const MAX_DROPS = 700;

/**
 * Rain streaks. The number of drops follows the precipitation rate, the slant
 * follows the wind. Two depths (small slow drops, long fast ones) give some parallax.
 */
export function rainLayer(seed = 5): Layer {
  const rnd = createRandom(seed);
  const drops: Drop[] = [];
  for (let i = 0; i < MAX_DROPS; i++) {
    const near = rnd() > 0.55;
    drops.push({
      x: rnd(),
      y: rnd(),
      len: near ? 0.03 + rnd() * 0.03 : 0.012 + rnd() * 0.016,
      speed: near ? 1.5 + rnd() * 0.7 : 0.8 + rnd() * 0.5,
      alpha: near ? 0.35 + rnd() * 0.25 : 0.15 + rnd() * 0.2,
    });
  }

  return {
    draw({ ctx, w, h, dt, dpr, inputs, quality }) {
      const intensity = clamp01(inputs.precipitationMmh / 8); // 8 mm/h is a heavy shower
      if (intensity < 0.02) return;
      const count = Math.round(MAX_DROPS * clamp01(areaScale(w, h)) * Math.pow(intensity, 0.7) * quality);
      if (count < 1) return;

      // Slant from the wind: a drop travels a little sideways, away from where the wind comes from.
      const slant = clamp(inputs.windSpeedKmh / 60, 0, 1) * 0.45 * Math.sign(Math.sin((inputs.windDirDeg * Math.PI) / 180) || 1);
      const fall = h * (0.9 + intensity * 0.5); // pixels per second at speed 1

      ctx.lineCap = "round";
      ctx.lineWidth = Math.max(1, dpr * 1.1);
      for (let i = 0; i < count; i++) {
        const d = drops[i]!;
        d.y += (d.speed * fall * dt) / h;
        d.x += (d.speed * fall * dt * slant) / w;
        if (d.y > 1.05) {
          d.y -= 1.1;
          d.x = (d.x + 1.3) % 1.2;
        }
        const x = d.x * w;
        const y = d.y * h;
        const dy = d.len * h;
        ctx.strokeStyle = `rgba(205,222,255,${d.alpha * (0.5 + 0.5 * intensity)})`;
        ctx.beginPath();
        ctx.moveTo(x, y);
        ctx.lineTo(x - dy * slant, y - dy);
        ctx.stroke();
      }
    },
  };
}
