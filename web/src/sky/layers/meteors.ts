import { clamp01, smoothstep } from "../color";
import { createRandom } from "../random";
import type { Layer } from "../types";

const MIN_GAP_S = 7;
const MAX_GAP_S = 18;
const LIFE_S = 1.1;

interface Meteor {
  x: number;
  y: number;
  vx: number;
  vy: number;
  age: number;
}

/** Seconds until the next shooting star. Pure, so it can be tested. */
export function nextGapSeconds(rnd: () => number): number {
  return MIN_GAP_S + rnd() * (MAX_GAP_S - MIN_GAP_S);
}

/** An occasional shooting star on clear nights. */
export function meteorsLayer(seed = 21): Layer {
  const rnd = createRandom(seed);
  let meteor: Meteor | null = null;
  let untilNext = 3 + rnd() * 4; // the first one comes soon, so a new visitor can see it

  return {
    draw({ ctx, w, h, dt, inputs, overcast, horizonY, dpr }) {
      const night = 1 - smoothstep(-14, -6, inputs.sunElevationDeg);
      const clear = clamp01(night * (1 - overcast * 1.2));
      if (clear < 0.2) {
        meteor = null;
        return;
      }

      if (!meteor) {
        untilNext -= dt;
        if (untilNext > 0) return;
        const speed = (0.55 + rnd() * 0.35) * w; // pixels per second, across a good part of the sky
        const angle = (0.35 + rnd() * 0.3) * Math.PI * 0.5; // a gentle downward slant
        meteor = {
          x: w * (0.15 + rnd() * 0.6),
          y: horizonY * (0.04 + rnd() * 0.4),
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          age: 0,
        };
        untilNext = nextGapSeconds(rnd);
      }

      meteor.age += dt;
      meteor.x += meteor.vx * dt;
      meteor.y += meteor.vy * dt;
      const life = meteor.age / LIFE_S;
      if (life >= 1 || meteor.y > horizonY) {
        meteor = null;
        return;
      }

      const fade = Math.sin(life * Math.PI); // in and out
      const tail = 0.16; // seconds of trail
      const tx = meteor.x - meteor.vx * tail;
      const ty = meteor.y - meteor.vy * tail;
      const g = ctx.createLinearGradient(tx, ty, meteor.x, meteor.y);
      g.addColorStop(0, "rgba(255,255,255,0)");
      g.addColorStop(1, `rgba(255,255,255,${0.9 * fade * clear})`);
      ctx.save();
      ctx.strokeStyle = g;
      ctx.lineWidth = 1.6 * dpr;
      ctx.lineCap = "round";
      ctx.beginPath();
      ctx.moveTo(tx, ty);
      ctx.lineTo(meteor.x, meteor.y);
      ctx.stroke();
      ctx.fillStyle = `rgba(255,255,255,${fade * clear})`;
      ctx.beginPath();
      ctx.arc(meteor.x, meteor.y, 1.5 * dpr, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
    },
  };
}
