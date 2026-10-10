import { css } from "./color";
import { skyGradient } from "./palette";
import { onVisibilityChange, prefersReducedMotion } from "./perf";
import type { SkyController, SkyInputs } from "./types";

const defaults: SkyInputs = {
  sunElevationDeg: 30,
  sunAzimuthDeg: 180,
  moonPhase: 0,
  cloudCoverPct: 0,
  precipitationMmh: 0,
  snowfallCmh: 0,
  windSpeedKmh: 0,
  windDirDeg: 0,
  condition: "clear",
};

/**
 * Creates the sky renderer for a canvas. The animation pauses while the tab is
 * hidden and is drawn once (no loop) when the user prefers reduced motion.
 */
export function createSky(canvas: HTMLCanvasElement): SkyController {
  const ctx = canvas.getContext("2d");
  let inputs: SkyInputs = { ...defaults };
  let raf = 0;
  let running = false;

  const resize = () => {
    const dpr = window.devicePixelRatio || 1;
    canvas.width = Math.floor(canvas.clientWidth * dpr);
    canvas.height = Math.floor(canvas.clientHeight * dpr);
  };

  const draw = (_timeMs: number) => {
    if (!ctx) return;
    const { top, bottom } = skyGradient(inputs.sunElevationDeg);
    const g = ctx.createLinearGradient(0, 0, 0, canvas.height);
    g.addColorStop(0, css(top));
    g.addColorStop(1, css(bottom));
    ctx.fillStyle = g;
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    // TODO(stage 6): stars, sun/moon, clouds, rain, snow, fog, lightning.
  };

  const loop = (t: number) => {
    if (!running) return;
    draw(t);
    raf = requestAnimationFrame(loop);
  };

  onVisibilityChange((visible) => {
    if (!visible) controller.stop();
    else controller.start();
  });

  const controller: SkyController = {
    start() {
      if (running) return;
      resize();
      running = true;
      if (prefersReducedMotion()) {
        draw(0);
        running = false;
        return;
      }
      raf = requestAnimationFrame(loop);
    },
    stop() {
      running = false;
      cancelAnimationFrame(raf);
    },
    update(next) {
      inputs = { ...inputs, ...next };
      if (!running) draw(0);
    },
  };

  window.addEventListener("resize", resize);
  return controller;
}
