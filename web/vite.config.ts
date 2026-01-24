import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/clips': {
				target: 'http://localhost:8080',
				changeOrigin: true
			},
			'/metadata': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
