import { defineConfig } from "vite";

// Built files carry a content hash in their name, and a manifest maps each entry
// ("landing", "base", ...) to its current file. The Go templates look names up in the
// manifest (see internal/assets), so a new build can never mix with cached old files.
export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    manifest: true, // dist/.vite/manifest.json
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
        admin: "src/entries/admin.ts",
        base: "src/styles/base.css",
      },
      output: {
        entryFileNames: "[name].[hash].js",
        chunkFileNames: "chunks/[name].[hash].js",
        assetFileNames: "[name].[hash][extname]",
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
