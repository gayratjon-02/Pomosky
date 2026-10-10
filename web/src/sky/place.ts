import { DEFAULT_PLACE } from "../core/places";
import type { StateData } from "../core/protocol";

/** Where the sky is drawn: the host's place when set, otherwise a default. */
export function skyPlace(state: StateData | null): { lat: number; lon: number } {
  const loc = state?.location;
  return loc ? { lat: loc.lat, lon: loc.lon } : { lat: DEFAULT_PLACE.lat, lon: DEFAULT_PLACE.lon };
}
