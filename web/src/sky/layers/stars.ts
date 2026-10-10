import { clamp01, smoothstep } from "../color";
import { areaScale } from "../geometry";
import { createRandom } from "../random";
import type { Layer } from "../types";

interface Star {
  x: number;
  y: number;
  r: number;
  a: number;
  tint: string;
  twinkle: boolean;
  phase: number;
  speed: number;
}

const TINTS = ["255,255,255", "205,225,255", "255,240,215"];

/**
 * A star field that is drawn into an offscreen canvas once per resize; only the
 * few brightest stars are redrawn every frame so they can twinkle.
 */
export function starsLayer(seed = 7): Layer {
  let stars: Star[] = [];
  let still: HTMLCanvasElement | null = null;
  let lastQuality = 0;
  let size = { w: 0, h: 0, dpr: 1 };

  const build = (quality: number) => {
    const rnd = createRandom(seed);
    const count = Math.round(260 * areaScale(size.w, size.h) * quality);
    stars = [];
    for (let i = 0; i < count; i++) {
      const bright = rnd() > 0.93;
      stars.push({
        x: rnd() * size.w,
        y: rnd() * size.h * 0.8,
        r: (bright ? 1.1 + rnd() * 0.9 : 0.4 + rnd() * 0.7) * size.dpr,
        a: bright ? 0.85 + rnd() * 0.15 : 0.35 + rnd() * 0.5,
        tint: TINTS[Math.floor(rnd() * TINTS.length)]!,
        twinkle: bright,
        phase: rnd() * Math.PI * 2,
        speed: 0.8 + rnd() * 1.6,
      });
    }
    still = document.createElement("canvas");
    still.width = size.w;
    still.height = size.h;
    const sctx = still.getContext("2d")!;
    for (const s of stars) {
      if (s.twinkle) continue;
      sctx.fillStyle = `rgba(${s.tint},${s.a})`;
      sctx.beginPath();
      sctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
      sctx.fill();
    }
    lastQuality = quality;
  };

  return {
    resize(w, h, dpr) {
      size = { w, h, dpr };
      still = null;
    },
    draw({ ctx, w, h, timeMs, inputs, overcast, quality, horizonY }) {
      const night = 1 - smoothstep(-16, -4, inputs.sunElevationDeg);
      const alpha = clamp01(night * (1 - 0.92 * overcast));
      if (alpha < 0.01) return;
      if (!still || Math.abs(quality - lastQuality) > 0.2) build(quality);

      ctx.save();
      ctx.beginPath();
      ctx.rect(0, 0, w, horizonY);
      ctx.clip();
      ctx.globalAlpha = alpha;
      ctx.drawImage(still!, 0, 0, w, h);
      for (const s of stars) {
        if (!s.twinkle) continue;
        const flicker = 0.65 + 0.35 * Math.sin(timeMs * 0.001 * s.speed + s.phase);
        ctx.fillStyle = `rgba(${s.tint},${s.a * flicker})`;
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
        ctx.fill();
      }
      ctx.restore();
    },
  };
}
