import { clamp, clamp01 } from "../color";
import { areaScale } from "../geometry";
import { createRandom } from "../random";
import type { Layer } from "../types";

interface Flake {
  x: number;
  y: number;
  r: number;
  speed: number;
  sway: number;
  phase: number;
  alpha: number;
}

const MAX_FLAKES = 420;

/** Snowflakes that fall slowly, sway and are pushed sideways by the wind. */
export function snowLayer(seed = 9): Layer {
  const rnd = createRandom(seed);
  const flakes: Flake[] = [];
  for (let i = 0; i < MAX_FLAKES; i++) {
    const near = rnd() > 0.6;
    flakes.push({
      x: rnd(),
      y: rnd(),
      r: near ? 1.6 + rnd() * 1.8 : 0.8 + rnd() * 1.1,
      speed: near ? 0.09 + rnd() * 0.05 : 0.05 + rnd() * 0.04,
      sway: 0.01 + rnd() * 0.025,
      phase: rnd() * Math.PI * 2,
      alpha: near ? 0.7 + rnd() * 0.3 : 0.4 + rnd() * 0.3,
    });
  }

  return {
    draw({ ctx, w, h, dt, dpr, timeMs, inputs, quality }) {
      const intensity = clamp01(inputs.snowfallCmh / 2.5);
      if (intensity < 0.02) return;
      const count = Math.round(MAX_FLAKES * clamp01(areaScale(w, h)) * Math.pow(intensity, 0.7) * quality);
      if (count < 1) return;

      const drift = clamp(inputs.windSpeedKmh / 50, 0, 1) * 0.06 * Math.sign(Math.sin((inputs.windDirDeg * Math.PI) / 180) || 1);
      const t = timeMs * 0.001;
      for (let i = 0; i < count; i++) {
        const f = flakes[i]!;
        f.y += f.speed * dt * (0.7 + intensity * 0.6);
        f.x += drift * dt;
        if (f.y > 1.02) {
          f.y -= 1.04;
          f.x = rnd();
        }
        if (f.x > 1.02) f.x -= 1.04;
        if (f.x < -0.02) f.x += 1.04;
        const x = (f.x + Math.sin(t * 0.9 + f.phase) * f.sway) * w;
        const y = f.y * h;
        ctx.fillStyle = `rgba(255,255,255,${f.alpha})`;
        ctx.beginPath();
        ctx.arc(x, y, f.r * dpr, 0, Math.PI * 2);
        ctx.fill();
      }
    },
  };
}
