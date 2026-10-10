import type { Condition } from "./types";

// Maps a WMO weather code to the coarse condition used by the renderer.
// TODO(stage 6): verify the code table in the Open-Meteo documentation.
export function conditionFromCode(_code: number): Condition {
  return "unknown";
}
