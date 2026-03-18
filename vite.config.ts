import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],

  build: {
    outDir: 'bck/internal/ui/dist',
    emptyOutDir: true,
  },

  resolve: {
    alias: {
      '@/*': '/src/*',
    },
  },

  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL,
        changeOrigin: true,
      },
    },
  },
});
