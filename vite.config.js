import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import viteCompression from 'vite-plugin-compression';

export default defineConfig(function () {
	const buildTimestamp = new Date();
	return {
		plugins: [svelte(), viteCompression()],
		define: {
			'import.meta.env.BUILD_TIMESTAMP': JSON.stringify(buildTimestamp.toLocaleString()),
		},
	};
});
