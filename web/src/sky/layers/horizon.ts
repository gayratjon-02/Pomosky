import { css, mix, hex } from "../color";
import { horizonColor, skyGradient } from "../palette";
import { createRandom } from "../random";
import type { Layer } from "../types";

interface Ridge {
  base: number; // baseline as a fraction of the horizon offset
  amp: number; // height of the hills as a fraction of the canvas height
  waves: Array<{ freq: number; phase: number; weight: number }>;
}

function makeRidge(rnd: () => number, base: number, amp: number): Ridge {
  return {
    base,
    amp,
    waves: [
      { freq: 1.1 + rnd() * 0.8, phase: rnd() * 6.28, weight: 1 },
      { freq: 2.7 + rnd() * 1.4, phase: rnd() * 6.28, weight: 0.55 },
      { freq: 6.2 + rnd() * 3, phase: rnd() * 6.28, weight: 0.22 },
    ],
  };
}

const waveSum = (r: Ridge, u: number): number => {
  let total = 0;
  let weight = 0;
  for (const w of r.waves) {
    total += Math.sin(u * Math.PI * 2 * w.freq + w.phase) * w.weight;
    weight += w.weight;
  }
  return 0.5 + (0.5 * total) / weight; // 0..1
};

/** Layered hills along the horizon. Their colour follows the sky and gets darker towards the front. */
export function horizonLayer(seed = 3): Layer {
  const rnd = createRandom(seed);
  const ridges = [makeRidge(rnd, 0.0, 0.07), makeRidge(rnd, 0.02, 0.06), makeRidge(rnd, 0.045, 0.05)];

  return {
    draw({ ctx, w, h, horizonY, inputs, overcast }) {
      const e = inputs.sunElevationDeg;
      const step = Math.max(6, Math.round(w / 140));
      ridges.forEach((ridge, i) => {
        const depth = i / (ridges.length - 1);
        const color = horizonColor(e, depth);
        // Hills pick up a little of the sky's colour, less so in the front.
        const tinted = mix(color, skyGradient(e, overcast).bottom, 0.22 * (1 - depth));
        const baseY = horizonY + ridge.base * h;

        ctx.beginPath();
        ctx.moveTo(0, h);
        for (let x = 0; x <= w + step; x += step) {
          const y = baseY - ridge.amp * h * waveSum(ridge, x / w);
          ctx.lineTo(x, y);
        }
        ctx.lineTo(w, h);
        ctx.closePath();

        const top = baseY - ridge.amp * h;
        const g = ctx.createLinearGradient(0, top, 0, h);
        g.addColorStop(0, css(tinted));
        g.addColorStop(1, css(mix(tinted, hex("#02030a"), 0.55)));
        ctx.fillStyle = g;
        ctx.fill();
      });
    },
  };
}
