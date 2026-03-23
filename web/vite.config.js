import {defineConfig} from "vite";
import react from "@vitejs/plugin-react";
import {nodePolyfills} from "vite-plugin-node-polyfills";
import path from "path";

export default defineConfig({
  plugins: [
    react(),
    nodePolyfills({
      include: ["buffer", "process"],
      globals: {
        Buffer: true,
        process: true,
      },
    }),
  ],
  // Allow JSX in .js files (CRA convention used throughout this project)
  esbuild: {
    loader: "jsx",
    include: /src\/.*\.js$/,
    exclude: [],
  },
  optimizeDeps: {
    esbuildOptions: {
      loader: {
        ".js": "jsx",
      },
    },
  },
  css: {
    preprocessorOptions: {
      less: {
        modifyVars: {
          "@primary-color": "rgb(89,54,213)",
          "@border-radius-base": "5px",
        },
        javascriptEnabled: true,
      },
    },
  },
  server: {
    port: 7001,
    proxy: {
      "/api": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/swagger": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/files": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/.well-known": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/cas": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/scim": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "build-temp",
    sourcemap: false,
  },
  resolve: {
    alias: {
      "~": path.resolve(__dirname, "src"),
    },
  },
});
