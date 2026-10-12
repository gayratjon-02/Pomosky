import { SILENT, type Scene } from "./ambience-scene";

// The sound of a place, made live with the Web Audio API (no audio files, nothing to
// license): noise shaped into wind and rain, and short synthesised voices for birds,
// crickets, drops, bells, an owl and thunder. A Scene says how much of each is wanted.
// The engine works on any audio context, so it can also be rendered offline (tests).

const PENTATONIC = [784, 987.8, 1174.7, 1318.5, 1568];
const OWL_HZ = 392;

type Rand = () => number;
type Kind = "birds" | "crickets" | "drips" | "chimes" | "owl";

export class AmbienceEngine {
  private readonly bus: GainNode;
  private readonly noise: AudioBuffer;
  private readonly level = new Map<keyof Scene, GainNode>();
  private readonly next: Record<Kind, number> = { birds: 0, crickets: 0, drips: 0, chimes: 0, owl: 0 };
  private scene: Scene = SILENT;
  private stopNodes: Array<AudioScheduledSourceNode> = [];

  constructor(
    private readonly ctx: BaseAudioContext,
    out: AudioNode,
    private readonly rand: Rand = Math.random,
  ) {
    const squash = ctx.createDynamicsCompressor(); // keeps a busy storm from clipping
    squash.threshold.value = -16;
    squash.ratio.value = 4;
    this.bus = ctx.createGain();
    this.bus.gain.value = 1.6;
    this.bus.connect(squash).connect(out);
    this.noise = this.makeNoise(7);
    this.buildBeds();
  }

  private makeNoise(seconds: number): AudioBuffer {
    const buf = this.ctx.createBuffer(1, Math.floor(this.ctx.sampleRate * seconds), this.ctx.sampleRate);
    const data = buf.getChannelData(0);
    for (let i = 0; i < data.length; i++) data[i] = this.rand() * 2 - 1;
    return buf;
  }

  private loopNoise(): AudioBufferSourceNode {
    const src = this.ctx.createBufferSource();
    src.buffer = this.noise;
    src.loop = true;
    src.start(0, this.rand() * 5);
    this.stopNodes.push(src);
    return src;
  }

  private lfo(hz: number, depth: number, target: AudioParam): void {
    const osc = this.ctx.createOscillator();
    const g = this.ctx.createGain();
    osc.frequency.value = hz;
    g.gain.value = depth;
    osc.connect(g).connect(target);
    osc.start();
    this.stopNodes.push(osc);
  }

  /** A gain stage whose value follows one layer of the scene. */
  private layer(name: keyof Scene): GainNode {
    const g = this.ctx.createGain();
    g.gain.value = 0;
    g.connect(this.bus);
    this.level.set(name, g);
    return g;
  }

  private filter(type: BiquadFilterType, freq: number, q = 0.7): BiquadFilterNode {
    const f = this.ctx.createBiquadFilter();
    f.type = type;
    f.frequency.value = freq;
    f.Q.value = q;
    return f;
  }

  /** The beds that run all the time and only change in loudness. */
  private buildBeds(): void {
    const c = this.ctx;

    // Wind: band-passed noise whose pitch and strength drift slowly.
    const wind = this.layer("wind");
    const windBand = this.filter("bandpass", 480, 0.7);
    const windSwell = c.createGain();
    windSwell.gain.value = 1;
    this.lfo(0.05, 220, windBand.frequency);
    this.lfo(0.11, 0.35, windSwell.gain);
    this.loopNoise().connect(windBand).connect(windSwell).connect(wind);

    // Rain: a bright hiss for the drops on leaves and a darker body for the roof of the sky.
    const rain = this.layer("rain");
    this.loopNoise().connect(this.filter("highpass", 900)).connect(this.filter("lowpass", 9000)).connect(rain);
    const body = c.createGain();
    body.gain.value = 0.5;
    this.loopNoise().connect(this.filter("bandpass", 420, 0.4)).connect(body).connect(rain);

    // Snowy hush: very soft, muffled air.
    const hush = this.layer("hush");
    const hushSwell = c.createGain();
    hushSwell.gain.value = 1;
    this.lfo(0.07, 0.25, hushSwell.gain);
    this.loopNoise().connect(this.filter("lowpass", 300)).connect(hushSwell).connect(hush);

    // Night drone: a deep open fifth, barely there.
    const drone = this.layer("drone");
    const droneTone = this.filter("lowpass", 320);
    droneTone.connect(drone);
    for (const [hz, amp] of [
      [130.8, 1],
      [196, 0.7],
      [261.6, 0.4],
    ] as const) {
      const osc = c.createOscillator();
      const g = c.createGain();
      osc.frequency.value = hz;
      g.gain.value = amp;
      this.lfo(0.04 + hz / 4000, 1.5, osc.detune);
      osc.connect(g).connect(droneTone);
      osc.start();
      this.stopNodes.push(osc);
    }

    // Distant rumble under a storm.
    const rumble = this.layer("thunder");
    this.loopNoise().connect(this.filter("lowpass", 90)).connect(rumble);
    this.layer("birds");
    this.layer("crickets");
    this.layer("drips");
    this.layer("chimes");
    this.layer("owl");
  }

  private static readonly BED_GAIN: Partial<Record<keyof Scene, number>> = {
    wind: 0.3,
    rain: 0.06,
    hush: 0.22,
    drone: 0.03,
    thunder: 0.12,
  };

  /** Moves the beds to a scene (fading over `fade` seconds) and remembers it for the events. */
  setScene(scene: Scene, fade = 2.5): void {
    this.scene = scene;
    const t = this.ctx.currentTime;
    for (const name of Object.keys(AmbienceEngine.BED_GAIN) as Array<keyof Scene>) {
      const g = this.level.get(name);
      if (!g) continue;
      const target = scene[name] * (AmbienceEngine.BED_GAIN[name] ?? 0);
      g.gain.cancelScheduledValues(t);
      if (fade <= 0) g.gain.setValueAtTime(target, t);
      else g.gain.setTargetAtTime(target, t, fade / 3);
    }
  }

  /** Schedules the single sounds (birds, crickets, drops, bells, owl) up to `until` (context seconds). */
  schedule(until: number): void {
    const now = this.ctx.currentTime;
    const s = this.scene;
    const run = (kind: Kind, lv: number, base: number, play: (t: number, lv: number) => void) => {
      if (lv < 0.03) {
        this.next[kind] = 0;
        return;
      }
      let t = Math.max(this.next[kind], now + 0.05 + this.rand() * 0.5);
      while (t < until) {
        play(t, lv);
        t += (base * (0.6 + 0.8 * this.rand())) / (0.25 + lv);
      }
      this.next[kind] = t;
    };
    run("birds", s.birds, 3, (t, lv) => this.birdPhrase(t, lv));
    run("crickets", s.crickets, 0.32, (t, lv) => this.cricket(t, lv));
    run("drips", s.drips, 0.6, (t, lv) => this.drip(t, lv));
    run("chimes", s.chimes, 7, (t, lv) => this.chime(t, lv));
    run("owl", s.owl, 22, (t, lv) => this.owl(t, lv));
  }

  private voice(): { pan: StereoPannerNode } {
    const pan = this.ctx.createStereoPanner();
    pan.pan.value = (this.rand() - 0.5) * 1.4;
    pan.connect(this.bus);
    return { pan };
  }

  private birdPhrase(t: number, lv: number): void {
    const { pan } = this.voice();
    const notes = 2 + Math.floor(this.rand() * 4);
    const base = 2100 + this.rand() * 2000;
    const rising = this.rand() > 0.5;
    for (let k = 0; k < notes; k++) {
      const at = t + k * (0.1 + this.rand() * 0.05);
      const dur = 0.06 + this.rand() * 0.06;
      const osc = this.ctx.createOscillator();
      const g = this.ctx.createGain();
      osc.type = "sine";
      const f0 = base * (0.92 + this.rand() * 0.16);
      osc.frequency.setValueAtTime(f0, at);
      osc.frequency.exponentialRampToValueAtTime(f0 * (rising ? 1.35 : 0.72), at + dur);
      g.gain.setValueAtTime(0.0001, at);
      g.gain.exponentialRampToValueAtTime(0.16 * lv + 0.04, at + dur * 0.35);
      g.gain.exponentialRampToValueAtTime(0.0001, at + dur);
      osc.connect(g).connect(pan);
      osc.start(at);
      osc.stop(at + dur + 0.02);
    }
  }

  private cricket(t: number, lv: number): void {
    const { pan } = this.voice();
    const osc = this.ctx.createOscillator();
    const g = this.ctx.createGain();
    osc.type = "sine";
    osc.frequency.value = 4200 + this.rand() * 1200;
    g.gain.value = 0.0001;
    const pulses = 3 + Math.floor(this.rand() * 3);
    const peak = (0.03 + this.rand() * 0.02) * (0.4 + lv);
    for (let p = 0; p < pulses; p++) {
      const at = t + p * 0.075;
      g.gain.setValueAtTime(0.0001, at);
      g.gain.linearRampToValueAtTime(peak, at + 0.012);
      g.gain.linearRampToValueAtTime(0.0001, at + 0.05);
    }
    osc.connect(g).connect(pan);
    osc.start(t);
    osc.stop(t + pulses * 0.075 + 0.1);
  }

  private drip(t: number, lv: number): void {
    const { pan } = this.voice();
    const osc = this.ctx.createOscillator();
    const g = this.ctx.createGain();
    const f = 1100 + this.rand() * 1700;
    osc.frequency.setValueAtTime(f, t);
    osc.frequency.exponentialRampToValueAtTime(f * 0.55, t + 0.1);
    g.gain.setValueAtTime(0.0001, t);
    g.gain.exponentialRampToValueAtTime((0.04 + this.rand() * 0.035) * (0.4 + lv), t + 0.004);
    g.gain.exponentialRampToValueAtTime(0.0001, t + 0.14);
    osc.connect(g).connect(pan);
    osc.start(t);
    osc.stop(t + 0.16);
  }

  private chime(t: number, lv: number): void {
    const { pan } = this.voice();
    const f = PENTATONIC[Math.floor(this.rand() * PENTATONIC.length)]!;
    for (const [mult, amp] of [
      [1, 1],
      [2.76, 0.25],
    ] as const) {
      const osc = this.ctx.createOscillator();
      const g = this.ctx.createGain();
      osc.frequency.value = f * mult;
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(0.1 * lv * amp + 0.0002, t + 0.006);
      g.gain.exponentialRampToValueAtTime(0.0001, t + 3.2);
      osc.connect(g).connect(pan);
      osc.start(t);
      osc.stop(t + 3.3);
    }
  }

  private owl(t: number, lv: number): void {
    const { pan } = this.voice();
    const tone = this.filter("lowpass", 900);
    tone.connect(pan);
    const hoots: Array<[number, number]> = [
      [0, 0.28],
      [0.5, 0.28],
      [1.1, 0.7],
    ];
    for (const [offset, len] of hoots) {
      const at = t + offset;
      for (const [mult, amp] of [
        [1, 1],
        [2, 0.25],
      ] as const) {
        const osc = this.ctx.createOscillator();
        const g = this.ctx.createGain();
        osc.frequency.setValueAtTime(OWL_HZ * mult * 1.06, at);
        osc.frequency.linearRampToValueAtTime(OWL_HZ * mult * 0.96, at + len);
        g.gain.setValueAtTime(0.0001, at);
        g.gain.linearRampToValueAtTime(0.13 * lv * amp + 0.0002, at + 0.07);
        g.gain.linearRampToValueAtTime(0.0001, at + len + 0.08);
        osc.connect(g).connect(tone);
        osc.start(at);
        osc.stop(at + len + 0.12);
      }
    }
  }

  /** A crack and a rolling rumble starting at context time t (when `thunder` is in the scene). */
  thunderAt(t: number): void {
    const th = this.scene.thunder;
    if (th < 0.05) return;
    const roll = (at: number, peak: number, length: number) => {
      const src = this.ctx.createBufferSource();
      src.buffer = this.noise;
      const low = this.filter("lowpass", 520);
      low.frequency.setValueAtTime(520, at);
      low.frequency.exponentialRampToValueAtTime(70, at + length);
      const g = this.ctx.createGain();
      g.gain.setValueAtTime(0.0001, at);
      g.gain.exponentialRampToValueAtTime(peak, at + 0.12);
      g.gain.setTargetAtTime(0.0001, at + 0.5, length / 4);
      src.connect(low).connect(g).connect(this.bus);
      src.start(at, this.rand() * 4);
      src.stop(at + length + 0.5);
    };
    const crack = this.ctx.createBufferSource();
    crack.buffer = this.noise;
    const hi = this.filter("highpass", 1400);
    const cg = this.ctx.createGain();
    cg.gain.setValueAtTime(0.0001, t);
    cg.gain.exponentialRampToValueAtTime(0.07 * th, t + 0.01);
    cg.gain.exponentialRampToValueAtTime(0.0001, t + 0.16);
    crack.connect(hi).connect(cg).connect(this.bus);
    crack.start(t, this.rand() * 4);
    crack.stop(t + 0.2);
    roll(t + 0.05, 0.4 * th, 5.2);
    roll(t + 1.3 + this.rand() * 0.8, 0.22 * th, 3.6);
  }

  /** Silences everything and releases the always-on sources. */
  dispose(): void {
    for (const n of this.stopNodes) {
      try {
        n.stop();
      } catch {
        // already stopped
      }
    }
    this.stopNodes = [];
    this.bus.disconnect();
  }
}
