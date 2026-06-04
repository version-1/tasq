import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": "/src",
    },
  },
  server: {
    proxy: {
      "/tracker": "http://127.0.0.1:37653",
      "/orchestrator": "http://127.0.0.1:37653",
    },
  },
});
