import { defineConfig } from "vite";

// Stable output names so the Go templates can reference /static/<name>.js directly.
export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    rollupOptions: {
      input: {
        landing: "src/entries/landing.ts",
        room: "src/entries/room.ts",
        host: "src/entries/host.ts",
        record: "src/entries/record.ts",
        skylab: "src/entries/skylab.ts",
        ambient: "src/entries/ambient.ts",
        account: "src/entries/account.ts",
        site: "src/entries/site.ts",
        base: "src/styles/base.css",
      },
      output: {
        entryFileNames: "[name].js",
        chunkFileNames: "chunks/[name].js",
        assetFileNames: "[name][extname]",
      },
    },
  },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/ws": { target: "ws://localhost:8080", ws: true },
    },
  },
});
