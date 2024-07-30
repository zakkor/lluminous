import { get } from 'svelte/store';
import {
	openaiAPIKey,
	groqAPIKey,
	openrouterAPIKey,
	remoteServer,
	mistralAPIKey,
	anthropicAPIKey,
} from './stores.js';

export const providers = [
	{
		name: 'OpenRouter',
		url: 'https://openrouter.ai/api',
		completionUrl: '/v1/chat/completions',
		apiKeyFn: () => get(openrouterAPIKey),
	},
	{
		name: 'Anthropic',
		url: 'http://localhost:8090/https://api.anthropic.com',
		completionUrl: '/v1/messages',
		apiKeyFn: () => get(anthropicAPIKey),
	},
	{
		name: 'OpenAI',
		url: 'https://api.openai.com',
		completionUrl: '/v1/chat/completions',
		apiKeyFn: () => get(openaiAPIKey),
	},
	{
		name: 'Groq',
		url: 'https://api.groq.com/openai',
		completionUrl: '/v1/chat/completions',
		apiKeyFn: () => get(groqAPIKey),
	},
	{
		name: 'Mistral',
		url: 'https://api.mistral.ai',
		completionUrl: '/v1/chat/completions',
		apiKeyFn: () => get(mistralAPIKey),
	},
	// { name: 'Local', url: 'http://localhost:8081', apiKeyFn: () => get(remoteServer).password },
].filter(Boolean);

// OpenAI provider: OpenAI doesn't provide any metadata for their models, so we have to harddcode which ones are multimodal
export const additionalModelsMultimodal = ['gpt-4o', 'gpt-4-turbo', 'gpt-4-turbo-2024-04-09'];

// OpenAI provider:
export const imageGenerationModels = ['dall-e-3'];

export function hasCompanyLogo(model) {
	return (
		model &&
		model.provider &&
		(model.provider == 'Anthropic' ||
			model.provider === 'OpenAI' ||
			model.provider === 'Mistral' ||
			model.provider === 'Groq' ||
			model.id.startsWith('openai') ||
			model.id.startsWith('anthropic') ||
			model.id.startsWith('meta-llama') ||
			model.id.startsWith('mistralai') ||
			model.id.startsWith('cohere') ||
			model.id.startsWith('nous') ||
			model.id.startsWith('google') ||
			model.id.startsWith('perplexity') ||
			model.id.startsWith('deepseek'))
	);
}

export function formatModelName(model, short = false) {
	if (model.id === null) {
		return model.name;
	}

	if (short) {
		const split = model.name.split(': ');
		if (split.length > 1) {
			return split[1];
		}
	}

	let name = model.name;

	// If providers clash, disambiguate provider name
	const disambiguate =
		(get(openaiAPIKey).length > 0 && get(openrouterAPIKey).length > 0) ||
		(get(anthropicAPIKey).length > 0 && get(openrouterAPIKey).length > 0);

	if (model.provider === 'OpenAI') {
		name = name.replace('gpt', 'GPT').replace('dall-e-3', 'DALL-E 3');
	}

	if (model.provider === 'Groq') {
		return model.provider + ': ' + name;
	}
	if (disambiguate && model.provider === 'OpenRouter' && name.includes(': ')) {
		return model.provider + ': ' + name.split(': ')[1];
	}
	if (disambiguate) {
		return model.provider + ': ' + name;
	}

	return name;
}
