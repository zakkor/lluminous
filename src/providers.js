import { get } from 'svelte/store';
import {
	openaiAPIKey,
	groqAPIKey,
	openrouterAPIKey,
	remoteServer,
	mistralAPIKey,
} from './stores.js';

export const providers = [
	{ name: 'OpenAI', url: 'https://api.openai.com', apiKeyFn: () => get(openaiAPIKey) },
	{ name: 'OpenRouter', url: 'https://openrouter.ai/api', apiKeyFn: () => get(openrouterAPIKey) },
	{ name: 'Groq', url: 'https://api.groq.com/openai', apiKeyFn: () => get(groqAPIKey) },
	{ name: 'Mistral', url: 'https://api.mistral.ai', apiKeyFn: () => get(mistralAPIKey) },
	{ name: 'Local', url: 'http://localhost:8081', apiKeyFn: () => get(remoteServer).password },
].filter(Boolean);
