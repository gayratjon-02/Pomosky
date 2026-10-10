// Server clock synchronisation. serverTime ~= clientTime + offset.

/** Estimates the offset from one request/response round trip. */
export function estimateOffset(serverNowMs: number, sentAtMs: number, receivedAtMs: number): number {
  const rtt = receivedAtMs - sentAtMs;
  return serverNowMs - (sentAtMs + rtt / 2);
}

/** Keeps the median of the latest samples, which is robust against slow round trips. */
export class ClockSync {
  private samples: number[] = [];

  constructor(private readonly maxSamples = 8) {}

  add(offsetMs: number): void {
    this.samples.push(offsetMs);
    if (this.samples.length > this.maxSamples) this.samples.shift();
  }

  get sampleCount(): number {
    return this.samples.length;
  }

  get offsetMs(): number {
    if (this.samples.length === 0) return 0;
    const sorted = [...this.samples].sort((a, b) => a - b);
    return sorted[Math.floor(sorted.length / 2)]!;
  }

  serverNow(clientNowMs: number = Date.now()): number {
    return clientNowMs + this.offsetMs;
  }
}
