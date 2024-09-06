import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import viteCompression from 'vite-plugin-compression';

export default defineConfig(function () {
	const buildTimestamp = new Date();
	return {
		plugins: [
			svelte(),
			viteCompression({
				filter: /^(?!.*pdf\.worker\.min-[A-Z0-9]+\.mjs$).*\.(js|mjs|json|css|html)$/i,
			}),
		],
		define: {
			'import.meta.env.BUILD_TIMESTAMP': JSON.stringify(buildTimestamp.toLocaleString()),
		},
	};
});
