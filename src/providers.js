import { get } from 'svelte/store';
import { openaiAPIKey, groqAPIKey, openrouterAPIKey } from './stores.js';

export const providers = [
	{ name: 'OpenAI', url: 'https://api.openai.com', apiKeyFn: () => get(openaiAPIKey) },
	{ name: 'OpenRouter', url: 'https://openrouter.ai/api', apiKeyFn: () => get(openrouterAPIKey) },
	{ name: 'Groq', url: 'https://api.groq.com/openai', apiKeyFn: () => get(groqAPIKey) },
	// TODO:
	// { name: 'Local', url: 'http://localhost:8081' },
].filter(Boolean);
