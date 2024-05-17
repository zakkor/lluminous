<script>
	import { onMount, tick } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import {
		complete,
		conversationToString,
		formatModelName,
		hasCompanyLogo,
		additionalModelsMultimodal,
		readFileAsDataURL,
	} from './convo.js';
	import KnobsSidebar from './KnobsSidebar.svelte';
	import Button from './Button.svelte';
	import {
		faArrowUp,
		faArrowUpRightFromSquare,
		faArrowsRotate,
		faBarsStaggered,
		faCheck,
		faCheckDouble,
		faChevronDown,
		faChevronLeft,
		faChevronRight,
		faCircleNotch,
		faEllipsis,
		faGear,
		faPaperclip,
		faPen,
		faPlus,
		faSliders,
		faStop,
		faVial,
		faXmark,
	} from '@fortawesome/free-solid-svg-icons';
	import { faLightbulb, faTrashCan } from '@fortawesome/free-regular-svg-icons';
	import { marked } from 'marked';
	import markedKatex from './marked-katex-extension';
	import Markdown from './svelte-marked/markdown/Markdown.svelte';
	import JsonView from './svelte-json-view/JsonView.svelte';

	import Icon from './Icon.svelte';
	import { persisted } from './indexeddb.js';
	import { getRelativeDate } from './date.js';
	import { compressAndEncode, decodeAndDecompress } from './share.js';
	import { providers } from './providers.js';
	import ModelSelector from './ModelSelector.svelte';
	import CompanyLogo from './CompanyLogo.svelte';
	import {
		pick,
		controller,
		remoteServer,
		openaiAPIKey,
		groqAPIKey,
		openrouterAPIKey,
	} from './stores.js';
	import { writable } from 'svelte/store';
	import SettingsModal from './SettingsModal.svelte';

	marked.use(
		markedKatex({
			throwOnError: false,
		})
	);

	// Initialize history and convo with a blank slate while we wait for IndexedDB to start
	let history = writable({
		convoId: null,
		entries: {},
	});
	let convo = writable({
		model: {},
		messages: [],
		// Set initial conversation `shared` to true, to prevent the SettingsModal from popping up on first render.
		shared: true,
	});

	$: isMultimodal =
		$convo.model.modality === 'multimodal' || additionalModelsMultimodal.includes($convo.model.id);

	let historyBuckets = [];
	$: {
		historyBuckets = [];
		for (const entry of Object.values($history.entries).sort((a, b) => b.id - a.id)) {
			if (entry.shared || isNaN(new Date(entry.id).getTime())) {
				continue;
			}

			const bucketKey = getRelativeDate(entry.id);

			const existingBucket = historyBuckets.find((bucket) => bucket.relativeDate === bucketKey);
			if (!existingBucket) {
				historyBuckets.push({ relativeDate: bucketKey, convos: [entry] });
			} else {
				existingBucket.convos.push(entry);
			}
		}
		historyBuckets.sort((a, b) => b.convos[0].id - a.convos[0].id);
	}

	let content = '';
	let imageUrls = [];
	let imageUrlsBlacklist = [];
	const imageUrlRegex = /https?:\/\/[^\s]+?\.(png|jpe?g)(?=\s|$)/gi;
	let generating = false;

	let currentTokens = 0;
	let totalTokens = 0;
	let hidingTokenCount = false;

	let historyOpen = false;
	let knobsOpen = false;

	let scrollableEl = null;
	let textareaEls = [];
	let inputTextareaEl;
	let fileInputEl;

	function submitEdit(i) {
		const message = $convo.messages[i];
		if (message.submitted || message.generated) {
			saveVersion(message, i);
		}

		$convo.messages = $convo.messages.slice(0, i + 1);

		submitCompletion();
	}

	async function submitCompletion(insertUnclosed = true) {
		if (!$convo.model.provider) {
			$convo.messages.push({
				id: Date.now(),
				role: 'assistant',
				error: 'No model selected. Please add at least one API key and select a model to begin.',
				content: '',
			});
			$convo.messages = $convo.messages;
			return;
		}

		if (generating) {
			$controller.abort();
		}

		generating = true;

		if ($convo.model.provider === 'Local') {
			totalTokens = await tokenizeCount(conversationToString($convo));
		}

		if (insertUnclosed) {
			$convo.messages.push({
				id: Date.now(),
				role: 'assistant',
				content: '',
				unclosed: true,
				generated: true,
				model: $convo.model,
			});
			$convo.messages = $convo.messages;
		}

		for (let i = 0; i < $convo.messages.length; i++) {
			if ($convo.messages[i].editing) {
				$convo.messages[i].content = $convo.messages[i].pendingContent;
			}
			$convo.messages[i].pendingContent = '';
			$convo.messages[i].editing = false;
			$convo.messages[i].submitted = true;
		}
		await tick();
		scrollableEl.scrollTop = scrollableEl.scrollHeight;

		const i = $convo.messages.length - 1;

		const onupdate = async (chunk) => {
			if (chunk.error) {
				$convo.messages[i].error = chunk.error.message || chunk.error;
				generating = false;
				return;
			}

			if ($convo.model.provider === 'Local') {
				$convo.messages[i].content += chunk.content;
				totalTokens = await tokenizeCount(conversationToString($convo));
			} else {
				if (chunk.choices.length === 0) {
					$convo.messages[i].error = 'Refused to respond';
					generating = false;
					return;
				}

				const choice = chunk.choices[0];

				if (choice.delta.content) {
					$convo.messages[i].content += choice.delta.content;
				}

				if (choice.delta.tool_calls) {
					if (!$convo.messages[i].toolcalls) {
						$convo.messages[i].toolcalls = [];
					}

					for (const tool_call of choice.delta.tool_calls) {
						const index = tool_call.index;
						if (!$convo.messages[i].toolcalls[index]) {
							$convo.messages[i].toolcalls[index] = {
								id: tool_call.id,
								name: tool_call.function.name,
								arguments: '',
								expanded: true,
							};
						} else {
							$convo.messages[i].toolcalls[index].arguments += tool_call.function.arguments;
						}
					}
				}
			}

			// Check for stoppage:
			// For local models, `.stop` will be true.
			// For external models, `.choices[0].finish_reason` will be 'stop'.
			if ($convo.model.provider === 'Local' && chunk.stop) {
				generating = false;
				return;
			}

			// External tool calls:
			if (
				$convo.model.provider !== 'Local' &&
				chunk.choices &&
				(chunk.choices[0].finish_reason === 'stop' ||
					chunk.choices[0].finish_reason === 'tool_calls')
			) {
				generating = false;

				// Toolcall arguments are now finalized, we can parse them:
				if ($convo.messages[i].toolcalls) {
					let toolPromises = [];

					for (let ti = 0; ti < $convo.messages[i].toolcalls.length; ti++) {
						const toolcall = $convo.messages[i].toolcalls[ti];

						toolcall.arguments = JSON.parse(toolcall.arguments);

						// Call the tool
						const promise = fetch(`${$remoteServer.address}/tool`, {
							method: 'POST',
							headers: {
								Authorization: `Basic ${$remoteServer.password}`,
							},
							body: JSON.stringify(toolcall),
						}).then((resp) => {
							// Mark tool call as finished to we can display it nicely in the UI
							// (still need to await all tool calls to deliver the final response).
							$convo.messages[i].toolcalls[ti].finished = true;

							return resp.text();
						});

						toolPromises.push(promise);
					}

					const toolResponses = await Promise.all(toolPromises);

					for (let ti = 0; ti < toolResponses.length; ti++) {
						$convo.messages.push({
							id: Date.now(),
							role: 'tool',
							tool_call_id: $convo.messages[i].toolcalls[ti].id,
							name: $convo.messages[i].toolcalls[ti].name,
							content: toolResponses[ti],
						});
						$convo.messages = $convo.messages;
					}

					submitCompletion();
				}

				return;
			}
		};

		const onabort = () => {
			generating = false;
		};

		const ondirect = async (chatResp) => {
			if ($convo.model.provider === 'Local') {
				// TODO:
				return;
			}

			const choice = chatResp.choices[0];

			if (choice.message.content) {
				$convo.messages[i].content = choice.message.content;
			}

			let toolPromises = [];
			if (choice.message.tool_calls) {
				for (let ti = 0; ti < choice.message.tool_calls.length; ti++) {
					const toolcall = choice.message.tool_calls[ti];

					if (!$convo.messages[i].toolcalls) {
						$convo.messages[i].toolcalls = [];
					}

					$convo.messages[i].toolcalls[ti] = {
						id: toolcall.id,
						name: toolcall.function.name,
						arguments: JSON.parse(toolcall.function.arguments),
						expanded: true,
					};

					// Call the tool
					const promise = fetch(`${$remoteServer.address}/tool`, {
						method: 'POST',
						headers: {
							Authorization: `Basic ${$remoteServer.password}`,
						},
						body: JSON.stringify({
							name: $convo.messages[i].toolcalls[ti].name,
							arguments: $convo.messages[i].toolcalls[ti].arguments,
						}),
					}).then((resp) => {
						// Mark tool call as finished to we can display it nicely in the UI
						// (still need to await all tool calls to deliver the final response).
						$convo.messages[i].toolcalls[ti].finished = true;

						return resp.text();
					});

					toolPromises.push(promise);
				}

				const toolResponses = await Promise.all(toolPromises);

				for (let ti = 0; ti < toolResponses.length; ti++) {
					const toolcall = choice.message.tool_calls[ti];
					$convo.messages.push({
						id: Date.now(),
						role: 'tool',
						tool_call_id: toolcall.id,
						name: toolcall.function.name,
						content: toolResponses[ti],
					});
					$convo.messages = $convo.messages;
				}

				submitCompletion();
			}

			generating = false;
		};

		complete($convo, onupdate, onabort, ondirect);
	}

	async function tokenizeCount(content) {
		const tokenizeResp = await fetch(`${$remoteServer.address}/tokenize_count`, {
			method: 'POST',
			headers: {
				Authorization: `Basic ${$remoteServer.password}`,
			},
			body: JSON.stringify({ content }),
		});
		return parseInt(await tokenizeResp.text());
	}

	async function insertSystemPrompt() {
		$convo.messages.unshift({ role: 'system', content: '', editing: true });
		$convo.messages = $convo.messages;
		await tick();
		textareaEls[0].focus();
	}

	function autoresizeTextarea() {
		inputTextareaEl.style.height = 'auto';
		if (window.innerWidth > 880) {
			inputTextareaEl.style.height = Math.max(74, inputTextareaEl.scrollHeight + 2) + 'px';
		} else {
			inputTextareaEl.style.height = inputTextareaEl.scrollHeight + 2 + 'px';
		}
	}

	async function sendMessage() {
		if (content.length > 0) {
			const msg = {
				id: Date.now(),
				role: 'user',
				content: content,
				submitted: true,
			};

			const imageUrlMapper = (url) => ({
				type: 'image_url',
				image_url: {
					url,
					detail: 'low',
				},
			});

			if (imageUrls.length > 0) {
				msg.contentParts = [...imageUrls.map(imageUrlMapper)];
			}

			$convo.messages.push(msg);
			$convo.messages = $convo.messages;
			await tick();
			scrollableEl.scrollTop = scrollableEl.scrollHeight;

			content = '';
			imageUrls = [];
			imageUrlsBlacklist = [];
			currentTokens = 0;

			await tick();
			if (window.innerWidth < 880) {
				inputTextareaEl.blur();
			}
			autoresizeTextarea();

			submitCompletion();
		}
	}

	function cleanShareLink() {
		const params = new URLSearchParams(window.location.search);
		if (params.has('s') || params.has('sl')) {
			window.history.pushState('', document.title, window.location.pathname);
		}
	}

	function newConversation() {
		cleanShareLink();

		if ($convo.messages.length === 0) {
			historyOpen = false;
			inputTextareaEl.focus();
			return;
		}

		const existingNewConvo = Object.values($history.entries).find(
			(convo) => convo.messages.length === 0
		);
		if (existingNewConvo) {
			const oldModel = $convo.model;
			$history.convoId = existingNewConvo.id;
			convo = pick(history, (h) => h.entries[h.convoId]);
			$convo.model = oldModel;
			historyOpen = false;
			inputTextareaEl.focus();
			return;
		}

		const convoData = {
			id: Date.now(),
			model: $convo.model || models.find((m) => m.id === 'meta-llama/llama-3-8b-instruct'),
			messages: [],
		};
		$history.convoId = convoData.id;
		$history.entries[convoData.id] = convoData;
		convo = pick(history, (h) => h.entries[h.convoId]);

		historyOpen = false;

		inputTextareaEl.focus();
	}

	// Split history at this point:
	function saveVersion(message, i) {
		if (!$convo.versions) {
			$convo.versions = {};
		}
		if (!$convo.versions[message.id]) {
			$convo.versions[message.id] = [null];
		}
		const nullIdx = $convo.versions[message.id].findIndex((v) => v === null);
		$convo.versions[message.id][nullIdx] = structuredClone($convo.messages.slice(i)).map((m) => {
			return {
				...m,
				editing: false,
				pendingContent: '',
			};
		});
		$convo.versions[message.id].push(null);
	}

	function shiftVersion(dir, message, i) {
		const activeVersionIndex = $convo.versions[message.id].findIndex((v) => v === null);
		const newVersionIndex = activeVersionIndex + dir;

		$convo.versions[message.id][activeVersionIndex] = $convo.messages.slice(i);

		const newMessages = $convo.versions[message.id][newVersionIndex];

		$convo.messages = $convo.messages.slice(0, i).concat(newMessages);

		$convo.versions[message.id][newVersionIndex] = null;
	}

	function closeSidebars(event) {
		if (
			historyOpen &&
			!event.target.closest('[data-sidebar="history"]') &&
			!event.target.closest('[data-trigger="history"]')
		) {
			historyOpen = false;
		}
		if (
			knobsOpen &&
			!event.target.closest('[data-trigger="knobs"]') &&
			!event.target.closest('[data-sidebar="knobs"]')
		) {
			knobsOpen = false;
		}
	}

	async function shareConversation() {
		const sharePromise = new Promise(async (resolve) => {
			const encoded = await compressAndEncode({
				model: $convo.model,
				messages: $convo.messages,
			});
			const share = `${window.location.protocol}//${window.location.host}/?s=${encoded}`;
			if (share.length > 200) {
				const data = new FormData();
				data.append('pwd', 'muie_webshiti');
				data.append('f:1', new Blob([encoded], { type: 'text/plain' }), 'content.txt');
				const response = await fetch(`https://zak.oni2025.ro`, {
					method: 'POST',
					body: data,
				});

				const shortenedLink = await response.text();
				resolve(
					`${window.location.protocol}//${window.location.host}/?sl=${shortenedLink.slice(25, shortenedLink.length - 13)}`
				);
			} else {
				resolve(`${window.location.protocol}//${window.location.host}/?s=${encoded}`);
			}
		});

		try {
			const clipboardItem = new ClipboardItem({
				'text/plain': sharePromise,
			});
			await navigator.clipboard.write([clipboardItem]);
		} catch (err) {
			await navigator.clipboard.writeText(await sharePromise);
		}
	}

	async function restoreConversation() {
		const params = new URLSearchParams(window.location.search);
		let share;
		if (params.has('s')) {
			share = params.get('s');
		} else if (params.has('sl')) {
			const response = await fetch(`https://zak.oni2025.ro/p/${params.get('sl')}/`);
			share = await response.text();
		}
		if (!share) {
			return;
		}

		decodeAndDecompress(share)
			.then((decoded) => {
				if (Array.isArray(decoded)) {
					decoded = { name: 'Shared conversation', messages: decoded };
				}
				let id = Date.now();
				const existingShared = Object.values($history.entries).find((convo) => convo.shared);
				if (existingShared) {
					id = existingShared.id;
				}
				const convoData = {
					id,
					shared: true,
					model: decoded.model,
					messages: decoded.messages,
				};
				$history.convoId = convoData.id;
				$history.entries[convoData.id] = convoData;
				convo = pick(history, (h) => h.entries[h.convoId]);
			})
			.catch((err) => {
				console.error('Error decoding shared conversation:', err);
			});
	}

	let loading = false;

	async function loadModel(newModel) {
		if (newModel.provider === 'Local') {
			loading = true;

			// For local models, we need to tell the server to load them:
			// FIXME: local
			await fetch(`${$remoteServer.address}/model`, {
				method: 'POST',
				headers: {
					Authorization: `Basic ${$remoteServer.password}`,
				},
				body: JSON.stringify({
					model: newModel,
				}),
			});

			loading = false;
		}

		setModel(newModel);
	}

	function setModel(newModel) {
		$convo.model = newModel;
	}

	let models = [];
	let loadingModels = false;

	async function fetchModels() {
		loadingModels = true;
		try {
			const promises = providers.map((provider) => {
				if (!provider.apiKeyFn()) {
					return [];
				}
				return fetch(`${provider.url}/v1/models`, {
					method: 'GET',
					headers: {
						Authorization: `Bearer ${provider.apiKeyFn()}`,
					},
				})
					.then((response) => response.json())
					.then((json) => {
						const externalModels = json.data.map((m) => ({
							id: m.id,
							name: m.name || m.id,
							provider: provider.name,
							modality: m.architecture?.modality,
						}));
						return externalModels;
					})
					.catch(() => {
						console.log('Error fetching models from provider', provider.name);
						return [];
					});
			});

			const ignoreIds = [
				'dall-e-3',
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

			const results = await Promise.all(promises);
			const externalModels = results.flat().filter((m) => !ignoreIds.includes(m.id));

			const priorityOrder = [
				{ exactly: ['gpt-4o'] },
				{ fromProvider: 'OpenAI' },
				{ exactly: ['openai/gpt-4o', 'openai/gpt-4-turbo', 'openai/gpt-3.5-turbo'] },
				{
					exactly: [
						'anthropic/claude-3-opus',
						'anthropic/claude-3-sonnet',
						'anthropic/claude-3-haiku',
					],
				},
				{ fromProvider: 'Groq', exactlyNot: ['llama2-70b-4096', 'gemma-7b-it'] },
				{ exactly: ['meta-llama/llama-3-70b-instruct', 'meta-llama/llama-3-8b-instruct'] },
				{
					exactly: [
						'perplexity/llama-3-sonar-large-32k-online',
						'perplexity/llama-3-sonar-small-32k-online',
					],
				},
				{ exactly: ['google/gemini-flash-1.5', 'google/gemini-pro-1.5'] },
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
				{ startsWith: ['mistralai/', 'cohere/', 'nous'] },
			];

			function getPriorityIndex(model) {
				for (let i = 0; i < priorityOrder.length; i++) {
					const rule = priorityOrder[i];
					if (rule.exactly) {
						const exactIndex = rule.exactly.indexOf(model.id);
						if (exactIndex !== -1) {
							return [i, exactIndex];
						}
					}
					if (rule.startsWith) {
						for (let j = 0; j < rule.startsWith.length; j++) {
							if (model.id.startsWith(rule.startsWith[j])) {
								if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
									continue;
								}
								return [i, j];
							}
						}
					}
					if (rule.fromProvider && model.provider === rule.fromProvider) {
						if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
							continue;
						}
						return [i, -1];
					}
				}
				return [priorityOrder.length, -1];
			}

			externalModels.sort((a, b) => {
				const [aIndex, aExactIndex] = getPriorityIndex(a);
				const [bIndex, bExactIndex] = getPriorityIndex(b);

				if (aIndex === bIndex) {
					if (aExactIndex === bExactIndex) {
						return a.id.localeCompare(b.id);
					}
					return aExactIndex - bExactIndex;
				}
				return aIndex - bIndex;
			});

			models = externalModels;
		} catch (error) {
			console.error('Error:', error);
		} finally {
			loadingModels = false;
		}
	}

	function initializePWAStyles() {
		if (window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone) {
			document.body.classList.add('standalone');
		}
	}

	onMount(async () => {
		persisted('history', {
			convoId: null,
			entries: {},
		}).then((store) => {
			history = store;

			if (!$history.convoId) {
				const convoData = {
					id: Date.now(),
					model: { id: null, name: 'No model loaded', provider: null },
					messages: [],
				};
				$history.convoId = convoData.id;
				$history.entries[convoData.id] = convoData;
			}

			convo = pick(history, (h) => h.entries[h.convoId]);
		});

		initializePWAStyles();

		await restoreConversation();

		await fetchModels();
	});
</script>

<svelte:window
	on:touchstart={closeSidebars}
	on:click={closeSidebars}
	on:keydown={(event) => {
		if (
			event.key === 'Escape' &&
			generating &&
			$convo.messages.filter((msg) => msg.generated).length > 0
		) {
			$controller.abort();
		}
	}}
/>

<main class="flex h-dvh w-screen flex-col">
	<div class="flex items-center border-b border-slate-200 px-2 py-1 md:hidden">
		<button
			on:click={newConversation}
			class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
		>
			<Icon icon={faPlus} class="ml-auto h-4 w-4 text-slate-700" />
		</button>
		<button
			data-trigger="history"
			class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={() => (historyOpen = !historyOpen)}
		>
			<Icon icon={faBarsStaggered} class="m-auto h-4 w-4 text-slate-700" />
		</button>

		{#if !$convo.shared}
			<ModelSelector {convo} {models} class="!absolute left-1/2 z-[99] -translate-x-1/2" />
		{:else if $convo.model}
			<p
				class="!absolute left-1/2 line-clamp-1 flex -translate-x-1/2 items-center gap-x-2 whitespace-nowrap text-sm font-semibold"
			>
				<CompanyLogo model={$convo.model} size="w-4 h-4" />
				{formatModelName($convo.model)}
			</p>
		{/if}

		<button
			class="ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={shareConversation}
		>
			<Icon icon={faArrowUpRightFromSquare} class="m-auto h-4 w-4 text-slate-700" />
		</button>
		<button
			data-trigger="knobs"
			class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={() => (knobsOpen = !knobsOpen)}
		>
			<Icon icon={faSliders} class="m-auto h-4 w-4 text-slate-700" />
		</button>
	</div>
	<div class="relative flex h-full flex-1 overflow-hidden">
		<aside
			data-sidebar="history"
			class="{historyOpen
				? ''
				: '-translate-x-full'} fixed top-0 z-[100] flex h-full w-[230px] flex-col border-r bg-white pl-3 pt-4 transition-transform duration-300 ease-in-out md:static md:translate-x-0"
		>
			<div class="mb-1 pr-3">
				<button
					on:click={newConversation}
					class="flex w-full items-center rounded-lg border py-2.5 pl-3 pr-4 text-left text-sm font-medium hover:bg-gray-100"
				>
					New chat
					<Icon icon={faPlus} class="ml-auto h-3.5 w-3.5 text-slate-700" />
				</button>
			</div>
			<ol
				class="flex list-none flex-col overflow-y-auto pb-3 pr-3 pt-5 !scrollbar-white scrollbar-slim hover:!scrollbar-slim"
			>
				{#each historyBuckets as { relativeDate, convos } (relativeDate)}
					<li class="mb-2 ml-3 text-xs font-medium text-slate-600 [&:not(:first-child)]:mt-6">
						{relativeDate}
					</li>
					{#each convos as convo (convo.id)}
						<li class="group relative">
							<button
								on:click={() => {
									if (generating) {
										$controller.abort();
									}

									historyOpen = false;

									$history.convoId = convo.id;
									convo = pick(history, (h) => h.entries[h.convoId]);

									cleanShareLink();
								}}
								class="{$history.convoId === convo.id
									? 'bg-gray-100'
									: ''} leading-0 w-full rounded-lg px-3 py-2 text-left text-sm group-hover:bg-gray-100"
							>
								<span class="line-clamp-1">
									{convo.messages.length === 0
										? 'New conversation'
										: convo.messages[0].content.split(' ').slice(0, 5).join(' ')}
								</span>
							</button>
							<button
								on:click={() => {
									if ($history.convoId === convo.id) {
										const newId = Object.values($history.entries).find(
											(e) => e.id !== convo.id && !e.shared
										)?.id;
										if (!newId) {
											newConversation();
										} else {
											$history.convoId = newId;
										}
									}
									delete $history.entries[convo.id];
									$history.entries = $history.entries;
								}}
								class="z-1 absolute right-0 top-0 flex h-full w-12 rounded-br-lg rounded-tr-lg bg-gradient-to-l {$history.convoId ===
								convo.id
									? 'from-gray-100'
									: 'from-white group-hover:from-gray-100'} from-65% to-transparent pr-3 transition-opacity sm:from-gray-100 sm:opacity-0 sm:group-hover:opacity-100"
							>
								<Icon icon={faTrashCan} class="m-auto mr-0 h-3 w-3 shrink-0 text-slate-700" />
							</button>
						</li>
					{/each}
				{/each}
			</ol>

			<div class="-ml-3 mt-auto flex pb-3">
				<button
					data-trigger="settings"
					class="mx-3 flex flex-1 items-center gap-x-4 rounded-lg border border-slate-200 px-4 py-3 text-left text-sm font-medium hover:bg-gray-100"
					on:click={() => {
						historyOpen = false;
					}}
				>
					Settings
					<Icon icon={faGear} class="ml-auto h-4 w-4 text-slate-700" />
				</button>
			</div>
		</aside>
		<div class="flex flex-1 flex-col">
			<div class="relative hidden items-center border-b border-slate-200 px-2 py-1 md:flex">
				{#if !$convo.shared}
					<ModelSelector {convo} {models} class="!absolute left-1/2 z-[99] -translate-x-1/2" />
				{:else if $convo.model}
					<p
						class="!absolute left-1/2 flex -translate-x-1/2 items-center gap-x-2 text-sm font-semibold"
					>
						<CompanyLogo model={$convo.model} size="w-4 h-4" />
						{formatModelName($convo.model)}
					</p>
				{/if}

				<button
					class="ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
					on:click={shareConversation}
				>
					<Icon icon={faArrowUpRightFromSquare} class="m-auto h-4 w-4 text-slate-700" />
				</button>
				<button
					data-trigger="knobs"
					class="flex rounded-full p-3 transition-colors hover:bg-gray-100 xl:hidden"
					on:click={() => (knobsOpen = !knobsOpen)}
				>
					<Icon icon={faSliders} class="m-auto h-4 w-4 text-slate-700" />
				</button>
			</div>
			<section
				bind:this={scrollableEl}
				class="scrollable flex h-full w-full flex-col overflow-y-auto pb-[130px] scrollbar-slim md:pb-[150px]"
				on:scroll={() => {
					if (
						scrollableEl.scrollTop + scrollableEl.clientHeight >=
						scrollableEl.scrollHeight - 100
					) {
						hidingTokenCount = false;
					} else {
						hidingTokenCount = true;
					}
				}}
			>
				{#if $convo.messages.length > 0}
					<ul
						class="mb-3 flex w-full !list-none flex-col divide-y divide-slate-200/50 border-b border-slate-200/50"
					>
						{#each $convo.messages as message, i}
							{#if ['system', 'user', 'assistant'].includes(message.role)}
								{@const hasLogo = hasCompanyLogo(message.model)}
								<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
								<li
									data-role={message.role}
									class="{!message.generated &&
									!message.submitted &&
									message.role !== 'system' &&
									message.role !== 'tool'
										? 'bg-yellow-50/40'
										: message.role === 'assistant'
											? 'bg-slate-50/30'
											: ''} group relative px-5 pb-10 pt-6 ld:px-8"
									style="z-index: {$convo.messages.length - i};"
									on:touchstart={(event) => {
										// Make click trigger hover on mobile:
										event.target.dispatchEvent(new MouseEvent('mouseenter'));
									}}
								>
									{#if i === 0 && message.role !== 'system'}
										<Button
											variant="outline"
											class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 rounded-t-none !border-t-0 border-dashed text-xs opacity-0 transition-opacity group-hover:opacity-100"
											on:click={insertSystemPrompt}
										>
											<Icon icon={faVial} class="mr-2 h-3 w-3 text-slate-600" />
											Add system prompt
										</Button>
									{/if}
									<div
										class="relative mx-auto flex w-full max-w-[680px] gap-x-3.5 self-start md:gap-x-5 ld:max-w-[768px]"
									>
										<button
											disabled={message.role === 'system'}
											on:click={() => {
												// Toggle between user and assistant:
												if (message.role === 'user') {
													message.role = 'assistant';
												} else {
													message.role = 'user';
												}
											}}
											class="flex h-8 w-8 shrink-0 rounded-md md:h-9 md:w-9 md:rounded-[7px] {message.role ===
											'system'
												? 'bg-blue-200'
												: message.role === 'user'
													? 'bg-red-200'
													: message.role === 'assistant' && !hasLogo
														? 'bg-teal-200'
														: ''}"
										>
											{#if message.role === 'assistant' && hasLogo}
												<CompanyLogo
													model={message.model}
													size="w-full h-full"
													rounded="rounded-[inherit]"
												/>
											{:else}
												<span class="m-auto text-base">
													{#if message.role === 'system'}
														S
													{:else if message.role === 'assistant'}
														A
													{:else}
														U
													{/if}
												</span>
											{/if}
										</button>

										{#if generating && message.role === 'assistant' && i === $convo.messages.length - 1 && message.content === '' && !message.toolcalls}
											<div
												class="mt-2 h-3 w-3 shrink-0 animate-pulse rounded-full bg-slate-700/50"
											/>
										{/if}
										<!-- svelte-ignore a11y-no-static-element-interactions -->
										{#if message.editing}
											<textarea
												bind:this={textareaEls[i]}
												class="w-full resize-none rounded-lg border-none bg-transparent p-0 leading-[28px] text-slate-800 outline-none focus:ring-0"
												rows={1}
												bind:value={message.pendingContent}
												on:keydown={(event) => {
													if (
														event.key === 'Enter' &&
														!event.shiftKey &&
														message.role === 'user' &&
														message.submitted &&
														message.pendingContent &&
														message.content !== message.pendingContent
													) {
														event.preventDefault();
														submitEdit(i);
														event.target.blur();
													}
												}}
												on:input={(event) => {
													// Resize textarea as content grows:
													event.target.style.height = 'auto';
													event.target.style.height = event.target.scrollHeight + 'px';
												}}
											/>
										{:else}
											<div class="flex w-full flex-col gap-6">
												{#if message.error}
													<span class="text-slate-600">{message.error}</span>
												{:else if message.content}
													<div
														class="markdown prose prose-slate flex w-full max-w-none flex-col break-words prose-p:whitespace-pre-wrap prose-p:text-slate-800 prose-a:break-all prose-code:break-all prose-pre:my-0 prose-pre:whitespace-pre-wrap prose-pre:break-all prose-pre:border prose-pre:border-slate-200 prose-pre:bg-white prose-pre:text-slate-800 prose-img:mb-2"
													>
														{#if message.contentParts}
															{#each message.contentParts as part}
																<img
																	src={part.image_url.url}
																	alt=""
																	class="max-h-[400px] w-min rounded-lg"
																/>
															{/each}
														{/if}
														<Markdown source={message.content} />
													</div>
												{/if}

												<!-- OAI toolcalls will always be at the end -->
												{#if message.toolcalls}
													{#each message.toolcalls as toolcall, ti}
														{@const toolResponse = $convo.messages.find(
															(msg) => msg.tool_call_id === toolcall.id
														)}
														{@const finished = toolcall.finished || toolResponse}
														<div class="mb-1 flex w-full flex-col bg-white">
															<button
																class="{toolcall.expanded
																	? ''
																	: 'rounded-b-lg'} flex items-center gap-3 rounded-t-lg border border-slate-200 py-3 pl-4 pr-5 text-sm text-slate-700 transition-colors hover:bg-gray-50"
																on:click={() => {
																	$convo.messages[i].toolcalls[ti].expanded =
																		!$convo.messages[i].toolcalls[ti].expanded;
																}}
															>
																{#key finished}
																	<span in:fade={{ duration: 300 }}>
																		<Icon
																			icon={finished ? faCheck : faCircleNotch}
																			class="{finished
																				? ''
																				: 'animate-spin'} h-4 w-4 text-slate-700"
																		/>
																	</span>
																{/key}
																<span>
																	{finished ? 'Used' : 'Using'} tool:
																	<code class="ml-1 font-semibold">{toolcall.name}</code>
																</span>
																<Icon
																	icon={faChevronDown}
																	class="{toolcall.expanded
																		? 'rotate-180'
																		: ''} ml-auto h-3 w-3 text-slate-700 transition-transform"
																/>
															</button>
															{#if toolcall.expanded}
																<div transition:slide={{ duration: 300 }}>
																	<div
																		class="{toolResponse
																			? 'border-b-0'
																			: 'rounded-b-lg'} whitespace-pre-wrap break-all border border-t-0 border-slate-200 bg-white px-4 py-3 font-mono text-sm text-slate-800"
																	>
																		{#if typeof toolcall.arguments === 'object'}
																			{#if Object.keys(toolcall.arguments).length === 1}
																				{toolcall.arguments[Object.keys(toolcall.arguments)[0]]}
																			{:else}
																				<JsonView json={toolcall.arguments} />
																			{/if}
																		{:else}
																			{toolcall.arguments}
																		{/if}
																	</div>
																	{#if toolResponse}
																		<div
																			class="h-px w-full border-t border-dashed border-slate-300"
																		/>
																		<div
																			class="flex flex-col rounded-b-lg border border-t-0 border-slate-200"
																		>
																			<span
																				class="px-4 pt-3 text-sm font-medium tracking-[0.01em] text-slate-700"
																				>Result:</span
																			>
																			<div
																				class="whitespace-pre-wrap break-all rounded-[inherit] bg-white px-4 py-3 font-mono text-sm text-slate-800"
																			>
																				{#if toolResponse.content}
																					{toolResponse.content}
																				{:else}
																					<span class="italic">blank</span>
																				{/if}
																			</div>
																		</div>
																	{/if}
																</div>
															{/if}
														</div>
													{/each}
												{/if}
											</div>
										{/if}

										{#if message.editing}
											<div class="absolute -bottom-8 right-1 flex gap-x-1 md:right-0">
												{#if $convo.messages.filter((msg) => msg.role !== 'system' && !msg.submitted).length >= 2 && i === $convo.messages.length - 1 && message.role !== 'assistant'}
													<button
														class="flex items-center gap-x-1 rounded-full bg-green-100 px-3 py-2"
														on:click={() => {
															submitCompletion();
														}}
													>
														<Icon icon={faCheckDouble} class="h-3.5 w-3.5 text-slate-600" />
														<span class="text-xs text-slate-600"> Submit all </span>
													</button>
												{/if}
												{#if message.role !== 'assistant' && message.pendingContent && message.pendingContent !== message.content}
													<button
														class="flex items-center gap-x-1 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
														on:click={(event) => {
															if (message.role === 'system') {
																// If system message, accept the edit instead of submitting at point:
																$convo.messages[i].content = message.pendingContent;
																$convo.messages[i].pendingContent = '';
																$convo.messages[i].editing = false;
																return;
															}
															submitEdit(i);
															event.target.blur();
														}}
													>
														<Icon icon={faCheck} class="h-3.5 w-3.5 text-slate-600" />
														<span class="text-xs text-slate-600">
															{#if message.role === 'system'}
																Set system prompt
															{:else}
																Submit
															{/if}
														</span>
													</button>
												{/if}
												{#if message.role === 'assistant' && message.pendingContent && message.pendingContent !== message.content && message.content !== '...' && i === $convo.messages.length - 1}
													<button
														class="flex items-center gap-x-1.5 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
														on:click={async () => {
															$convo.messages[i].unclosed = true;
															submitCompletion(false);
														}}
													>
														<Icon icon={faEllipsis} class="h-3.5 w-3.5 text-slate-600" />
														<span class="text-xs text-slate-600">Pre-filled response</span>
													</button>
												{/if}
												<button
													class="flex items-center gap-x-1 rounded-full bg-gray-50 px-3 py-2 hover:bg-gray-100"
													on:click={() => {
														$convo.messages[i].editing = false;
														$convo.messages[i].pendingContent = '';
													}}
												>
													<Icon icon={faXmark} class="h-3.5 w-3.5 text-slate-600" />
													<span class="text-xs text-slate-600">Cancel</span>
												</button>
											</div>
										{/if}
										{#if !message.editing}
											<div
												class="absolute bottom-[-32px] left-11 flex items-center gap-x-4 md:bottom-[-28px] md:left-14"
											>
												{#if message.role === 'user' && $convo.versions?.[message.id]}
													{@const versions = $convo.versions[message.id]}
													{@const versionIndex = versions.findIndex((v) => v === null)}
													<div class="flex items-center md:gap-x-1">
														<button
															class="group flex h-6 w-6 shrink-0 rounded-full md:h-3 md:w-3"
															disabled={versionIndex === 0}
															on:click={() => {
																shiftVersion(-1, message, i);
															}}
														>
															<Icon
																icon={faChevronLeft}
																class="m-auto h-3 w-3 text-slate-800 group-disabled:text-slate-500 md:h-2 md:w-2"
															/>
														</button>
														<span class="text-xs tabular-nums">
															{versionIndex + 1} / {versions.length}
														</span>
														<button
															class="group flex h-6 w-6 shrink-0 rounded-full md:h-3 md:w-3"
															disabled={versionIndex === versions.length - 1}
															on:click={() => {
																shiftVersion(1, message, i);
															}}
														>
															<Icon
																icon={faChevronRight}
																class="m-auto h-3 w-3 text-slate-800 group-disabled:text-slate-500 md:h-2 md:w-2"
															/>
														</button>
													</div>
												{/if}

												{#if (message.role === 'assistant' && i > 2 && $convo.messages[i - 2].role === 'assistant' && message.model && $convo.messages[i - 2].model && $convo.messages[i - 2].model.id !== message.model.id) || (message.role === 'assistant' && (i === 1 || i === 2) && message.model && $convo.model.id !== message.model.id)}
													<p class="text-[10px]">{formatModelName(message.model)}</p>
												{/if}
											</div>

											<div
												class="absolute bottom-[-32px] right-0 flex gap-x-0.5 opacity-0 transition-opacity group-hover:opacity-100"
											>
												<button
													class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
													on:click={async () => {
														$convo.messages[i].editing = true;
														$convo.messages[i].pendingContent = $convo.messages[i].content;
														await tick();
														textareaEls[i].style.height = 'auto';
														textareaEls[i].style.height = textareaEls[i].scrollHeight + 'px';
														textareaEls[i].focus();
													}}
												>
													<Icon icon={faPen} class="m-auto h-[11px] w-[11px] text-slate-600" />
												</button>
												{#if message.role !== 'system'}
													<button
														class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
														on:click={() => {
															if (message.role === 'user') {
																saveVersion(message, i);

																// If user message, remove all messages after this one, then regenerate:
																$convo.messages = $convo.messages.slice(0, i + 1);
																submitCompletion();
															} else {
																// History is split on the user message, so get the message before this (which will be the user's):
																const previousUserMessage = $convo.messages[i - 1];
																saveVersion(previousUserMessage, i - 1);

																// If assistant message, remove all messages after this one, including this one, then regenerate:
																$convo.messages = $convo.messages.slice(0, i);
																submitCompletion();
															}
														}}
													>
														<Icon
															icon={faArrowsRotate}
															class="m-auto h-[12px] w-[12px] text-slate-600"
														/>
													</button>
												{/if}
												<button
													class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
													on:click={() => {
														// Remove this message from the conversation:
														$convo.messages = $convo.messages
															.slice(0, i)
															.concat($convo.messages.slice(i + 1));
													}}
												>
													<Icon icon={faXmark} class="m-auto h-[14px] w-[14px] text-slate-600" />
												</button>
											</div>
										{/if}
									</div>
									<button
										on:click={async () => {
											// Insert a blank message inbetween the next message and the next next message:
											let role;
											if (message.role === 'assistant' || message.role === 'system') {
												role = 'user';
											} else {
												role = 'assistant';
											}
											$convo.messages.splice(i + 1, 0, {
												id: Date.now(),
												role,
												content: '',
												editing: true,
											});
											$convo.messages = $convo.messages;
											await tick();
											textareaEls[i + 1].focus();
										}}
										class="z-1 absolute bottom-0 left-1/2 flex h-6 w-6 -translate-x-1/2 translate-y-1/2 items-center justify-center rounded-md border border-gray-300 bg-white opacity-0 transition-opacity hover:bg-gray-200 group-hover:opacity-100"
									>
										<Icon icon={faPlus} class="m-auto h-3 w-3 text-slate-600" />
									</button>
								</li>
							{/if}
						{/each}
					</ul>
				{:else}
					<Button
						variant="outline"
						class="z-[98] mx-auto rounded-t-none !border-t-0 border-dashed text-xs"
						on:click={insertSystemPrompt}
					>
						<Icon icon={faVial} class="mr-2 h-3 w-3 text-slate-600" />
						Add system prompt
					</Button>
					<div class="m-auto flex flex-col items-center gap-5 text-center">
						<Icon icon={faLightbulb} class="h-12 w-12 text-slate-800" />
						<h3 class="text-4xl font-semibold tracking-tight text-slate-800">lluminous</h3>
					</div>
				{/if}
			</section>
			<section
				class="section-input-bottom fixed bottom-4 left-1/2 z-[99] flex w-full max-w-[680px] -translate-x-1/2 flex-col px-5 md:left-[calc((100vw+230px)*0.5)] lg:px-0 ld:max-w-[768px] xl:left-1/2"
			>
				<div class="absolute bottom-full mb-3 flex gap-4 self-center">
					{#if !generating && $convo.messages.filter((msg) => msg.generated).length > 0}
						<Button
							variant="outline"
							on:click={() => {
								const i = $convo.messages.length - 2;
								// Split history on the last user message:
								saveVersion($convo.messages[i], i);

								// Remove last message and run completion again:
								$convo.messages = $convo.messages.slice(0, $convo.messages.length - 1);
								submitCompletion();
							}}
						>
							<Icon icon={faArrowsRotate} class="mr-2 h-3.5 w-3.5 text-slate-600" />
							Regenerate
						</Button>
					{/if}
					{#if generating && $convo.messages.filter((msg) => msg.generated).length > 0}
						<Button
							variant="outline"
							on:click={() => {
								$controller.abort();
							}}
						>
							<Icon icon={faStop} class="mr-2 h-3.5 w-3.5 text-slate-500" />
							<span>Stop generating</span>
						</Button>
					{/if}
				</div>
				{#if $convo.model.provider === 'Local' && !hidingTokenCount && totalTokens > 0}
					<span
						class="absolute bottom-full right-3 mb-3 text-xs"
						transition:fade={{ duration: 300 }}
					>
						{#if currentTokens > 0}
							{currentTokens} tokens in input,
						{/if}
						{totalTokens} tokens in total
					</span>
				{/if}
				<div class="relative flex">
					{#if imageUrls.length > 0}
						<div class="absolute left-[50px] top-2.5 flex gap-x-3">
							{#each imageUrls as url, i}
								<div class="relative">
									<img
										src={url}
										alt=""
										class="h-16 w-16 rounded-lg border border-slate-300 object-cover"
									/>
									<button
										on:click={() => {
											imageUrls.splice(i, 1);
											imageUrls = imageUrls;
											imageUrlsBlacklist.push(url);
											tick().then(() => {
												autoresizeTextarea();
											});
										}}
										class="absolute -right-1 -top-1 flex h-3 w-3 rounded-full bg-black"
									>
										<Icon icon={faXmark} class="m-auto h-2.5 w-2.5 text-white" />
									</button>
								</div>
							{/each}
						</div>
					{/if}
					{#if isMultimodal}
						<button
							class="absolute left-3 top-2.5 rounded-md border border-slate-300 bg-white p-2 transition-colors"
							on:click={() => fileInputEl.click()}
						>
							<input
								type="file"
								accept="image/*"
								class="hidden"
								bind:this={fileInputEl}
								on:change={async (event) => {
									const files = event.target.files;
									for (let i = 0; i < files.length; i++) {
										const file = files[i];
										if (file.type.startsWith('image/')) {
											const dataUrl = await readFileAsDataURL(file);
											imageUrls.push(dataUrl);
											imageUrls = imageUrls;
											tick().then(() => {
												autoresizeTextarea();
											});
										}
									}
								}}
							/>
							<Icon
								icon={faPaperclip}
								class="h-3 w-3 text-slate-800 transition-colors group-disabled:text-slate-400"
							/>
						</button>
					{/if}
					<textarea
						bind:this={inputTextareaEl}
						class="{isMultimodal ? '!pl-[50px]' : ''} {imageUrls.length > 0
							? '!pt-[88px]'
							: ''} h-[50px] max-h-[90dvh] w-full resize-none rounded-xl border border-slate-300 py-3 pl-4 pr-11 font-normal text-slate-800 shadow-sm transition-colors scrollbar-slim focus:border-slate-400 focus:outline-none md:h-[74px] md:px-4"
						rows={1}
						bind:value={content}
						on:paste={async (event) => {
							const items = (event.clipboardData || event.originalEvent.clipboardData).items;
							for (let i = 0; i < items.length; i++) {
								if (items[i].kind === 'file' && items[i].type.startsWith('image/')) {
									const file = items[i].getAsFile();
									const dataUrl = await readFileAsDataURL(file);
									imageUrls.push(dataUrl);
									imageUrls = imageUrls;
									tick().then(() => {
										autoresizeTextarea();
									});
								}
							}
						}}
						on:keydown={(event) => {
							if (event.key === 'Enter' && !event.shiftKey && window.innerWidth > 880) {
								event.preventDefault();
								sendMessage();
							}
						}}
						on:input={async () => {
							autoresizeTextarea();

							const imageLinkedUrls = content.match(imageUrlRegex) || [];
							for (const url of imageLinkedUrls) {
								if (!imageUrls.includes(url) && !imageUrlsBlacklist.includes(url)) {
									imageUrls.push(url);
									imageUrls = imageUrls;
									tick().then(() => {
										autoresizeTextarea();
									});
								}
							}

							if ($convo.model.provider === 'Local') {
								currentTokens = await tokenizeCount(content);
							}
						}}
					/>
					<button
						disabled={content.length === 0}
						class="group absolute bottom-2.5 right-3.5 rounded-md border border-slate-400 bg-white p-2 transition-colors disabled:border-slate-200 md:hidden"
						on:click={sendMessage}
					>
						<Icon
							icon={faArrowUp}
							class="h-3 w-3 text-slate-800 transition-colors group-disabled:text-slate-400"
						/>
					</button>
				</div>
			</section>
		</div>

		<KnobsSidebar {knobsOpen} />
	</div>
</main>

<SettingsModal
	open={!$convo.shared && $openaiAPIKey === '' && $groqAPIKey === '' && $openrouterAPIKey === ''}
	trigger="settings"
	on:fetchModels={fetchModels}
/>

<style lang="postcss">
	:global(.standalone .section-input-bottom) {
		bottom: 32px;
	}
	:global(.standalone .scrollable) {
		@apply pb-[140px];
	}

	:global(.markdown.prose :where(p):not(:where([class~='not-prose'], [class~='not-prose'] *))) {
		@apply my-3;
	}

	:global(
			.markdown.prose
				:where(.prose > :first-child):not(:where([class~='not-prose'], [class~='not-prose'] *))
		) {
		@apply mt-0;
	}

	:global(
			.markdown.prose
				:where(.prose > :last-child):not(:where([class~='not-prose'], [class~='not-prose'] *))
		) {
		@apply mb-0;
	}
</style>
