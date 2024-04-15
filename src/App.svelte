<script>
	import { tick } from 'svelte';
	import { slide, fade } from 'svelte/transition';
	import { complete, conversationToString } from './convo.js';
	import Toolbar from './Toolbar.svelte';
	import Button from './Button.svelte';
	import {
		faArrowsRotate,
		faCheck,
		faPenToSquare,
		faPlus,
		faStop,
		faXmark,
	} from '@fortawesome/free-solid-svg-icons';
	import Markdown from '@magidoc/plugin-svelte-marked';
	import Icon from './Icon.svelte';
	import { persisted, persistedPicked } from './localstorage.js';

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
	let convo = persistedPicked(history, (h) => h.entries[h.convoId]);

	let content = '';
	let generating = false;

	let currentTokens = 0;
	let totalTokens = 0;
	let hidingTokenCount = false;

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
</script>

<main class="flex h-screen w-screen">
	<aside class="flex flex-col w-[230px] px-3 py-4 border-r">
		<button
			on:click={() => {
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
			}}
			class="w-full text-sm flex items-center border mb-2 text-left hover:bg-gray-100 py-2 pl-3 pr-4 rounded-lg"
		>
			New chat
			<Icon icon={faPlus} class="w-3 h-3 text-slate-600 ml-auto" />
		</button>
		<ol class="list-none flex flex-col gap-1">
			{#each Object.values($history.entries) as convo}
				<li>
					<button
						on:click={() => {
							$history.convoId = convo.id;
							convo = persistedPicked(history, (h) => h.entries[h.convoId]);
						}}
						class="{$history.convoId === convo.id
							? 'bg-gray-100'
							: ''} text-sm w-full text-left hover:bg-gray-100 py-2 px-3 rounded-lg"
					>
						{new Intl.DateTimeFormat('en-UK', { dateStyle: 'short', timeStyle: 'short' }).format(
							convo.id
						)}
					</button>
				</li>
			{/each}
		</ol>
	</aside>
	<div class="flex-1 flex flex-col">
		<section
			bind:this={scrollableEl}
			class="scrollable w-full h-full flex flex-col overflow-y-auto pb-[180px]"
			on:scroll={() => {
				if (scrollableEl.scrollTop + scrollableEl.clientHeight >= scrollableEl.scrollHeight - 100) {
					hidingTokenCount = false;
				} else {
					hidingTokenCount = true;
				}
			}}
		>
			{#if $convo.messages.length > 0}
				<ul
					class="w-full flex flex-col !list-none divide-y divide-slate-200 border-b border-slate-200 mb-3"
				>
					{#each $convo.messages as message, i}
						{#if ['system', 'user', 'assistant'].includes(message.role)}
							<li
								data-role={message.role}
								class="{!message.generated &&
								!message.submitted &&
								message.role !== 'system' &&
								message.role !== 'tool'
									? 'bg-yellow-50/40'
									: message.role === 'assistant'
										? 'bg-slate-50/75'
										: ''} group pt-6 pb-8"
							>
								<div class="relative gap-x-5 flex self-start w-full lg:max-w-[768px] mx-auto">
									<button
										on:click={() => {
											// Toggle between user and assistant:
											if (message.role === 'user') {
												message.role = 'assistant';
											} else {
												message.role = 'user';
											}
										}}
										class="flex w-10 h-10 shrink-0 rounded {message.role === 'assistant'
											? 'bg-teal-300'
											: message.role === 'system'
												? 'bg-blue-200'
												: message.role === 'tool'
													? 'bg-blue-100 border-1 border-slate-300 border-dashed'
													: 'bg-red-200'}"
									>
										<span class="text-lg m-auto">
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
									<div class="flex self-start w-full mt-1">
										{#if generating && message.role === 'assistant' && i === $convo.messages.length - 1 && message.content === ''}
											<div class="mt-1 w-3 h-3 rounded-full bg-slate-700/75 animate-pulse" />
										{/if}
										<!-- svelte-ignore a11y-no-static-element-interactions -->
										{#if message.editing}
											<textarea
												bind:this={textareaEls[i]}
												class="border-none leading-[28px] text-[#334155] rounded-lg p-0 bg-transparent outline-none focus:ring-0 resize-none w-full"
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
												class="prose markdown max-w-none prose-slate prose-p:text-slate-800 prose-pre:bg-white prose-pre:text-slate-800 prose-pre:border prose-pre:border-slate-200 break-words flex flex-col w-full"
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
									</div>
									{#if message.editing}
										<div class="flex gap-x-0.5 absolute bottom-[8px] translate-y-full right-0">
											{#if message.role !== 'assistant' && message.pendingContent && message.pendingContent !== message.content}
												<button
													class="hover:bg-green-100 bg-green-50 gap-x-1 items-center rounded-full py-2 px-3 flex"
													on:click={(event) => {
														submitEdit(i);
														event.target.blur();
													}}
												>
													<Icon icon={faCheck} class="text-slate-600 w-3.5 h-3.5" />
													<span class="text-slate-600 text-xs">Submit</span>
												</button>
											{/if}
											{#if message.role === 'assistant' && message.pendingContent && message.pendingContent !== message.content && message.content !== '...' && i === $convo.messages.length - 1}
												<button
													class="hover:bg-green-100 bg-green-50 gap-x-1 items-center rounded-full py-2 px-3 flex"
													on:click={async () => {
														$convo.messages[i].unclosed = true;
														submitCompletion(false);
													}}
												>
													<Icon icon={faCheck} class="text-slate-600 w-3.5 h-3.5" />
													<span class="text-slate-600 text-xs">Continue generation</span>
												</button>
											{/if}
											<button
												class="hover:bg-gray-100 bg-gray-50 gap-x-1 items-center rounded-full py-2 px-3 flex"
												on:click={() => {
													$convo.messages[i].editing = false;
													$convo.messages[i].pendingContent = '';
												}}
											>
												<Icon icon={faXmark} class="text-slate-600 w-3.5 h-3.5" />
												<span class="text-slate-600 text-xs">Cancel</span>
											</button>
										</div>
									{/if}
									{#if !message.editing}
										<div
											class="flex gap-x-0.5 absolute opacity-0 group-hover:opacity-100 transition-opacity -top-3 right-0 translate-x-full"
										>
											<button
												class="hover:bg-gray-100 rounded-full p-2 flex"
												on:click={async () => {
													$convo.messages[i].editing = true;
													$convo.messages[i].pendingContent = $convo.messages[i].content;
													await tick();
													textareaEls[i].style.height = 'auto';
													textareaEls[i].style.height = textareaEls[i].scrollHeight + 2 + 'px';
													textareaEls[i].focus();
												}}
											>
												<Icon icon={faPenToSquare} class="w-3.5 text-slate-600 h-3.5 mt-0.5" />
											</button>
											{#if message.role !== 'system'}
												<button
													class="hover:bg-gray-100 rounded-full p-2 flex"
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
													<Icon icon={faArrowsRotate} class="text-slate-600 w-3.5 h-3.5 mt-0.5" />
												</button>
											{/if}
											<button
												class="hover:bg-gray-100 rounded-full p-2 flex"
												on:click={() => {
													// Remove this message from the conversation:
													$convo.messages = $convo.messages
														.slice(0, i)
														.concat($convo.messages.slice(i + 1));
												}}
											>
												<Icon icon={faXmark} class="text-slate-600 w-4 h-4 mr-[0.5px] mt-px" />
											</button>
										</div>
									{/if}
									<button
										on:click={() => {
											// Insert a blank message after this one:
											$convo.messages.push({ role: 'assistant', content: '...' });
											$convo.messages = $convo.messages;
										}}
										class="z-1 absolute bottom-0 translate-y-[43px] bg-white opacity-0 group-hover:opacity-100 transition-opacity left-1/2 -translate-x-1/2 w-6 h-6 border border-gray-300 hover:bg-gray-200 rounded-md flex items-center justify-center"
									>
										<Icon icon={faPlus} class="text-slate-600 w-3 h-3 m-auto" />
									</button>
								</div>
							</li>
						{/if}
					{/each}
				</ul>
			{/if}
		</section>
		<section class="left-1/2 bottom-4 flex flex-col -translate-x-1/2 fixed w-full lg:max-w-[768px]">
			<div class="absolute bottom-full flex gap-4 self-center mb-3">
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
						<Icon icon={faArrowsRotate} class="text-slate-600 w-3.5 h-3.5 mr-2" />
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
						<Icon icon={faStop} class="text-slate-500 w-3.5 h-3.5 mr-2" />
						<span>Stop generating</span>
					</Button>
				{/if}
			</div>

			{#if $convo.local && !hidingTokenCount && totalTokens > 0}
				<span class="absolute bottom-full mb-3 text-xs right-3" transition:fade={{ duration: 300 }}>
					{#if currentTokens > 0}
						{currentTokens} tokens in input,
					{/if}
					{totalTokens} tokens in total
				</span>
			{/if}

			<textarea
				bind:this={inputTextareaEl}
				class="border border-slate-300 w-full rounded-lg resize-none focus:border-slate-500 transition-colors focus:ring-0 px-4 py-3 shadow-sm"
				rows={3}
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
		on:rerender={async () => {
			$convo = $convo;
			$convo.messages = $convo.messages;
			if ($convo.local) {
				totalTokens = await tokenizeCount(conversationToString($convo));
			}
		}}
	/>
</main>

<style lang="postcss">
	/* Render <tool_call> and <tool_response> inside assistant messages */
	[data-role='assistant'] {
		:global(tool_call) {
			display: block;
			@apply font-mono border text-sm mt-2 border-slate-300 border-dashed bg-white px-5 py-3 rounded-lg whitespace-normal transition-all;
		}
		:global(tool_call)::before {
			content: 'Function call';
			@apply font-sans block pb-2 text-slate-600 text-sm;
		}
		:global(tool_call:has(+ .tool_response)) {
			@apply rounded-b-none;
		}
		:global(tool_response) {
			@apply flex flex-col font-mono border border-t-0 rounded-t-none text-sm border-slate-300 border-dashed bg-white px-5 py-3 rounded-lg whitespace-pre transition-all;
		}
		:global(tool_response)::before {
			content: 'Output';
			@apply font-sans block pb-2 text-slate-600 text-sm;
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
