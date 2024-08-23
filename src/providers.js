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
		url: 'https://api.anthropic.com',
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
	{
		name: 'Local',
		url: get(remoteServer).address || 'http://localhost:8081',
		completionUrl: '/v1/chat/completions',
		apiKeyFn: () => get(remoteServer).password,
	},
].filter(Boolean);

// Anthropic provider:
export const anthropicModels = [
	{
		id: 'claude-3-5-sonnet-20240620',
		name: 'Claude 3.5 Sonnet',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
	{
		id: 'claude-3-opus-20240229',
		name: 'Claude 3 Opus',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
	{
		id: 'claude-3-sonnet-20240229',
		name: 'Claude 3 Sonnet',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
	{
		id: 'claude-3-haiku-20240307',
		name: 'Claude 3 Haiku',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
];

// OpenAI provider: OpenAI doesn't provide any metadata for their models, so we have to harddcode which ones are multimodal
export const openAIAdditionalModelsMultimodal = ['gpt-4o', 'gpt-4-turbo', 'gpt-4-turbo-2024-04-09'];

// OpenAI provider:
export const openAIImageGenerationModels = ['dall-e-3'];

export const openAIIgnoreIds = [
	'dall-e-2',
	'whisper-1',
	'davinci-002',
	'tts-1-hd-1106',
	'tts-1-hd',
	'tts-1',
	'babbage-002',
	'tts-1-1106',
	'text-embedding-3-large',
	'text-embedding-3-small',
	'text-embedding-ada-002',
];

export const priorityOrder = [
	{ fromProvider: 'Local' },
	{
		exactly: ['openai/gpt-4-turbo', 'openai/gpt-4o', 'openai/gpt-4o-mini'],
	},
	{ exactly: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'dall-e-3'] },
	{
		exactly: [
			'anthropic/claude-3.5-sonnet',
			'anthropic/claude-3-opus',
			'anthropic/claude-3-sonnet',
			'anthropic/claude-3-haiku',
		],
	},
	{
		fromProvider: 'Anthropic',
		exactly: [
			'claude-3-5-sonnet-20240620',
			'claude-3-opus-20240229',
			'claude-3-sonnet-20240229',
			'claude-3-haiku-20240307',
		],
	},
	{
		exactly: ['llama-3.1-70b-versatile', 'llama-3.1-8b-instant'],
	},
	{
		exactly: [
			'meta-llama/llama-3.1-405b-instruct',
			'meta-llama/llama-3.1-70b-instruct',
			'meta-llama/llama-3.1-8b-instruct:free',
			'meta-llama/llama-3-70b-instruct',
			'meta-llama/llama-3-8b-instruct:free',
		],
	},
	{ startsWith: ['deepseek'] },
	{
		exactly: [
			'mistralai/mixtral-8x22b-instruct',
			'mistralai/mistral-large',
			'mistralai/mistral-medium',
			'mistralai/mistral-small',
		],
	},
	{ exactly: ['google/gemini-flash-1.5', 'google/gemini-pro-1.5'] },
	{ startsWith: ['cohere/'] },
	{
		exactly: [
			'perplexity/llama-3-sonar-large-32k-online',
			'perplexity/llama-3-sonar-small-32k-online',
		],
	},
	{ startsWith: ['nousresearch/'] },
	{ fromProvider: 'OpenAI' },
	{
		startsWith: [
			'anthropic/claude-2',
			'anthropic/claude-2.1',
			'anthropic/claude-2.0',
			'anthropic/claude-instant-1',
		],
		exactlyNot: [
			'anthropic/claude-2',
			'anthropic/claude-2.1',
			'anthropic/claude-2.0',
			'anthropic/claude-instant-1',
			'anthropic/claude-instant-1.0',
			'anthropic/claude-instant-1.1',
			'anthropic/claude-instant-1.2',
			'anthropic/claude-1.2',
			'anthropic/claude-1',
			'anthropic/claude-2:beta',
			'anthropic/claude-2.0:beta',
			'anthropic/claude-2.1:beta',
			'anthropic/claude-instant-1:beta',
		],
	},
	{
		startsWith: ['openai/gpt-3.5-turbo', 'openai/gpt-4'],
		exactlyNot: [
			'openai/gpt-3.5-turbo-0125',
			'openai/gpt-3.5-turbo-0301',
			'openai/gpt-3.5-turbo-0613',
			'openai/gpt-3.5-turbo-1106',
			'openai/gpt-3.5-turbo-instruct',
			'openai/gpt-4',
			'openai/gpt-4-0314',
			'openai/gpt-4-1106-preview',
			'openai/gpt-4-32k-0314',
		],
	},
	{ fromProvider: 'Mistral' },
	{ startsWith: ['mistralai/'] },
];

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
