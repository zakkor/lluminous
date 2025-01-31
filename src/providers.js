import { get } from 'svelte/store';
import {
	openaiAPIKey,
	groqAPIKey,
	openrouterAPIKey,
	remoteServer,
	mistralAPIKey,
	anthropicAPIKey,
} from './stores.js';

// Anthropic provider:
export const anthropicModels = [
	{
		id: 'claude-3-5-sonnet-20241022',
		name: 'Claude 3.5 Sonnet',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
	{
		id: 'claude-3-5-sonnet-20240620',
		name: 'Claude 3.5 Sonnet (older)',
		provider: 'Anthropic',
		modality: 'text+image->text',
	},
	{
		id: 'claude-3-5-haiku-20241022',
		name: 'Claude 3.5 Haiku',
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

export const providers = [
	{
		name: 'OpenRouter',
		url: 'https://openrouter.ai/api',
		completionUrl: '/v1/chat/completions',
		modelsUrl: '/v1/models',
		apiKeyFn: () => get(openrouterAPIKey),
	},
	{
		name: 'Anthropic',
		url: 'https://api.anthropic.com',
		completionUrl: '/v1/messages',
		modelsUrl: null,
		models: anthropicModels,
		apiKeyFn: () => get(anthropicAPIKey),
	},
	{
		name: 'OpenAI',
		url: 'https://api.openai.com',
		completionUrl: '/v1/chat/completions',
		modelsUrl: '/v1/models',
		apiKeyFn: () => get(openaiAPIKey),
	},
	{
		name: 'Groq',
		url: 'https://api.groq.com/openai',
		completionUrl: '/v1/chat/completions',
		modelsUrl: '/v1/models',
		apiKeyFn: () => get(groqAPIKey),
	},
	{
		name: 'Mistral',
		url: 'https://api.mistral.ai',
		completionUrl: '/v1/chat/completions',
		modelsUrl: '/v1/models',
		apiKeyFn: () => get(mistralAPIKey),
	},
	{
		name: 'Ollama',
		url: 'http://localhost:11434',
		completionUrl: '/v1/chat/completions',
		modelsUrl: '/api/tags',
		responseMapperFn: (json) => {
			return json.models.map((m) => ({
				id: m.name,
				name: m.name,
				provider: 'Ollama',
				modality: 'text->text',
			}));
		},
		apiKeyFn: () => false,
	},
].filter(Boolean);

// OpenAI provider: OpenAI doesn't provide any metadata for their models, so we have to harddcode which ones are multimodal
export const openAIAdditionalModelsMultimodal = ['gpt-4o', 'gpt-4-turbo', 'gpt-4-turbo-2024-04-09'];

export const thinkingModels = ['openai/o1-preview', 'openai/o1-mini'];

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
	{ fromProvider: 'Ollama' },
	{
		exactly: [
			'anthropic/claude-3.5-sonnet',
			'anthropic/claude-3.5-haiku',
			'anthropic/claude-3-opus',
		],
	},
	{
		exactly: ['claude-3-5-sonnet-20241022', 'claude-3-5-haiku-20241022', 'claude-3-opus-20240229'],
	},
	{ exactly: ['deepseek/deepseek-r1', 'deepseek/deepseek-chat'] },
	{
		exactly: [
			'openai/o1-preview',
			'openai/o1-mini',
			'openai/gpt-4-turbo',
			'openai/chatgpt-4o-latest',
			'openai/gpt-4o',
			'openai/gpt-4o-mini',
		],
	},
	{ exactly: ['o1-preview', 'o1-mini', 'gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'dall-e-3'] },
	{ startsWith: ['qwen'] },
	{
		exactly: [
			'meta-llama/llama-3.3-70b-instruct',
			'meta-llama/llama-3.1-405b-instruct',
			'meta-llama/llama-3.1-405b',
		],
	},
	{ fromProvider: 'Groq' },
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
			model.id.startsWith('deepseek') ||
			model.id.startsWith('qwen'))
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

export function formatMultipleModelNames(models, short = false) {
	if (models.length === 1) {
		return formatModelName(models[0], short);
	}
	if (models.length === 2) {
		return formatModelName(models[0], short) + ', ' + formatModelName(models[1], short);
	}
	return (
		formatModelName(models[0], short) +
		', ' +
		formatModelName(models[1], short) +
		', + ' +
		(models.length - 2)
	);
}
