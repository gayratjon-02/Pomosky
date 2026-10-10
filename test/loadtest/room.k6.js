// k6 load test: many viewers connect to one room over WebSocket.
// Run: k6 run -e BASE=wss://pomosky.example.com -e ROOM=abcd2345 test/loadtest/room.k6.js
// TODO(stage 8): tune stages to the target concurrency and watch VM CPU/RAM.
import ws from "k6/ws";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 100 },
    { duration: "1m", target: 100 },
    { duration: "30s", target: 0 },
  ],
};

export default function () {
  const url = `${__ENV.BASE}/ws/${__ENV.ROOM}`;
  const res = ws.connect(url, {}, (socket) => {
    socket.on("open", () => socket.setInterval(() => socket.send(JSON.stringify({ type: "ping" })), 25000));
    socket.setTimeout(() => socket.close(), 60000);
  });
  check(res, { "status is 101": (r) => r && r.status === 101 });
}
