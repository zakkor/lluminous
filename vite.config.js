import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import viteCompression from 'vite-plugin-compression';

export default defineConfig({
	plugins: [svelte(), viteCompression()],
});
