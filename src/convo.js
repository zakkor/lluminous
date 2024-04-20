import { get } from 'svelte/store';
import { toolSchema } from './stores.js';
import { providers } from './providers.js';

export function conversationToString(convo) {
	let result = '';
	convo.messages.forEach((msg) => {
		result += messageToString(msg, convo.tmpl);
	});
	return result;
}

function conversationStop(convo) {
	switch (convo.tmpl) {
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

export async function complete(convo, onupdate, onabort) {
	convo.controller = new AbortController();

	let response;
	if (convo.local) {
		response = await fetch('http://localhost:8080/completion', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			signal: convo.controller.signal,
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
	} else {
		const messages = convo.messages.map((msg) => {
			const msgOAI = {
				role: msg.role,
				content: msg.content,
			};
			// Additional data for tool calls
			if (msg.toolcall) {
				msgOAI.tool_calls = [
					{
						id: msg.toolcall.id,
						type: 'function',
						function: {
							name: msg.toolcall.name,
							arguments: JSON.stringify(msg.toolcall.arguments),
						},
					},
				];
			}
			// Additional data for tool responses
			if (msg.tool_call_id) {
				msgOAI.tool_call_id = msg.tool_call_id;
			}
			return msgOAI;
		});

		// Filter out unclosed messages from being submitted if using external models
		if (convo.messages[convo.messages.length - 1].unclosed) {
			messages.pop();
		}

		const schema = get(toolSchema);
		const provider = providers.find((p) => p.name === convo.model.provider);
		response = await fetch(`${provider.url}/v1/chat/completions`, {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${provider.apiKeyFn()}`,
				'HTTP-Referer': 'https://lluminous.chat',
				'X-Title': 'lluminous',
				'Content-Type': 'application/json',
			},
			signal: convo.controller.signal,
			body: JSON.stringify({
				stream: true,
				model: convo.model.id,
				temperature: 0,
				tools: schema.length > 0 ? schema : undefined,
				messages,
			}),
		});
	}

	streamResponse(response.body, onupdate, onabort);
}

async function streamResponse(readableStream, onupdate, onabort) {
	try {
		const reader = readableStream.getReader();

		const decoder = new TextDecoder();
		let done, value;
		while (!done) {
			({ value, done } = await reader.read());

			if (done) {
				return;
			}

			const decoded = decoder.decode(value);
			const lines = decoded.split('\n\n');
			for (const line of lines) {
				if (line === '') {
					continue;
				}
				if (line.startsWith('data: ')) {
					// Strip "data: " from the start of the url
					const stripped = line.substring(6);
					// OpenAI-compatible APIs send "data: [DONE]" at the end of the stream
					if (stripped === '[DONE]') {
						onabort();
						return;
					}

					onupdate(JSON.parse(stripped));
				} else if (line.startsWith('error: ')) {
					console.error('received error event:', line);
					onabort();
					return;
				} else {
					console.warn('received unknown event:', decoded);
				}
			}
		}
	} catch (error) {
		if (error instanceof DOMException && error.name === 'AbortError') {
			onabort();
		} else {
			console.log(`error.name:`, error.name);
			throw error;
		}
	}
}
