import { maybeById } from "../core/dom";
import { load, save } from "../core/storage";
import { RoomClient } from "../core/room-client";
import { createSky } from "../sky";
import { astroInputs } from "../sky/inputs";
import { setViewers } from "../ui/presence";
import { initReactions, setReactionCounts } from "../ui/reactions";
import { initShare } from "../ui/share";
import { setSoundEnabled, soundEnabled } from "../ui/sound";
import { TimerRenderer } from "../ui/timer-view";
import { toast } from "../ui/toast";
import { createConnectionChip } from "../ui/connection";

// TODO(stage 6): replace with the host's rounded location once the room carries it.
const DEFAULT_LAT = 37.57;
const DEFAULT_LON = 126.98;
const NICK_KEY = "pomosky.nick";

/** Viewer page: connects to the room, shows the timer, reactions and the live sky. */
export function initRoomPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const canvas = maybeById<HTMLCanvasElement>("sky");
  const sky = canvas ? createSky(canvas) : null;
  const refreshSky = () => sky?.update(astroInputs(new Date(), DEFAULT_LAT, DEFAULT_LON));
  refreshSky();
  sky?.start();
  window.setInterval(refreshSky, 30_000);

  const nickname = load(NICK_KEY) ?? "";
  const nickInput = maybeById<HTMLInputElement>("nick");
  if (nickInput) nickInput.value = nickname;

  const client = new RoomClient(roomId, nickname ? { nickname } : {});
  const renderer = new TimerRenderer();
  const chip = createConnectionChip();
  const floatReaction = initReactions((kind) => client.react(kind));
  initShare();

  client.on("status", chip);
  client.on("state", (s) => {
    setViewers(s.viewers);
    setReactionCounts(s.reactions);
  });
  client.on("tick", (t) => {
    floatReaction(t.kind);
    setReactionCounts({ [t.kind]: t.count });
  });
  client.on("error", (e) => {
    if (e.code === "rate_limited") toast("Slow down a little");
  });

  window.setInterval(() => {
    if (client.state) renderer.render(client.state.timer, client.clock.serverNow());
  }, 250);

  maybeById("nick-form")?.addEventListener("submit", (e) => {
    e.preventDefault();
    const value = (nickInput?.value ?? "").trim();
    save(NICK_KEY, value);
    client.setNickname(value);
    toast(value ? `Hello, ${value}` : "Name cleared");
  });

  const soundBtn = maybeById<HTMLButtonElement>("sound-toggle");
  const paintSound = () => {
    if (!soundBtn) return;
    soundBtn.textContent = soundEnabled() ? "Sound on" : "Sound off";
    soundBtn.setAttribute("aria-pressed", String(soundEnabled()));
  };
  soundBtn?.addEventListener("click", () => {
    setSoundEnabled(!soundEnabled());
    paintSound();
  });
  paintSound();

  client.connect();
}
