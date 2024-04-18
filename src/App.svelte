<script>
	import { onMount, tick } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import { complete, conversationToString, detectFormat, isModelLocal } from './convo.js';
	import Toolbar from './Toolbar.svelte';
	import Button from './Button.svelte';
	import {
		faArrowsRotate,
		faBarsStaggered,
		faCheck,
		faGear,
		faPen,
		faPlus,
		faShareFromSquare,
		faStop,
		faXmark,
	} from '@fortawesome/free-solid-svg-icons';
	import { faLightbulb, faTrashCan } from '@fortawesome/free-regular-svg-icons';
	import Markdown from '@magidoc/plugin-svelte-marked';
	import Icon from './Icon.svelte';
	import { persisted, persistedPicked } from './localstorage.js';
	import { getRelativeDate } from './date.js';
	import ModelSelector from './ModelSelector.svelte';
	import { compressAndEncode, decodeAndDecompress } from './share.js';
	import { writable } from 'svelte/store';

	// import { marked } from 'marked';
	// import markedKatex from 'marked-katex-extension';
	// import markedFootnote from 'marked-footnote';
	// marked.use(
	// 	markedKatex({
	// 		throwOnError: false,
	// 	}),
	// 	markedFootnote({
	// 		refMarkers: true,
	// 	})
	// );

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
			summary: null,
			local: false,
			model: 'openchat/openchat-7b:free',
			tmpl: 'none',
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
		$convo.messages = $convo.messages.slice(0, i + 1);
		submitCompletion();
	}

	async function submitCompletion(insertUnclosed = true) {
		generating = true;

		if ($convo.local) {
			totalTokens = await tokenizeCount(conversationToString($convo));
		}

		if (insertUnclosed) {
			$convo.messages.push({ role: 'assistant', content: '', unclosed: true, generated: true });
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
			// Check for stoppage:
			// For local models, `.stop` will be true.
			// For external models, `.choices[0].finish_reason` will be 'stop'.
			if ($convo.local && chunk.stop) {
				generating = false;

				if (chunk.stopping_word === '</tool_call>') {
					// Stopping word is not included, so we add it back manually:
					$convo.messages[i].content += chunk.stopping_word;

					// Parse the response and extract what is inside <tool_call>...</tool_call>:
					const toolcallContent = parseToolcall($convo.messages[i].content);
					if (toolcallContent.name && toolcallContent.arguments) {
						// Call the tool
						const resp = await fetch('http://localhost:8081/tool', {
							method: 'POST',
							body: JSON.stringify(toolcallContent),
						});
						const toolresponse = await resp.json();
						$convo.messages.push({
							role: 'tool',
							content: `<tool_response>${toolresponse}</tool_response>`,
						});
						$convo.messages = $convo.messages;

						submitCompletion();
					}
				}

				return;
			}
			if (!$convo.local && chunk.choices && chunk.choices[0].finish_reason === 'stop') {
				generating = false;
				return;
			}

			if ($convo.local) {
				$convo.messages[i].content += chunk.content;
				totalTokens = await tokenizeCount(conversationToString($convo));
			} else {
				$convo.messages[i].content += chunk.choices[0].delta.content;
			}
		};

		const onabort = () => {
			generating = false;
		};

		complete($convo, onupdate, onabort);
	}

	async function tokenizeCount(content) {
		const tokenizeResp = await fetch('http://localhost:8081/tokenize_count', {
			method: 'POST',
			body: JSON.stringify({ content }),
		});
		return parseInt(await tokenizeResp.text());
	}

	function parseToolcall(content) {
		const startTag = '<tool_call>';
		const endTag = '</tool_call>';

		const startIndex = content.indexOf(startTag);
		const endIndex = content.indexOf(endTag);

		if (startIndex !== -1 && endIndex !== -1 && endIndex > startIndex + startTag.length) {
			try {
				const toolcall = JSON.parse(content.slice(startIndex + startTag.length, endIndex));
				return toolcall;
			} catch (e) {
				console.error('Error parsing tool call:', e);
				return {};
			}
		}

		return {};
	}

	async function onEnterMessage(event) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			if (content.length > 0) {
				$convo.messages.push({ role: 'user', content, submitted: true });
				$convo.messages = $convo.messages;
				await tick();
				scrollableEl.scrollTop = scrollableEl.scrollHeight;

				content = '';
				currentTokens = 0;

				await tick();
				inputTextareaEl.style.height = 'auto';
				inputTextareaEl.style.height = inputTextareaEl.scrollHeight + 2 + 'px';

				submitCompletion();
			}
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
			$history.convoId = existingNewConvo.id;
			convo = persistedPicked(history, (h) => h.entries[h.convoId]);
			historyOpen = false;
			inputTextareaEl.focus();
			return;
		}

		const convoData = {
			id: Date.now(),
			summary: null,
			local: false,
			model: $convo.model || 'openchat/openchat-7b:free',
			tmpl: 'none',
			messages: [],
		};
		$history.convoId = convoData.id;
		$history.entries[convoData.id] = convoData;
		convo = persistedPicked(history, (h) => h.entries[h.convoId]);

		historyOpen = false;

		inputTextareaEl.focus();
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
			const encoded = await compressAndEncode($convo.messages);
			const share = `${window.location.protocol}//${window.location.host}/?s=${encoded}`;
			if (share.length > 2000) {
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
		loading = true;

		const local = isModelLocal(newModel);
		if (local) {
			// For local models, we need to tell the server to load them:
			await fetch(`http://localhost:8081/model`, {
				method: 'POST',
				body: JSON.stringify({
					model: newModel,
				}),
			});
		}
		setModel(newModel);

		loading = false;
	}

	let autodetectedFormat = false;

	function setModel(newModel) {
		$convo.model = newModel;
		const { detected, local } = detectFormat(newModel);
		if (detected) {
			$convo.tmpl = detected;
			autodetectedFormat = true;
		} else {
			autodetectedFormat = false;
		}
		$convo.local = local;
	}

	let models = [];

	onMount(async () => {
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
				.then((messages) => {
					let id = Date.now();
					const existingShared = Object.values($history.entries).find((convo) => convo.shared);
					if (existingShared) {
						id = existingShared.id;
					}
					const convoData = {
						id,
						shared: true,
						local: false,
						model: $convo.model || 'openchat/openchat-7b:free',
						tmpl: 'none',
						messages,
					};
					$history.convoId = convoData.id;
					$history.entries[convoData.id] = convoData;
					convo = persistedPicked(history, (h) => h.entries[h.convoId]);
				})
				.catch((err) => {
					console.error('Error decoding shared conversation:', err);
				});
		}

		fetch('https://openrouter.ai/api/v1/models', { method: 'GET' })
			.then((response) => response.json())
			.then((json) => {
				const externalModels = json.data;
				// Reorder the model list, by placing a few handpicked models at the start of the list:

				// Then append external models to the local models (if any):
				models = models.concat(json.data).map((m) => {
					return {
						id: m.id,
						name: m.name,
						local: false,
					};
				});
			})
			.catch((error) => {
				console.error('Error:', error);
			});

		try {
			const modelsData = await (
				await fetch('http://localhost:8081/models', { method: 'GET' })
			).json();
			models = modelsData.models.map((m) => {
				return {
					id: m,
					name: m,
					local: true,
				};
			});

			// Get the model that is currently loaded on the llama.cpp server, if any:
			const modelData = await (
				await fetch('http://localhost:8081/model', { method: 'GET' })
			).json();
			setModel(modelData.model || null);
		} catch (error) {
			console.warn('Local llama.cpp server is not running, running in external mode only.');
		}
	});
</script>

<svelte:window on:touchstart={closeSidebars} on:click={closeSidebars} />

<main class="flex h-[calc(100dvh)] w-screen flex-col">
	<div class="flex items-center border-b border-slate-200 px-4 py-1 md:hidden">
		<button on:click={newConversation} class="flex p-3">
			<Icon icon={faPlus} class="ml-auto h-4 w-4 text-slate-700" />
		</button>
		<button data-trigger="history" class="flex p-3" on:click={() => (historyOpen = !historyOpen)}>
			<Icon icon={faBarsStaggered} class="m-auto h-4 w-4 text-slate-700" />
		</button>

		{#if !$convo.shared}
			<ModelSelector {convo} {models} {loading} {loadModel} class="mx-auto" />
		{:else}
			<p class="mx-auto text-sm font-semibold">Shared conversation</p>
		{/if}

		<button
			class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
			on:click={shareConversation}
		>
			<Icon icon={faShareFromSquare} class="m-auto h-4 w-4 text-slate-700" />
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
										$history.convoId = Object.values($history.entries)[0].id;
									}
									delete $history.entries[convo.id];
									$history.entries = $history.entries;
								}}
								class="z-1 absolute right-0 top-0 flex h-full w-12 rounded-br-lg rounded-tr-lg bg-gradient-to-l from-gray-100 from-65% to-transparent pr-3 opacity-0 transition-opacity group-hover:opacity-100"
							>
								<Icon icon={faTrashCan} class="m-auto mr-0 h-3 w-3 shrink-0 text-slate-700" />
							</button>
						</li>
					{/each}
				{/each}
			</ol>
		</aside>
		<div class="flex flex-1 flex-col">
			<section
				bind:this={scrollableEl}
				class="scrollable flex h-full w-full flex-col overflow-y-auto pb-[160px]"
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
				<div class="hidden items-center border-b border-slate-200 px-4 py-1 md:flex">
					<div class="" />

					{#if !$convo.shared}
						<ModelSelector {convo} {models} {loading} {loadModel} class="mx-auto" />
					{:else}
						<p class="mx-auto text-sm font-semibold">Shared conversation</p>
					{/if}

					<button
						class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
						on:click={shareConversation}
					>
						<Icon icon={faShareFromSquare} class="m-auto h-4 w-4 text-slate-700" />
					</button>
					<button
						data-trigger="settings"
						class="flex rounded-full p-3 transition-colors hover:bg-gray-100 xl:hidden"
						on:click={() => (settingsOpen = !settingsOpen)}
					>
						<Icon icon={faGear} class="m-auto h-4 w-4 text-slate-700" />
					</button>
				</div>
				{#if $convo.messages.length > 0}
					<ul
						class="mb-3 flex w-full !list-none flex-col divide-y divide-slate-200 border-b border-slate-200"
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
											? 'bg-slate-50/75'
											: ''} group relative px-5 pb-12 pt-6 ld:px-8"
									style="z-index: {$convo.messages.length - i};"
									on:click={(event) => {
										// Make click trigger hover on mobile:
										event.target.dispatchEvent(new MouseEvent('mouseenter'));
									}}
								>
									<div
										class="relative mx-auto flex w-full max-w-[680px] gap-x-3.5 self-start md:gap-x-5 ld:max-w-[768px]"
									>
										<button
											on:click={() => {
												// Toggle between user and assistant:
												if (message.role === 'user') {
													message.role = 'assistant';
												} else {
													message.role = 'user';
												}
											}}
											class="flex h-8 w-8 shrink-0 rounded md:h-10 md:w-10 md:rounded-[5px] {message.role ===
											'assistant'
												? 'bg-teal-300'
												: message.role === 'system'
													? 'bg-blue-200'
													: message.role === 'tool'
														? 'border-1 border-dashed border-slate-300 bg-blue-100'
														: 'bg-red-200'}"
										>
											<span class="m-auto text-base md:text-lg">
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

										{#if generating && message.role === 'assistant' && i === $convo.messages.length - 1 && message.content === ''}
											<div
												class="mt-2 h-3 w-3 shrink-0 animate-pulse rounded-full bg-slate-700/75"
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
													event.target.style.height = event.target.scrollHeight + 2 + 'px';
												}}
											/>
										{:else}
											<div
												class="markdown prose prose-slate flex w-full max-w-none flex-col break-words prose-p:whitespace-pre-wrap prose-p:text-slate-800 prose-pre:my-0 prose-pre:whitespace-pre-wrap prose-pre:border prose-pre:border-slate-200 prose-pre:bg-white prose-pre:text-slate-800"
											>
												<Markdown source={message.content} />
												<!-- Render tool response inline by merging in the next 'tool' message -->
												{#if i !== $convo.messages.length - 1 && $convo.messages[i + 1].role === 'tool'}
													<div class="tool_response" transition:slide={{ duration: 500 }}>
														{@html $convo.messages[i + 1].content}
													</div>
												{/if}
											</div>
										{/if}

										{#if message.editing}
											<div
												class="absolute -bottom-9 right-1 flex gap-x-0.5 md:bottom-2 md:right-0 md:translate-y-full"
											>
												{#if message.role !== 'assistant' && message.pendingContent && message.pendingContent !== message.content}
													<button
														class="flex items-center gap-x-1 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
														on:click={(event) => {
															submitEdit(i);
															event.target.blur();
														}}
													>
														<Icon icon={faCheck} class="h-3.5 w-3.5 text-slate-600" />
														<span class="text-xs text-slate-600">Submit</span>
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
											<div
												class="absolute -bottom-9 right-1 flex gap-x-0.5 opacity-0 transition-opacity group-hover:opacity-100 ld:-top-2 ld:bottom-auto ld:right-0 ld:translate-x-full"
											>
												<button
													class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
													on:click={async () => {
														$convo.messages[i].editing = true;
														$convo.messages[i].pendingContent = $convo.messages[i].content;
														await tick();
														textareaEls[i].style.height = 'auto';
														textareaEls[i].style.height = textareaEls[i].scrollHeight + 2 + 'px';
														textareaEls[i].focus();
													}}
												>
													<Icon
														icon={faPen}
														class="m-auto h-[11px] w-[11px] text-slate-500 md:h-3 md:w-3 md:text-slate-600"
													/>
												</button>
												{#if message.role !== 'system'}
													<button
														class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
														on:click={() => {
															// If user message, remove all messages after this one, then regenerate:
															if (message.role === 'user') {
																$convo.messages = $convo.messages.slice(0, i + 1);
																submitCompletion();
															} else {
																// If assistant message, remove all messages after this one, including this one, then regenerate:
																$convo.messages = $convo.messages.slice(0, i);
																submitCompletion();
															}
														}}
													>
														<Icon
															icon={faArrowsRotate}
															class="m-auto h-3 w-3 text-slate-500 md:h-3.5 md:w-3.5 md:text-slate-600"
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
													<Icon
														icon={faXmark}
														class="m-auto h-3.5 w-3.5 text-slate-500 md:h-4 md:w-4 md:text-slate-600"
													/>
												</button>
											</div>
										{/if}
									</div>
									<button
										on:click={() => {
											// Insert a blank message after this one:
											$convo.messages.push({ role: 'assistant', content: '...' });
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
					<div class="m-auto flex flex-col items-center gap-5 text-center">
						<Icon icon={faLightbulb} class="h-12 w-12 text-slate-800" />
						<h3 class="text-4xl font-semibold tracking-tight text-slate-800">lluminous</h3>
					</div>
				{/if}
			</section>
			<section
				class="fixed bottom-4 left-1/2 z-[99] flex w-full max-w-[680px] -translate-x-1/2 flex-col px-5 md:left-[calc((100vw+230px)*0.5)] lg:px-0 ld:max-w-[768px] xl:left-1/2"
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
				{#if $convo.local && !hidingTokenCount && totalTokens > 0}
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
				<textarea
					bind:this={inputTextareaEl}
					class="w-full resize-none rounded-lg border border-slate-300 px-4 py-3 font-normal text-slate-800 shadow-sm transition-colors focus:border-slate-500 focus:ring-0"
					rows={2}
					enterkeyhint="send"
					bind:value={content}
					on:keydown={onEnterMessage}
					on:input={async () => {
						inputTextareaEl.style.height = 'auto';
						inputTextareaEl.style.height = inputTextareaEl.scrollHeight + 2 + 'px';
						if ($convo.local) {
							currentTokens = await tokenizeCount(content);
						}
					}}
				/>
			</section>
		</div>
		<Toolbar
			{convo}
			{settingsOpen}
			on:rerender={async () => {
				$convo = $convo;
				$convo.messages = $convo.messages;
				if ($convo.local) {
					totalTokens = await tokenizeCount(conversationToString($convo));
				}
			}}
		/>
	</div>
</main>

<style lang="postcss">
	/* Render <tool_call> and <tool_response> inside assistant messages */
	[data-role='assistant'] {
		:global(tool_call) {
			display: block;
			@apply mt-2 whitespace-normal rounded-lg border border-dashed border-slate-300 bg-white px-5 py-3 font-mono text-sm transition-all;
		}
		:global(tool_call)::before {
			content: 'Function call';
			@apply block pb-2 font-sans text-sm text-slate-600;
		}
		:global(tool_call:has(+ .tool_response)) {
			@apply rounded-b-none;
		}
		:global(tool_response) {
			@apply flex flex-col whitespace-pre rounded-lg rounded-t-none border border-t-0 border-dashed border-slate-300 bg-white px-5 py-3 font-mono text-sm transition-all;
		}
		:global(tool_response)::before {
			content: 'Output';
			@apply block pb-2 font-sans text-sm text-slate-600;
		}
	}

	textarea {
		-ms-overflow-style: none; /* Internet Explorer 10+ */
		scrollbar-width: none; /* Firefox */
	}
	textarea::-webkit-scrollbar {
		display: none; /* Safari and Chrome */
	}

	.scrollable {
		-ms-overflow-style: none; /* Internet Explorer 10+ */
		scrollbar-width: none; /* Firefox */
	}
	.scrollable::-webkit-scrollbar {
		display: none; /* Safari and Chrome */
	}
</style>
