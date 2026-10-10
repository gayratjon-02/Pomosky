import { initialOf } from "../stats/format";

/** A round picture of a person, or their initial when there is no picture or it fails to load. */
export function avatarElement(name: string, url: string | undefined, small = true): HTMLElement {
  const box = document.createElement("span");
  box.className = small ? "avatar avatar-sm" : "avatar";
  const initial = () => {
    const s = document.createElement("span");
    s.className = "avatar-initial";
    s.textContent = initialOf(name);
    return s;
  };
  if (!url) {
    box.appendChild(initial());
    return box;
  }
  const img = document.createElement("img");
  img.className = "avatar-img";
  img.src = url;
  img.alt = "";
  img.width = 64;
  img.height = 64;
  img.referrerPolicy = "no-referrer";
  img.loading = "lazy";
  img.addEventListener("error", () => box.replaceChildren(initial()));
  box.appendChild(img);
  return box;
}
