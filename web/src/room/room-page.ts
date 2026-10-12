import { maybeById } from "../core/dom";
import { load, save } from "../core/storage";
import { RoomClient } from "../core/room-client";
import { startLiveSky } from "../sky/live-sky";
import { setPeople } from "../ui/people";
import { setViewers } from "../ui/presence";
import { initReactions, setReactionCounts } from "../ui/reactions";
import { initShare } from "../ui/share";
import { initMusic } from "../ui/music-events";
import { initNotifyToggle, notifyEnd } from "../ui/notify";
import { initSounds } from "../ui/sound-events";
import { initWakeLock } from "../ui/wake-lock";
import { TimerRenderer } from "../ui/timer-view";
import { toast } from "../ui/toast";
import { createConnectionChip } from "../ui/connection";
import { t } from "../core/i18n";

const NICK_KEY = "pomosky.nick";

/** Viewer page: connects to the room, shows the timer, reactions and the live sky. */
export function initRoomPage(): void {
  const roomId = location.pathname.split("/")[2];
  if (!roomId) return;

  const nickname = load(NICK_KEY) ?? "";
  const client = new RoomClient(roomId, nickname ? { nickname } : {});

  startLiveSky(maybeById<HTMLCanvasElement>("sky"), client, roomId);

  const nickInput = maybeById<HTMLInputElement>("nick");
  if (nickInput) nickInput.value = nickname;

  const renderer = new TimerRenderer();
  const chip = createConnectionChip();
  const floatReaction = initReactions((kind) => client.react(kind));
  initShare();

  client.on("status", chip);
  client.on("state", (s) => {
    setViewers(s.viewers);
    setReactionCounts(s.reactions);
  });
  client.on("people", setPeople);
  client.on("tick", (t) => {
    floatReaction(t.kind);
    setReactionCounts({ [t.kind]: t.count });
  });
  client.on("error", (e) => {
    if (e.code === "rate_limited") toast(t("Slow down a little"));
  });

  window.setInterval(() => {
    if (client.state) renderer.render(client.state.timer, client.clock.serverNow());
  }, 250);

  maybeById("nick-form")?.addEventListener("submit", (e) => {
    e.preventDefault();
    const value = (nickInput?.value ?? "").trim();
    save(NICK_KEY, value);
    client.setNickname(value);
    toast(value ? t("Hello, {name}", { name: value }) : t("Name cleared"));
  });

  initSounds(client);
  renderer.onEnd = notifyEnd;
  initNotifyToggle(client, roomId);
  initMusic();
  initWakeLock(client);

  client.connect();
}
