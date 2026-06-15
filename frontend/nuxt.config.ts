import { defineNuxtConfig } from "nuxt/config";

// https://v3.nuxtjs.org/api/configuration/nuxt.config
export default defineNuxtConfig({
  ssr: false,
  vite: {
    // Nuxt's dev warmup can fail on project paths with spaces (e.g. "Go Projects").
    warmupEntry: false,
  },
  runtimeConfig: {
    public: {
      pbUrl: process.env.NUXT_PUBLIC_PB_URL || (process.env.NODE_ENV === "development" ? "http://localhost:7745" : ""),
    },
  },
  modules: [
    "@nuxtjs/tailwindcss",
    "@pinia/nuxt",
    "@vueuse/nuxt",
    "@vite-pwa/nuxt",
    "unplugin-icons/nuxt",
  ],
  nitro: {
    devProxy: {
      "/api": {
        target: "http://localhost:7745/api",
        ws: true,
        changeOrigin: true,
      },
    },
  },
  css: ["@/assets/css/main.css"],
  pwa: {
    registerType: "prompt",
    injectRegister: "script",
    workbox: {
      navigateFallback: "/index.html",
      navigateFallbackDenylist: [/^\/api/],
      globPatterns: ["**/*.{js,css,html,ico,png,svg,jpg,webmanifest}"],
    },
    devOptions: {
      // Enable to troubleshoot during development
      enabled: false,
    },
    manifest: {
      id: "/",
      name: "Homebox",
      short_name: "Homebox",
      description: "Home Inventory App",
      theme_color: "#5b7f67",
      background_color: "#FFFFFF",
      display: "standalone",
      scope: "/",
      start_url: "/home",
      icons: [
        {
          src: "pwa-192x192.png",
          sizes: "192x192",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
          purpose: "maskable",
        },
      ],
    },
  },
});
