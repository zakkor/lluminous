<script>
	import { tick } from 'svelte';
	import { v4 as uuidv4 } from 'uuid';
	import Button from './Button.svelte';
	import CompanyLogo from './CompanyLogo.svelte';
	import {
		feCheck,
		feCheckCircle,
		feChevronDown,
		feChevronLeft,
		feChevronRight,
		feCpu,
		feEdit2,
		feMessageCircle,
		feMoreHorizontal,
		fePlus,
		feRefreshCw,
		feUser,
		feX,
	} from './feather';
	import Icon from './Icon.svelte';
	import MessageContent from './MessageContent.svelte';
	import { formatModelName, hasCompanyLogo } from './providers.js';
	import { config } from './stores.js';
	import Toolcall from './Toolcall.svelte';
	import ToolcallButton from './ToolcallButton.svelte';

	export let message;
	export let i;
	export let convo;
	export let generating;
	export let collapsedRanges;

	export let saveMessage;
	export let saveVersion;
	export let saveConversation;
	export let shiftVersion;
	export let insertSystemPrompt;
	export let submitCompletion;
	export let isChoosing;
	export let choiceHandler;
	export let question;
	export let choices;

	export let chose;
	export let activeToolcall;
	export let textareaEls;
	function submitEdit(i) {
		// Update the ID of the edited message:
		if (convo.messages[i].submitted || convo.messages[i].generated) {
			let vid = null;
			const msgBeforeEdit = { ...convo.messages[i] };
			msgBeforeEdit.editing = false;
			msgBeforeEdit.pendingContent = '';
			if (!msgBeforeEdit.vid) {
				vid = uuidv4();
				msgBeforeEdit.vid = vid;
			}
			saveMessage(msgBeforeEdit);

			saveVersion(msgBeforeEdit, i);

			convo.messages[i].id = uuidv4();
			if (!convo.messages[i].vid) {
				convo.messages[i].vid = vid;
			}
			convo.messages[i].editing = false;
			convo.messages[i].content = convo.messages[i].pendingContent;
			convo.messages[i].pendingContent = '';
			saveMessage(convo.messages[i]);
		}

		convo.messages = convo.messages.slice(0, i + 1);
		saveConversation(convo);

		submitCompletion();
	}

	// Sequences of tool calls which are not interrupted by messages also containing text content will be displayed on the same line.
	function collapsedToolcalls(collapsedRange, collapsedMessages, ci, message) {
		if (ci < collapsedMessages.length - 1) {
			// Don't show duplicated toolcalls if these toolcalls will
			// be collapsed into a single line in a later message.
			const nextMessage = collapsedMessages[ci + 1];
			if (nextMessage.role === 'assistant' && nextMessage.toolcalls && !nextMessage.content) {
				return [];
			}
		}

		const i = convo.messages.findIndex((m) => m.id === message.id);

		// Starting from the message `i`, and going backwards, collect all `.toolcalls` until
		// we are interrupted by a message that contains `.content`, or we reach `collapsedRange.starti`
		const toolcalls = [];
		for (let j = i; j >= collapsedRange.starti; j--) {
			const msgIter = convo.messages[j];
			if (msgIter.role === 'assistant' && msgIter.toolcalls) {
				toolcalls.push(msgIter.toolcalls);
			}
			if (msgIter.role === 'assistant' && msgIter.content) {
				break;
			}
		}
		return toolcalls.reverse().flat();
	}
</script>

{#if (['user', 'assistant'].includes(message.role) || (message.role === 'system' && (!message.customInstructions || (message.customInstructions && message.showCustomInstructions)))) && ($config.explicitToolView || !collapsedRanges.some((r) => i >= r.starti && i < r.endi))}
	{@const hasLogo = message.role === 'assistant' && message.models.find(hasCompanyLogo)}
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
		style="z-index: {convo.messages.length - i};"
		on:touchstart={(event) => {
			// Make click trigger hover on mobile:
			event.target.dispatchEvent(new MouseEvent('mouseenter'));
		}}
	>
		{#if i === 0 && message.role !== 'system'}
			<Button
				variant="outline"
				class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 border-dashed text-xs opacity-0 transition-[border-color,opacity] group-hover:opacity-100"
				on:click={insertSystemPrompt}
			>
				<Icon icon={feMessageCircle} class="mr-2 h-3 w-3 text-slate-600" />
				Add system prompt
			</Button>
		{:else if i === 1 && convo.messages[i - 1].role === 'system' && convo.messages[i - 1].customInstructions && !convo.messages[i - 1].showCustomInstructions}
			<Button
				variant="outline"
				class="absolute left-1/2 top-0 z-[98] -translate-x-1/2 rounded-t-none !border-t-0 text-xs opacity-0 transition-opacity group-hover:opacity-100"
				on:click={() => {
					convo.messages[i - 1].showCustomInstructions = true;
					saveMessage(convo.messages[i - 1]);
				}}
			>
				<Icon icon={feEdit2} class="mr-2 h-3 w-3 text-slate-600" />
				Custom instructions
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
					? 'border border-teal-200 bg-teal-100'
					: message.role === 'user'
						? 'border border-slate-200 bg-white'
						: message.role === 'assistant' && !hasLogo
							? 'border border-teal-200 bg-teal-100 pb-px'
							: ''}"
			>
				{#if message.role === 'assistant' && hasLogo}
					{#each message.models as model}
						<CompanyLogo {model} size="w-full h-full" rounded="rounded-[inherit]" />
					{/each}
				{:else}
					<span class="m-auto">
						{#if message.role === 'system'}
							<Icon icon={feMessageCircle} class="h-4 w-4 text-slate-800" />
						{:else if message.role === 'assistant'}
							<Icon icon={feCpu} class="h-4 w-4 text-slate-800" />
						{:else}
							<Icon icon={feUser} class="h-4 w-4 text-slate-800" />
						{/if}
					</span>
				{/if}
			</button>

			<!-- svelte-ignore a11y-no-static-element-interactions -->
			{#if message.editing}
				<textarea
					bind:this={textareaEls[i]}
					class="w-full resize-none border-none bg-transparent p-0 leading-[28px] text-slate-800 outline-none focus:ring-0"
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
					{#if !$config.explicitToolView}
						{@const collapsedRange = collapsedRanges.find((r) => i === r.endi)}
						{#if collapsedRange}
							{@const collapsedMessages = convo.messages
								.slice(collapsedRange.starti, collapsedRange.endi)
								.filter((m) => m.role === 'assistant')}
							{#if collapsedMessages.length > 0}
								{#each collapsedMessages as message, ci}
									{@const toolcallsOnLine = collapsedToolcalls(
										collapsedRange,
										collapsedMessages,
										ci,
										message
									)}

									<MessageContent {message} />

									{#if toolcallsOnLine?.length > 0}
										<div class="-mb-1 flex flex-wrap gap-3 [&:first-child]:mt-1">
											{#each toolcallsOnLine as toolcall, ti}
												{@const toolresponse = convo.messages.find(
													(msg) => msg.toolcallId === toolcall.id
												)}
												<ToolcallButton
													{toolcall}
													{toolresponse}
													active={toolcall.id === activeToolcall?.id}
													on:click={() => {
														activeToolcall = toolcall;
													}}
												/>
											{/each}
										</div>
									{/if}
								{/each}
							{/if}
						{/if}
					{/if}

					{#if message.reasoning}
						<div
							class="{message.thinking
								? 'animate-pulse'
								: ''} -mb-3 flex items-center gap-x-1 text-left text-gray-600"
						>
							{message.thinking ? 'Thinking' : 'Thought'} for {message.thinkingTime < 1
								? 'a bit'
								: Math.ceil(message.thinkingTime) + ' seconds'}
							{#if message.thoughts}
								<Icon
									icon={feChevronDown}
									class="{message.thoughtsExpanded
										? 'rotate-180'
										: ''} h-4 w-4 transition-transform"
								/>
							{/if}
						</div>
					{/if}

					{#if generating && message.role === 'assistant' && i === convo.messages.length - 1 && message.content === '' && !message.toolcalls}
						<div class="mt-2 h-3 w-3 shrink-0 animate-bounce rounded-full bg-slate-600" />
					{/if}

					<MessageContent {message} />

					{#if !$config.explicitToolView && message.toolcalls?.length > 0}
						<div class="-mb-1 flex flex-wrap gap-3 [&:first-child]:mt-1">
							{#each message.toolcalls as toolcall, ti}
								{@const toolresponse = convo.messages.find((msg) => msg.toolcallId === toolcall.id)}
								<ToolcallButton
									{toolcall}
									{toolresponse}
									active={toolcall.id === activeToolcall?.id}
									on:click={() => {
										activeToolcall = toolcall;
									}}
								/>
							{/each}
						</div>
					{/if}

					<!-- OAI toolcalls will always be at the end -->
					{#if message.toolcalls && $config.explicitToolView}
						{#each message.toolcalls as toolcall, ti}
							{@const toolresponse = convo.messages.find((msg) => msg.toolcallId === toolcall.id)}
							<Toolcall
								{toolcall}
								{toolresponse}
								bind:chose
								{isChoosing}
								{choiceHandler}
								{question}
								{choices}
								class="mb-1"
								on:click={() => {
									convo.messages[i].toolcalls[ti].expanded =
										!convo.messages[i].toolcalls[ti].expanded;
									saveMessage(convo.messages[i]);
								}}
							/>
						{/each}
					{/if}
				</div>
			{/if}

			{#if message.editing}
				<div class="absolute -bottom-8 right-1 flex gap-x-1 md:right-0">
					{#if convo.messages.filter((msg) => msg.role !== 'system' && !msg.submitted).length >= 2 && i === convo.messages.length - 1 && message.role !== 'assistant'}
						<button
							class="flex items-center gap-x-1 rounded-full bg-green-100 px-3 py-2"
							on:click={() => {
								submitCompletion();
							}}
						>
							<Icon icon={feCheckCircle} class="h-3.5 w-3.5 text-slate-600" />
							<span class="text-xs text-slate-600"> Submit all </span>
						</button>
					{/if}
					{#if message.role !== 'assistant' && message.pendingContent && message.pendingContent !== message.content}
						<button
							class="flex items-center gap-x-1 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
							on:click={(event) => {
								if (message.role === 'system') {
									// If system message, accept the edit instead of submitting at point:
									convo.messages[i].content = message.pendingContent;
									convo.messages[i].pendingContent = '';
									convo.messages[i].editing = false;
									saveMessage(convo.messages[i]);
									return;
								}
								submitEdit(i);
								event.target.blur();
							}}
						>
							<Icon icon={feCheck} class="h-3.5 w-3.5 text-slate-600" />
							<span class="text-xs text-slate-600">
								{#if message.role === 'system'}
									Set system prompt
								{:else}
									Submit
								{/if}
							</span>
						</button>
					{/if}
					{#if message.role === 'assistant' && message.pendingContent && message.pendingContent !== message.content && message.content !== '...' && i === convo.messages.length - 1}
						<button
							class="flex items-center gap-x-1.5 rounded-full bg-green-50 px-3 py-2 hover:bg-green-100"
							on:click={async () => {
								convo.messages[i].unclosed = true;
								saveMessage(convo.messages[i]);
								submitCompletion(false);
							}}
						>
							<Icon icon={feMoreHorizontal} class="h-3.5 w-3.5 text-slate-600" />
							<span class="text-xs text-slate-600">Pre-filled response</span>
						</button>
					{/if}
					<button
						class="flex items-center gap-x-1 rounded-full bg-gray-50 px-3 py-2 hover:bg-gray-100"
						on:click={() => {
							convo.messages[i].editing = false;
							convo.messages[i].pendingContent = '';
							saveMessage(convo.messages[i]);
						}}
					>
						<Icon icon={feX} class="h-3.5 w-3.5 text-slate-600" />
						<span class="text-xs text-slate-600">Cancel</span>
					</button>
				</div>
			{/if}
			{#if !message.editing}
				<div
					class="absolute bottom-[-32px] left-11 flex items-center gap-x-4 md:bottom-[-28px] md:left-14"
				>
					{#if message.role === 'user' && convo.versions?.[message.vid]}
						{@const versions = convo.versions[message.vid]}
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
									icon={feChevronLeft}
									class="m-auto h-3.5 w-3.5 text-slate-800 group-disabled:text-slate-500 md:h-3 md:w-3"
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
									icon={feChevronRight}
									class="m-auto h-3.5 w-3.5 text-slate-800 group-disabled:text-slate-500 md:h-3 md:w-3"
								/>
							</button>
						</div>
					{/if}

					{#if (message.role === 'assistant' && i > 2 && convo.messages[i - 2].role === 'assistant' && message.models?.length > 0 && convo.messages[i - 2].models?.length > 0 && convo.messages[i - 2].models?.[0].id !== message.models?.[0].id) || (message.role === 'assistant' && (i === 1 || i === 2) && message.models[0] && convo.models[0].id !== message.models[0].id)}
						<p class="text-[10px]">{formatModelName(message.models[0])}</p>
					{/if}
				</div>

				<div
					class="{!generating
						? 'group-hover:opacity-100'
						: ''} absolute bottom-[-32px] right-1 flex gap-x-2 opacity-0 transition-opacity md:gap-x-0.5"
				>
					<button
						class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
						on:click={async () => {
							convo.messages[i].editing = true;
							convo.messages[i].pendingContent = convo.messages[i].content;
							await tick();
							textareaEls[i].style.height = 'auto';
							textareaEls[i].style.height = textareaEls[i].scrollHeight + 'px';
							textareaEls[i].focus();
							saveMessage(convo.messages[i]);
						}}
					>
						<Icon icon={feEdit2} strokeWidth={3} class="m-auto h-[11px] w-[11px] text-slate-600" />
					</button>
					{#if message.role !== 'system'}
						<button
							class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
							on:click={() => {
								activeToolcall = null;

								if (message.role === 'user') {
									if (!message.vid) {
										message.vid = uuidv4();
										saveMessage(message);
									}
									saveVersion(message, i);

									// If user message, remove all messages after this one, then regenerate:
									convo.messages = convo.messages.slice(0, i + 1);
									submitCompletion();
								} else {
									// History is split on the user message, so get the message before this (which will be the user's):
									const previousUserMessage = convo.messages[i - 1];
									if (!previousUserMessage.vid) {
										previousUserMessage.vid = uuidv4();
										saveMessage(previousUserMessage);
									}
									saveVersion(previousUserMessage, i - 1);

									// If assistant message, remove all messages after this one, including this one, then regenerate:
									convo.messages = convo.messages.slice(0, i);
									submitCompletion();
								}
								saveConversation(convo);
							}}
						>
							<Icon
								icon={feRefreshCw}
								strokeWidth={3}
								class="m-auto h-[12px] w-[12px] text-slate-600"
							/>
						</button>
					{/if}
					<button
						class="flex h-7 w-7 shrink-0 rounded-full hover:bg-gray-100"
						on:click={() => {
							// Remove this message from the conversation:
							convo.messages = convo.messages.slice(0, i).concat(convo.messages.slice(i + 1));
							// FIXME: Delete message from db
							saveConversation(convo);
						}}
					>
						<Icon icon={feX} strokeWidth={3} class="m-auto h-[14px] w-[14px] text-slate-600" />
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
				const msg = {
					id: uuidv4(),
					role,
					content: '',
					editing: true,
				};
				convo.messages.splice(i + 1, 0, msg);
				convo.messages = convo.messages;
				await tick();
				textareaEls[i + 1].focus();

				saveMessage(msg);
				saveConversation(convo);
			}}
			class="{!generating
				? 'group-hover:opacity-100'
				: ''} z-1 absolute bottom-0 left-1/2 flex h-6 w-6 -translate-x-1/2 translate-y-1/2 items-center justify-center rounded-md border border-slate-200 bg-white opacity-0 transition-opacity hover:bg-gray-200"
		>
			<Icon icon={fePlus} class="m-auto h-3 w-3 text-slate-600" />
		</button>
	</li>
{/if}
