/** True when the user asked the OS to reduce motion. */
export function prefersReducedMotion(): boolean {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/** Calls onChange(true|false) when the tab becomes visible/hidden. */
export function onVisibilityChange(onChange: (visible: boolean) => void): () => void {
  const handler = () => onChange(document.visibilityState === "visible");
  document.addEventListener("visibilitychange", handler);
  return () => document.removeEventListener("visibilitychange", handler);
}
