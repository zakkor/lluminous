import { get } from 'svelte/store';
import { groqAPIKey, openrouterAPIKey } from './stores.js';

export const providers = [
	{ name: 'OpenRouter', url: 'https://openrouter.ai/api', apiKeyFn: () => get(openrouterAPIKey) },
	{ name: 'Groq', url: 'https://api.groq.com/openai', apiKeyFn: () => get(groqAPIKey) },
	// TODO:
	// { name: 'Local', url: 'http://localhost:8081' },
].filter(Boolean);
