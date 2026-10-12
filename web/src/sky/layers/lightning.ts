import { createRandom } from "../random";
import { publishBolt } from "../sky-feed";
import type { Layer } from "../types";

interface Bolt {
  points: Array<[number, number]>;
  branches: Array<Array<[number, number]>>;
}

/** Flash profile over the first 450 ms (soft on purpose: harsh flicker is unpleasant and looks bad on camera). */
const FLASH_MS = 450;
const flashAlpha = (ms: number): number => {
  if (ms < 60) return (ms / 60) * 0.3;
  if (ms < 130) return 0.3 - ((ms - 60) / 70) * 0.22;
  if (ms < 200) return 0.08 + ((ms - 130) / 70) * 0.2;
  return Math.max(0, 0.28 * (1 - (ms - 200) / (FLASH_MS - 200)));
};

function makeBolt(rnd: () => number): Bolt {
  const points: Array<[number, number]> = [];
  const branches: Array<Array<[number, number]>> = [];
  let x = 0.2 + rnd() * 0.6;
  let y = 0.02;
  points.push([x, y]);
  while (y < 0.74) {
    y += 0.03 + rnd() * 0.05;
    x += (rnd() - 0.5) * 0.07;
    points.push([x, y]);
    if (rnd() > 0.78) {
      const branch: Array<[number, number]> = [[x, y]];
      let bx = x;
      let by = y;
      for (let i = 0; i < 3 + Math.floor(rnd() * 3); i++) {
        by += 0.025 + rnd() * 0.03;
        bx += (rnd() - 0.2) * 0.06;
        branch.push([bx, by]);
      }
      branches.push(branch);
    }
  }
  return { points, branches };
}

/** Occasional lightning during storms: a soft flash of the whole sky and a short-lived bolt. */
export function lightningLayer(seed = 13): Layer {
  const rnd = createRandom(seed);
  let nextAt = 0;
  let startedAt = -1;
  let bolt: Bolt | null = null;

  return {
    draw({ ctx, w, h, timeMs, inputs }) {
      if (!inputs.thunder) {
        startedAt = -1;
        nextAt = 0;
        return;
      }
      if (nextAt === 0) nextAt = timeMs + 1500 + rnd() * 4000;
      if (startedAt < 0 && timeMs >= nextAt) {
        startedAt = timeMs;
        bolt = makeBolt(rnd);
        publishBolt(); // the ambient sound answers with thunder
      }
      if (startedAt < 0) return;

      const age = timeMs - startedAt;
      if (age > FLASH_MS) {
        startedAt = -1;
        nextAt = timeMs + 4000 + rnd() * 9000;
        return;
      }

      const a = flashAlpha(age);
      ctx.fillStyle = `rgba(215,225,255,${a})`;
      ctx.fillRect(0, 0, w, h);

      if (bolt && age < 220) {
        const boltAlpha = 1 - age / 220;
        ctx.lineJoin = "round";
        ctx.lineCap = "round";
        const draw = (pts: Array<[number, number]>, width: number) => {
          ctx.beginPath();
          pts.forEach(([px, py], i) => (i === 0 ? ctx.moveTo(px * w, py * h) : ctx.lineTo(px * w, py * h)));
          ctx.strokeStyle = `rgba(235,240,255,${boltAlpha})`;
          ctx.lineWidth = width;
          ctx.shadowColor = "rgba(160,185,255,0.9)";
          ctx.shadowBlur = 18;
          ctx.stroke();
        };
        draw(bolt.points, Math.max(1.5, w / 700));
        for (const b of bolt.branches) draw(b, Math.max(1, w / 1100));
        ctx.shadowBlur = 0;
      }
    },
  };
}
