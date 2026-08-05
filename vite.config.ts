import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

const backendUrl = 'http://127.0.0.1:8080';

export default defineConfig({
	plugins: [svelte(), tailwindcss()],
	server: {
		host: '127.0.0.1',
		strictPort: true,
		proxy: {
			'/auth': {
				target: backendUrl,
				changeOrigin: true,
			},
			'/users': {
				target: backendUrl,
				changeOrigin: true,
			},
			'/api': {
				target: backendUrl,
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api/, ''),
			},
		},
	},
});
