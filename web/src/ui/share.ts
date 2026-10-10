import { maybeById } from "../core/dom";
import { toast } from "./toast";

/** Copy-link and fullscreen buttons shared by the room, host and record pages. */
export function initShare(): void {
  maybeById("copy-link")?.addEventListener("click", async () => {
    const url = maybeById("copy-link")?.dataset["url"] ?? location.href.split("#")[0]!;
    try {
      await navigator.clipboard.writeText(url);
      toast("Link copied");
    } catch {
      toast("Could not copy. Select the link and copy it by hand.");
    }
  });

  const fs = maybeById("fullscreen");
  if (fs && !document.documentElement.requestFullscreen) fs.hidden = true;
  fs?.addEventListener("click", () => {
    if (document.fullscreenElement) void document.exitFullscreen();
    else void document.documentElement.requestFullscreen().catch(() => toast("Fullscreen is not available"));
  });
}
