import { get } from 'svelte/store';
import { controller, params, toolSchema } from './stores.js';
import { providers } from './providers.js';

export async function complete(convo, onupdate, onabort) {
	controller.set(new AbortController());

	const model = convo.models[0];

	if (model.provider === 'Local') {
		if (!model.template) {
			model.template = 'chatml';
		}
		const response = await fetch('http://localhost:8082/completion', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			signal: get(controller).signal,
			body: JSON.stringify({
				stream: true,
				prompt: conversationToString(convo),
				stop: conversationStop(convo),
				n_predict: -1,
				repeat_penalty: 1.1,
				cache_prompt: true,
				...(convo.grammar !== '' && { grammar: convo.grammar }),
			}),
		});

		streamResponse(model.provider, response.body, onupdate, onabort);
	} else {
		const param = get(params);

		const openAICompatibleFormat =
			model.provider === 'OpenAI' ||
			model.provider === 'OpenRouter' ||
			model.provider === 'Groq' ||
			model.provider === 'Mistral';

		let messages = convo.messages.map(
			openAICompatibleFormat ? messageToOpenAIFormat : messageToAnthropicFormat
		);

		let system = undefined;
		if (model.provider === 'Anthropic' && messages[0].role === 'system') {
			system = messages.shift().content;
		}

		if (param.messagesContextLimit > 0) {
			messages = limitMessagesContext(messages, param.messagesContextLimit);
		}

		// TODO: Actually it works with Anthropic also. How to show it as disabled for unsupported?
		// Filter out unclosed messages from being submitted if using external models
		if (
			convo.messages[convo.messages.length - 1].unclosed &&
			convo.messages[convo.messages.length - 1].content === ''
		) {
			messages.pop();
		}

		const schema = get(toolSchema)
			.map((group) => group.schema)
			.flat();

		const activeSchema = schema
			.filter((tool) => (convo.tools || []).includes(tool.function.name))
			.map((tool) => {
				if (model.provider === 'Anthropic') {
					return toolSchemaToAnthropicFormat(tool);
				}
				return {
					type: tool.type,
					function: tool.function,
				};
			});

		const provider = providers.find((p) => p.name === model.provider);

		const completions = async (signal) => {
			return fetch(`${provider.url}${provider.completionUrl}`, {
				method: 'POST',
				headers: {
					...(model.provider === 'OpenRouter' ||
					model.provider === 'OpenAI' ||
					model.provider === 'Groq' ||
					model.provider === 'Mistral'
						? {
								Authorization: `Bearer ${provider.apiKeyFn()}`,
							}
						: model.provider === 'Anthropic'
							? {
									'x-api-key': provider.apiKeyFn(),
									'anthropic-version': '2023-06-01',
									'anthropic-dangerous-direct-browser-access': 'true',
								}
							: {}),
					'Content-Type': 'application/json',
					...(model.provider === 'OpenRouter'
						? {
								'HTTP-Referer': 'https://lluminous.chat',
								'X-Title': 'lluminous',
							}
						: {}),
				},
				signal,
				body: JSON.stringify({
					stream: true,
					model: model.id,
					temperature: param.temperature,
					max_tokens:
						param.maxTokens != null && param.maxTokens > 0
							? param.maxTokens
							: model.provider === 'Anthropic'
								? 4096
								: undefined,
					tools: activeSchema.length > 0 ? activeSchema : undefined,
					system,
					messages,
				}),
			});
		};

		const response = await completions(get(controller).signal);
		streamResponse(model.provider, response.body, onupdate, onabort);
	}
}

export async function completeConsensus(convo, onupdate, onabort) {
	controller.set(new AbortController());
	const param = get(params);

	// Query all models in parallel
	const modelResults = await Promise.all(
		convo.models.map(async (model) => {
			const openAICompatibleFormat =
				model.provider === 'OpenAI' ||
				model.provider === 'OpenRouter' ||
				model.provider === 'Groq' ||
				model.provider === 'Mistral';

			let messages = convo.messages.map(
				openAICompatibleFormat ? messageToOpenAIFormat : messageToAnthropicFormat
			);

			if (param.messagesContextLimit > 0) {
				messages = limitMessagesContext(messages, param.messagesContextLimit);
			}

			const provider = providers.find((p) => p.name === model.provider);

			const response = await fetch(`${provider.url}${provider.completionUrl}`, {
				method: 'POST',
				headers: {
					...(model.provider === 'OpenRouter' ||
					model.provider === 'OpenAI' ||
					model.provider === 'Groq' ||
					model.provider === 'Mistral'
						? { Authorization: `Bearer ${provider.apiKeyFn()}` }
						: model.provider === 'Anthropic'
							? {
									'x-api-key': provider.apiKeyFn(),
									'anthropic-version': '2023-06-01',
								}
							: {}),
					'Content-Type': 'application/json',
				},
				signal: get(controller).signal,
				body: JSON.stringify({
					stream: false,
					model: model.id,
					temperature: param.temperature,
					messages,
				}),
			});

			const result = await response.json();
			return {
				model: model.id,
				response: result.choices[0].message.content,
			};
		})
	);

	// Prepare consensus query
	const consensusProvider = providers.find((p) => p.name === 'OpenRouter');
	const consensusModel = 'anthropic/claude-3.5-sonnet';
	const consensusMessages = [
		{
			role: 'system',
			content:
				'You are a helpful assistant that analyzes multiple model responses and determines if there is a consensus.',
		},
		{
			role: 'user',
			content: `Here are responses from different models to the same user query:\n\nUser query:\n${convo.messages[convo.messages.length-2].content}\n\n${modelResults
				.map((r) => `${r.model}: ${r.response}`)
				.join(
					'\n\n'
				)}\n\nIf there is a clear consensus across all responses, answer the user's query with a summary of the consensus and do not write anything else specifically mentioning the consensus. If no, explain the key differences across responses.`,
		},
	];

	const consensusResponse = await fetch(
		`${consensusProvider.url}${consensusProvider.completionUrl}`,
		{
			method: 'POST',
			headers: {
				Authorization: `Bearer ${consensusProvider.apiKeyFn()}`,
				'Content-Type': 'application/json',
			},
			signal: get(controller).signal,
			body: JSON.stringify({
				stream: false,
				model: consensusModel,
				temperature: param.temperature,
				messages: consensusMessages,
			}),
		}
	);

	const consensusResult = await consensusResponse.json();
	onupdate({
		choices: [{ delta: { content: consensusResult.choices[0].message.content } }],
	});
}

function toolSchemaToAnthropicFormat(schema) {
	return {
		name: schema.function.name,
		description: schema.function.description,
		input_schema: schema.function.parameters,
	};
}

function messageToOpenAIFormat(msg) {
	const msgConverted = {
		role: msg.role,
	};

	if (msg.contentParts) {
		msgConverted.content = [
			{
				type: 'text',
				text: msg.content,
			},
			...msg.contentParts,
		];
	} else if (msg.role === 'tool') {
		msgConverted.content =
			typeof msg.content === 'object' ? JSON.stringify(msg.content) : msg.content;
	} else {
		msgConverted.content = msg.content;
	}

	// Additional data for tool calls
	if (msg.toolcalls) {
		msgConverted.tool_calls = msg.toolcalls.map((t) => {
			return {
				id: t.id,
				type: 'function',
				function: {
					name: t.name,
					arguments: JSON.stringify(t.arguments),
				},
			};
		});
	}
	// Additional data for tool responses
	if (msg.toolcallId && msg.name) {
		msgConverted.tool_call_id = msg.toolcallId;
		msgConverted.name = msg.name;
	}

	return msgConverted;
}

function messageToAnthropicFormat(msg) {
	const msgConverted = {
		role: msg.role === 'tool' ? 'user' : msg.role,
	};

	if (msg.contentParts) {
		msgConverted.content = [
			{
				type: 'text',
				text: msg.content,
			},
			...msg.contentParts.map((part) => {
				return {
					type: 'image',
					source: {
						type: 'base64',
						media_type: 'image/png',
						data: part.image_url.url.slice('data:image/png;base64,'.length),
					},
				};
			}),
		];
	} else if (msg.role === 'tool') {
		let content;
		if (typeof msg.content === 'object') {
			if (msg.content.contentType === 'image/png') {
				content = [
					{
						type: 'image',
						source: {
							type: 'base64',
							media_type: 'image/png',
							data: msg.content.content.slice('data:image/png;base64,'.length),
						},
					},
				];
			} else {
				content = JSON.stringify(msg.content);
			}
		} else {
			content = msg.content;
		}
		msgConverted.content = [
			{
				type: 'tool_result',
				tool_use_id: msg.toolcallId,
				content,
			},
		];
	} else {
		// Additional data for tool calls
		if (msg.toolcalls) {
			msgConverted.content = [];
			if (msg.content !== '') {
				msgConverted.content.push({ type: 'text', text: msg.content });
			}
			for (const t of msg.toolcalls) {
				msgConverted.content.push({ type: 'tool_use', id: t.id, name: t.name, input: t.arguments });
			}
		} else {
			msgConverted.content = msg.content;
		}
	}

	return msgConverted;
}

async function streamResponse(provider, readableStream, onupdate, onabort) {
	try {
		const reader = readableStream.getReader();
		const decoder = new TextDecoder();

		let done, value;
		let leftover = '';

		while (!done) {
			({ value, done } = await reader.read());

			if (done) {
				return;
			}

			const string = decoder.decode(value);
			const pairs = string.split('\n\n');

			for (let pair of pairs) {
				if (pair === '') {
					continue;
				}

				// If we have leftover from the previous chunk, prepend it to the current line
				if (leftover !== '') {
					pair = leftover + pair;
					leftover = '';
				}

				// Ignore comments
				if (pair[0] === ':') {
					continue;
				}

				// OpenAI and only OpenAI sometimes sends "\ndata:"
				pair = pair.trim();

				let event = null;
				let data = null;

				if (pair.startsWith('event:')) {
					const [eventd, datad] = pair.split('\n');
					event = eventd.substring('event: '.length);
					data = datad.substring('data: '.length);
				} else if (pair.startsWith('data: ')) {
					data = pair.substring('data: '.length);
				} else if (pair.startsWith('error: ')) {
					onupdate({ error: pair.substring('error: '.length) });
					onabort();
					return;
				} else {
					// Unknown event
					onupdate({ error: pair });
					onabort();
					return;
				}

				if (provider === 'Anthropic') {
					extractResponseAnthropic(event, data, onupdate);
				} else {
					// OpenAI-compatible:
					extractResponseOpenAI(data, onupdate, () => {
						leftover = pair;
					});
				}
			}
		}
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') {
			onabort();
		} else {
			onupdate({ error: error.message });
		}
	}
}

function extractResponseOpenAI(data, onupdate, onincomplete) {
	try {
		const parsed = JSON.parse(data);
		onupdate(parsed);
	} catch (err) {
		// If the JSON parsing fails, we've got an incomplete event
		onincomplete();
	}
}

function extractResponseAnthropic(event, data, onupdate) {
	const datap = JSON.parse(data);

	switch (event) {
		case 'message_delta':
			if (datap.delta.stop_reason) {
				const openAIReasons = {
					end_turn: 'end_turn',
					tool_use: 'tool_calls',
				};
				onupdate({
					choices: [{ delta: {}, finish_reason: openAIReasons[datap.delta.stop_reason] }],
				});
			}
			break;
		case 'content_block_start':
			if (datap.content_block.type === 'tool_use') {
				onupdate({
					choices: [
						{
							delta: {
								tool_calls: [
									{
										index: datap.index,
										id: datap.content_block.id,
										function: {
											name: datap.content_block.name,
											arguments: '',
										},
									},
								],
							},
						},
					],
				});
			}
			break;
		case 'content_block_delta':
			if (datap.delta.type === 'text_delta') {
				onupdate({
					choices: [{ delta: { content: datap.delta.text } }],
				});
			} else if (datap.delta.type === 'input_json_delta') {
				onupdate({
					choices: [
						{
							delta: {
								tool_calls: [
									{
										index: datap.index,
										function: {
											arguments: datap.delta.partial_json,
										},
									},
								],
							},
						},
					],
				});
			}
			break;
	}
}

function limitMessagesContext(messages, messagesContextLimit) {
	if (messagesContextLimit <= 0) return messages;

	const isFirstMessageSystem = messages[0]?.role === 'system';
	const systemMessage = isFirstMessageSystem ? messages[0] : null;
	const conversationMessages = isFirstMessageSystem ? messages.slice(1) : messages;

	const turns = [];
	let currentTurn = [];

	for (const message of conversationMessages) {
		if (message.role === 'user' && currentTurn.length > 0) {
			turns.push(currentTurn);
			currentTurn = [];
		}
		currentTurn.push(message);
	}
	if (currentTurn.length > 0) {
		turns.push(currentTurn);
	}

	const limitedTurns = turns.slice(-messagesContextLimit);
	const limitedMessages = limitedTurns.flat();

	if (systemMessage) {
		limitedMessages.unshift(systemMessage);
	}

	return limitedMessages;
}

export async function generateImage(convo, { oncomplete }) {
	const provider = providers.find((p) => p.name === convo.model.provider);
	const userMessages = convo.messages.filter((msg) => msg.role === 'user');
	const lastMessage = userMessages[userMessages.length - 1];

	const resp = await fetch(`${provider.url}/v1/images/generations`, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${provider.apiKeyFn()}`,
			'HTTP-Referer': 'https://lluminous.chat',
			'X-Title': 'lluminous',
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			model: convo.model.id,
			prompt: lastMessage.content,
			n: 1,
			size: '1024x1024',
		}),
	});
	const json = await resp.json();
	oncomplete(json);
}

export function conversationToString(convo) {
	let result = '';
	convo.messages.forEach((msg) => {
		result += messageToString(msg, convo.model.template);
	});
	return result;
}

function conversationStop(convo) {
	switch (convo.model.template) {
		case 'chatml':
			return ['<|im_end|>', '<|im_start|>', '</tool_call>'];
		case 'deepseek':
			return ['### Instruction:', '### Response:'];
		case 'none':
			return ['</s>'];
		default:
			throw new Error('Unknown template');
	}
}

function messageToString(message, template) {
	switch (template) {
		case 'chatml':
			let s = '<|im_start|>' + message.role + '\n' + message.content;
			if (!message.unclosed) {
				s += '<|im_end|>\n';
			}
			return s;
		case 'deepseek':
			if (message.role === 'system') {
				return message.content + '\n';
			}
			if (message.role === 'user') {
				return '### Instruction:\n' + message.content + '\n';
			}
			if (message.role === 'assistant') {
				return '### Response:\n' + message.content + '\n';
			}
		case 'none':
			return message.content;
	}
}
