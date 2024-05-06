<script>
	import { onMount, tick } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import { complete, conversationToString } from './convo.js';
	import Toolbar from './Toolbar.svelte';
	import Button from './Button.svelte';
	import {
		faArrowUp,
		faArrowUpRightFromSquare,
		faArrowsRotate,
		faBarsStaggered,
		faCheck,
		faChevronDown,
		faChevronLeft,
		faChevronRight,
		faCircleNotch,
		faGear,
		faPen,
		faPlus,
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
	import { persisted, persistedPicked } from './localstorage.js';
	import { getRelativeDate } from './date.js';
	import { compressAndEncode, decodeAndDecompress } from './share.js';
	import { providers } from './providers.js';
	import ModelSelector from './ModelSelector.svelte';
	import CompanyLogo from './CompanyLogo.svelte';
	import { remoteServer } from './stores.js';
	import { get } from 'svelte/store';

	marked.use(
		markedKatex({
			throwOnError: false,
		})
	);

	const history = persisted('history');
	if (!$history) {
		$history = {
			convoId: null,
			entries: {},
		};
	}
	if (!$history.convoId) {
		const convoData = {
			id: Date.now(),
			model: { id: null, name: 'No model loaded', provider: null },
			messages: [],
		};
		$history.convoId = convoData.id;
		$history.entries[convoData.id] = convoData;
	}

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

	let convo = persistedPicked(history, (h) => h.entries[h.convoId]);

	let content = '';
	let generating = false;

	let currentTokens = 0;
	let totalTokens = 0;
	let hidingTokenCount = false;

	let historyOpen = false;
	let settingsOpen = false;

	let scrollableEl = null;
	let textareaEls = [];
	let inputTextareaEl;

	function submitEdit(i) {
		const message = $convo.messages[i];
		splitHistory(message, i);

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
			$convo.controller.abort();
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
				// FIXME: Parallel tool calls: loop through tool_calls
				// FIXME: Parallel tool calls: loop through tool_calls
				// FIXME: Parallel tool calls: loop through tool_calls
				// FIXME: Parallel tool calls: loop through tool_calls
				// FIXME: Parallel tool calls: loop through tool_calls
				// FIXME: Parallel tool calls: loop through tool_calls
				if (chunk.choices.length === 0) {
					$convo.messages[i].error = 'Refused to respond';
					generating = false;
					return;
				}

				if (chunk.choices[0].delta.tool_calls) {
					if (chunk.choices[0].delta.tool_calls[0].function.name) {
						if (!$convo.messages[i].toolcall) {
							$convo.messages[i].toolcall = {
								id: chunk.choices[0].delta.tool_calls[0].id,
								name: '',
								arguments: '',
								expanded: false,
							};
						}
						$convo.messages[i].toolcall.name += chunk.choices[0].delta.tool_calls[0].function.name;
					}
					$convo.messages[i].toolcall.arguments +=
						chunk.choices[0].delta.tool_calls[0].function.arguments;
				}
				if (chunk.choices[0].delta.content) {
					$convo.messages[i].content += chunk.choices[0].delta.content;
				}
			}

			// Check for stoppage:
			// For local models, `.stop` will be true.
			// For external models, `.choices[0].finish_reason` will be 'stop'.
			if ($convo.model.provider === 'Local' && chunk.stop) {
				generating = false;
				return;
			}

			if (
				!($convo.model.provider === 'Local') &&
				chunk.choices &&
				(chunk.choices[0].finish_reason === 'stop' ||
					chunk.choices[0].finish_reason === 'tool_calls')
			) {
				generating = false;

				if ($convo.messages[i].toolcall) {
					// Toolcall arguments are now finalized, we can parse them:
					$convo.messages[i].toolcall.arguments = JSON.parse($convo.messages[i].toolcall.arguments);

					// Call the tool
					const resp = await fetch(`${$remoteServer.address}/tool`, {
						method: 'POST',
						headers: {
							Authorization: `Basic ${$remoteServer.password}`,
						},
						body: JSON.stringify($convo.messages[i].toolcall),
					});
					const toolresponse = await resp.text();
					$convo.messages.push({
						id: Date.now(),
						tool_call_id: $convo.messages[i].toolcall.id,
						role: 'tool',
						content: toolresponse,
					});
					$convo.messages = $convo.messages;

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
				// ...
			} else {
				if (chatResp.choices[0].message.content) {
					$convo.messages[i].content = chatResp.choices[0].message.content;
					generating = false;
				} else if (chatResp.choices[0].message.tool_calls) {
					const toolcall = chatResp.choices[0].message.tool_calls[0];
					$convo.messages[i].toolcall = {
						id: toolcall.id,
						name: toolcall.function.name,
						arguments: JSON.parse(toolcall.function.arguments),
						expanded: false,
					};

					// Call the tool
					const resp = await fetch(`${$remoteServer.address}/tool`, {
						method: 'POST',
						headers: {
							Authorization: `Basic ${$remoteServer.password}`,
						},
						body: JSON.stringify($convo.messages[i].toolcall),
					});
					const toolresponse = await resp.text();
					$convo.messages.push({
						id: Date.now(),
						tool_call_id: $convo.messages[i].toolcall.id,
						role: 'tool',
						content: toolresponse,
					});
					$convo.messages = $convo.messages;

					submitCompletion();
				}
			}
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
			$convo.messages.push({ id: Date.now(), role: 'user', content, submitted: true });
			$convo.messages = $convo.messages;
			await tick();
			scrollableEl.scrollTop = scrollableEl.scrollHeight;

			content = '';
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
			convo = persistedPicked(history, (h) => h.entries[h.convoId]);
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
		convo = persistedPicked(history, (h) => h.entries[h.convoId]);

		historyOpen = false;

		inputTextareaEl.focus();
	}

	function splitHistory(message, i) {
		// Split history at this point:
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

	function closeSidebars(event) {
		if (
			historyOpen &&
			!event.target.closest('[data-sidebar="history"]') &&
			!event.target.closest('[data-trigger="history"]')
		) {
			historyOpen = false;
		}
		if (
			settingsOpen &&
			!event.target.closest('[data-trigger="settings"]') &&
			!event.target.closest('[data-sidebar="settings"]')
		) {
			settingsOpen = false;
		}
	}

	async function shareConversation() {
		const sharePromise = new Promise(async (resolve) => {
			const encoded = await compressAndEncode({
				name: $convo.model.name,
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
							name: m.name || `${provider.name}: ${m.id}`,
							provider: provider.name,
						}));
						return externalModels;
					})
					.catch(() => {
						console.log('Error fetching models from provider', provider.name);
						return [];
					});
			});

			const results = await Promise.all(promises);
			const externalModels = results.flat();

			const priorityOrder = [
				{ exactly: ['openai/gpt-4-turbo'] },
				{ exactly: ['openai/gpt-3.5-turbo'] },
				{ fromProvider: 'Groq', exactlyNot: ['llama2-70b-4096', 'gemma-7b-it'] },
				{ exactly: ['meta-llama/llama-3-70b-instruct', 'meta-llama/llama-3-8b-instruct'] },
				{
					startsWith: 'anthropic/',
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
					startsWith: 'openai/',
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
				{ startsWith: 'mistralai/' },
				{ startsWith: 'cohere/' },
			];

			function getPriorityIndex(model) {
				for (let i = 0; i < priorityOrder.length; i++) {
					const rule = priorityOrder[i];
					if (rule.exactly && rule.exactly.includes(model.id)) {
						return i;
					}
					if (rule.startsWith && model.id.startsWith(rule.startsWith)) {
						if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
							continue;
						}
						return i;
					}
					if (rule.fromProvider && model.provider === rule.fromProvider) {
						if (rule.exactlyNot && rule.exactlyNot.includes(model.id)) {
							continue;
						}
						return i;
					}
				}
				return priorityOrder.length;
			}

			externalModels.sort((a, b) => {
				const aIndex = getPriorityIndex(a);
				const bIndex = getPriorityIndex(b);

				if (aIndex === bIndex) {
					return a.id.localeCompare(b.id);
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

	onMount(async () => {
		if (window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone) {
			document.body.classList.add('standalone');
		}

		const params = new URLSearchParams(window.location.search);
		let share;
		if (params.has('s')) {
			share = params.get('s');
		} else if (params.has('sl')) {
			const response = await fetch(`https://zak.oni2025.ro/p/${params.get('sl')}/`);
			share = await response.text();
		}
		if (share) {
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
						model: { id: null, name: decoded.name, provider: null },
						messages: decoded.messages,
					};
					$history.convoId = convoData.id;
					$history.entries[convoData.id] = convoData;
					convo = persistedPicked(history, (h) => h.entries[h.convoId]);
				})
				.catch((err) => {
					console.error('Error decoding shared conversation:', err);
				});
		}

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
			$convo.controller.abort();
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
				class="!absolute left-1/2 line-clamp-1 -translate-x-1/2 whitespace-nowrap text-sm font-semibold"
			>
				{$convo.model.name}
			</p>
		{/if}

		<button
			class="ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={shareConversation}
		>
			<Icon icon={faArrowUpRightFromSquare} class="m-auto h-4 w-4 text-slate-700" />
		</button>
		<button
			data-trigger="settings"
			class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={() => (settingsOpen = !settingsOpen)}
		>
			<Icon icon={faGear} class="m-auto h-4 w-4 text-slate-700" />
		</button>
	</div>
	<div class="relative flex h-full flex-1 overflow-hidden">
		<aside
			data-sidebar="history"
			class="{historyOpen
				? ''
				: '-translate-x-full'} fixed top-0 z-[100] flex h-full w-[230px] flex-col border-r bg-white px-3 py-4 transition-transform duration-300 ease-in-out md:static md:translate-x-0"
		>
			<button
				on:click={newConversation}
				class="mb-1 flex w-full items-center rounded-lg border py-2.5 pl-3 pr-4 text-left text-sm font-medium hover:bg-gray-100"
			>
				New chat
				<Icon icon={faPlus} class="ml-auto h-3.5 w-3.5 text-slate-700" />
			</button>
			<ol class="flex list-none flex-col">
				{#each historyBuckets as { relativeDate, convos } (relativeDate)}
					<li class="mb-2 ml-3 mt-6 text-xs font-medium text-slate-600">
						{relativeDate}
					</li>
					{#each convos as convo (convo.id)}
						<li class="group relative">
							<button
								on:click={() => {
									if (generating) {
										get(convo).controller.abort();
									}

									historyOpen = false;

									$history.convoId = convo.id;
									convo = persistedPicked(history, (h) => h.entries[h.convoId]);

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
		</aside>
		<div class="flex flex-1 flex-col">
			<div class="relative hidden items-center border-b border-slate-200 px-2 py-1 md:flex">
				{#if !$convo.shared}
					<ModelSelector {convo} {models} class="!absolute left-1/2 z-[99] -translate-x-1/2" />
				{:else if $convo.model}
					<p class="!absolute left-1/2 -translate-x-1/2 text-sm font-semibold">
						{$convo.model.name}
					</p>
				{/if}

				<button
					class="ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
					on:click={shareConversation}
				>
					<Icon icon={faArrowUpRightFromSquare} class="m-auto h-4 w-4 text-slate-700" />
				</button>
				<button
					data-trigger="settings"
					class="flex rounded-full p-3 transition-colors hover:bg-gray-100 xl:hidden"
					on:click={() => (settingsOpen = !settingsOpen)}
				>
					<Icon icon={faGear} class="m-auto h-4 w-4 text-slate-700" />
				</button>
			</div>
			<section
				bind:this={scrollableEl}
				class="scrollable slimscrollbar flex h-full w-full flex-col overflow-y-auto pb-[130px] md:pb-[150px]"
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
									on:click={(event) => {
										// Make click trigger hover on mobile:
										event.target.dispatchEvent(new MouseEvent('mouseenter'));
									}}
								>
									{#if i === 0 && message.role !== 'system'}
										<Button
											variant="outline-small"
											class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 rounded-t-none !border-t-0  border-dashed opacity-0 transition-opacity group-hover:opacity-100"
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
											'assistant'
												? 'bg-teal-300'
												: message.role === 'system'
													? 'bg-blue-200'
													: message.role === 'tool'
														? 'border-1 border-dashed border-slate-300 bg-blue-100'
														: 'bg-red-200'}"
										>
											<span class="m-auto text-base">
												{#if message.role === 'system'}
													S
												{:else if message.role === 'assistant'}
													A
												{:else if message.role === 'tool'}
													T
												{:else}
													U
												{/if}
											</span>
										</button>

										{#if generating && message.role === 'assistant' && i === $convo.messages.length - 1 && message.content === '' && !message.toolcall}
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
														class="markdown prose prose-slate flex w-full max-w-none flex-col break-words prose-p:whitespace-pre-wrap prose-p:text-slate-800 prose-code:break-all prose-pre:my-0 prose-pre:whitespace-pre-wrap prose-pre:break-all prose-pre:border prose-pre:border-slate-200 prose-pre:bg-white prose-pre:text-slate-800"
													>
														<Markdown source={message.content} />
													</div>
												{/if}

												<!-- OAI toolcalls will always be at the end -->
												{#if message.toolcall}
													{@const hasToolResponse =
														i !== $convo.messages.length - 1 &&
														message.toolcall &&
														$convo.messages[i + 1].role === 'tool'}
													<div class="flex w-full flex-col bg-white">
														<button
															class="{message.toolcall.expanded
																? ''
																: 'rounded-b-lg'} flex items-center gap-3 rounded-t-lg border border-slate-200 py-3 pl-4 pr-5 text-sm text-slate-700 transition-colors hover:bg-gray-50"
															on:click={() => {
																$convo.messages[i].toolcall.expanded =
																	!$convo.messages[i].toolcall.expanded;
															}}
														>
															{#if !hasToolResponse}
																<Icon
																	icon={faCircleNotch}
																	class="h-4 w-4 animate-spin text-slate-700"
																/>
															{:else}
																<Icon icon={faCheck} class="h-4 w-4 text-slate-700" />
															{/if}
															<span>
																{hasToolResponse ? 'Used' : 'Using'} tool:
																<code class="ml-1 font-semibold">{message.toolcall.name}</code>
															</span>
															<Icon
																icon={faChevronDown}
																class="{message.toolcall.expanded
																	? 'rotate-180'
																	: ''} ml-auto h-3 w-3 text-slate-700 transition-transform"
															/>
														</button>
														{#if message.toolcall.expanded}
															<div transition:slide={{ duration: 300 }}>
																<div
																	class="{hasToolResponse
																		? 'border-b-0'
																		: 'rounded-b-lg'} whitespace-pre-wrap break-all border border-t-0 border-slate-200 bg-white px-4 py-3 font-mono text-sm text-slate-800"
																>
																	{#if typeof message.toolcall.arguments === 'object'}
																		{#if Object.keys(message.toolcall.arguments).length === 1}
																			{message.toolcall.arguments[
																				Object.keys(message.toolcall.arguments)[0]
																			]}
																		{:else}
																			<JsonView json={message.toolcall.arguments} />
																		{/if}
																	{:else}
																		{message.toolcall.arguments}
																	{/if}
																</div>
																{#if i !== $convo.messages.length - 1 && message.toolcall && $convo.messages[i + 1].role === 'tool'}
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
																			{#if $convo.messages[i + 1].content}
																				{$convo.messages[i + 1].content}
																			{:else}
																				<span class="italic">blank</span>
																			{/if}
																		</div>
																	</div>
																{/if}
															</div>
														{/if}
													</div>
												{/if}
											</div>
										{/if}

										{#if message.editing}
											<div class="absolute -bottom-8 right-1 flex gap-x-1 md:right-0">
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
														class="flex items-center gap-x-1 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
														on:click={async () => {
															$convo.messages[i].unclosed = true;
															submitCompletion(false);
														}}
													>
														<Icon icon={faCheck} class="h-3.5 w-3.5 text-slate-600" />
														<span class="text-xs text-slate-600">Continue generation</span>
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
											<div class="absolute bottom-[-28px] left-14 flex items-center gap-x-4">
												{#if message.role === 'user' && $convo.versions?.[message.id]}
													{@const versions = $convo.versions[message.id]}
													<div class="flex items-center gap-x-1">
														<button
															class="group flex h-3 w-3 shrink-0 rounded-full"
															disabled={versions.findIndex((v) => v === null) === 0}
															on:click={() => {
																const activeVersionIndex = versions.findIndex((v) => v === null);
																const newVersionIndex = activeVersionIndex - 1;

																$convo.versions[message.id][activeVersionIndex] =
																	$convo.messages.slice(i);

																const newMessages = $convo.versions[message.id][newVersionIndex];

																$convo.messages = $convo.messages.slice(0, i).concat(newMessages);

																$convo.versions[message.id][newVersionIndex] = null;
															}}
														>
															<Icon
																icon={faChevronLeft}
																class="m-auto h-2 w-2 text-slate-800 group-disabled:text-slate-500"
															/>
														</button>
														<span class="text-xs tabular-nums">
															{versions.findIndex((v) => v === null) + 1} / {versions.length}
														</span>
														<button
															class="group flex h-3 w-3 shrink-0 rounded-full"
															disabled={versions.findIndex((v) => v === null) ===
																versions.length - 1}
															on:click={() => {
																const activeVersionIndex = versions.findIndex((v) => v === null);
																const newVersionIndex = activeVersionIndex + 1;

																$convo.versions[message.id][activeVersionIndex] =
																	$convo.messages.slice(i);

																const newMessages = $convo.versions[message.id][newVersionIndex];

																$convo.messages = $convo.messages.slice(0, i).concat(newMessages);

																$convo.versions[message.id][newVersionIndex] = null;
															}}
														>
															<Icon
																icon={faChevronRight}
																class="m-auto h-2 w-2 text-slate-800 group-disabled:text-slate-500"
															/>
														</button>
													</div>
												{/if}

												{#if (message.role === 'assistant' && i > 2 && $convo.messages[i - 2].role === 'assistant' && message.model && $convo.messages[i - 2].model && $convo.messages[i - 2].model.id !== message.model.id) || (message.role === 'assistant' && (i === 1 || i === 2) && message.model && $convo.model.id !== message.model.id)}
													<div class="flex items-center gap-x-1.5">
														<CompanyLogo model={message.model} size="h-2.5 w-2.5" />
														<p class="text-[10px]">{message.model.name}</p>
													</div>
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
																splitHistory(message, i);

																// If user message, remove all messages after this one, then regenerate:
																$convo.messages = $convo.messages.slice(0, i + 1);
																submitCompletion();
															} else {
																// History is split on the user message, so get the message before this (which will be the user's):
																const previousUserMessage = $convo.messages[i - 1];
																splitHistory(previousUserMessage, i - 1);

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
										on:click={() => {
											// Insert a blank message inbetween the next message and the next next message:
											$convo.messages.splice(i + 1, 0, { role: 'assistant', content: '...' });
											$convo.messages = $convo.messages;
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
						variant="outline-small"
						class="z-[98] mx-auto rounded-t-none !border-t-0 border-dashed"
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
					{#if $convo.messages.find((msg) => msg.editing) && $convo.messages.findIndex((msg) => msg.editing) !== $convo.messages.length - 1 && $convo.messages[$convo.messages.length - 1].role !== 'assistant'}
						<Button
							variant="outline"
							class="!border-green-300/80 hover:!border-green-300"
							on:click={() => {
								submitCompletion();
							}}
						>
							Submit all edits
						</Button>
					{/if}
					{#if !generating && $convo.messages.filter((msg) => msg.generated).length > 0}
						<Button
							variant="outline"
							on:click={() => {
								const i = $convo.messages.length - 2;
								// Split history on the last user message:
								splitHistory($convo.messages[i], i);

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
								$convo.controller.abort();
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
					<textarea
						bind:this={inputTextareaEl}
						class="slimscrollbar h-[50px] max-h-[90dvh] w-full resize-none rounded-xl border border-slate-300 py-3 pl-4 pr-11 font-normal text-slate-800 shadow-sm transition-colors focus:border-slate-400 focus:outline-none md:h-[74px] md:px-4"
						rows={1}
						bind:value={content}
						on:keydown={(event) => {
							if (event.key === 'Enter' && !event.shiftKey && window.innerWidth > 880) {
								event.preventDefault();
								sendMessage();
							}
						}}
						on:input={async () => {
							autoresizeTextarea();

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
		<Toolbar
			{settingsOpen}
			on:rerender={async () => {
				$convo = $convo;
				$convo.messages = $convo.messages;
				if ($convo.model.provider === 'Local') {
					totalTokens = await tokenizeCount(conversationToString($convo));
				}
			}}
			on:fetchModels={fetchModels}
		/>
	</div>
</main>

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
