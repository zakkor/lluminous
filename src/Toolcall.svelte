<script>
	import { createEventDispatcher } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import JsonView from './svelte-json-view/JsonView.svelte';
	import Icon from './Icon.svelte';
	import { feCheck, feChevronDown, feLoader, feX } from './feather.js';

	const dispatch = createEventDispatcher();

	export let toolcall;
	export let toolresponse;
	export let collapsable = true;
	export let closeButton = false;
	let className = '';
	export { className as class };

	$: finished = toolcall.finished || toolresponse;

	let displayType = null;
	let displayTypeDisabled = false;
	let displayedContent = null;
	$: if (toolresponse && toolresponse.content && toolresponse.content.contentType) {
		displayType = null;
		displayedContent = null;
		// When a special contentType is returned, if a `content` field is not present,
		// we automatically fill it with the value of whatever argument (single) was passed into the tool call.
		// This is done in order to save tokens, because it prevents us having to re-send the content again to the model.
		if (!toolresponse.content.content && Object.keys(toolcall.arguments).length === 1) {
			displayedContent = toolcall.arguments[Object.keys(toolcall.arguments)[0]];
		} else {
			displayedContent = toolresponse.content.content;
		}
		if (displayedContent) {
			const contentType = toolresponse.content.contentType;
			// Get display type for this type of content:
			if (contentType.startsWith('image/')) {
				displayType = 'image';
			} else if (contentType === 'text/html') {
				displayType = 'webpage';
			} else if (contentType === 'text/markdown') {
				displayType = 'markdown';
			}
			// ...
		}
	} else {
		displayType = null;
		displayedContent = null;
	}
</script>

<div class="{className} flex h-full w-full flex-col bg-white">
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<svelte:element
		this={collapsable ? 'button' : 'div'}
		class="{!collapsable || toolcall.expanded ? '' : 'rounded-b-lg'} {collapsable
			? 'transition-colors hover:bg-gray-50'
			: ''} flex items-center gap-4 rounded-t-lg border border-slate-200 py-3 pl-4 pr-5 text-sm text-slate-700 sm:gap-3"
		on:click
	>
		{#if finished}
			<span in:fade={{ duration: 300 }}>
				<Icon icon={feCheck} class="h-5 w-5 text-slate-700" />
			</span>
		{:else}
			<span in:fade={{ duration: 300 }}>
				<Icon icon={feLoader} class="h-5 w-5 animate-spin text-slate-700" />
			</span>
		{/if}
		<div class="flex flex-col gap-1.5 sm:flex-row sm:items-baseline">
			<span>{finished ? 'Used' : 'Using'} tool:</span>
			<code class="font-semibold tracking-tight">{toolcall.name}</code>
		</div>
		<div class="-my-1.5 -mr-2 ml-auto flex items-center gap-2">
			{#if displayType !== null}
				<button
					on:click={(event) => {
						event.stopPropagation();
						displayTypeDisabled = !displayTypeDisabled;
					}}
					class="flex whitespace-nowrap rounded-full border border-slate-200 px-3 py-1 text-[10px] font-medium transition-colors hover:bg-gray-100 md:text-xs"
				>
					{#if displayTypeDisabled}
						View component
					{:else}
						View code
					{/if}
				</button>
			{/if}
			{#if collapsable}
				<Icon
					icon={feChevronDown}
					class="{toolcall.expanded
						? 'rotate-180'
						: ''} h-4 w-4 text-slate-700 transition-transform"
				/>
			{:else if closeButton}
				<button
					on:click={() => {
						dispatch('close');
					}}
					class="flex rounded-full p-3 transition-colors hover:bg-gray-100"
				>
					<Icon icon={feX} class="ml-auto h-4 w-4 text-slate-700" />
				</button>
			{/if}
		</div>
	</svelte:element>
	{#if !collapsable || toolcall.expanded}
		<div
			transition:slide={{ duration: 300 }}
			class="h-full max-h-[80vh] overflow-y-auto scrollbar-ultraslim"
		>
			{#if displayType === null || displayTypeDisabled}
				<div
					class="{toolresponse
						? 'border-b-0'
						: 'rounded-b-lg'} flex flex-col whitespace-pre-wrap border border-t-0 border-slate-200 bg-white px-4 py-3 font-mono text-sm text-slate-800 [overflow-wrap:anywhere]"
				>
					{#if typeof toolcall.arguments === 'object'}
						{#if Object.keys(toolcall.arguments).length === 1}
							{toolcall.arguments[Object.keys(toolcall.arguments)[0]]}
						{:else}
							<JsonView json={toolcall.arguments} />
						{/if}
					{:else if toolcall.arguments}
						{toolcall.arguments}
					{:else}
						<div class="h-3 w-3 shrink-0 animate-bounce rounded-full bg-slate-600" />
					{/if}
				</div>
				{#if toolresponse}
					<div class="h-px w-full border-t border-dashed border-slate-300" />
					<div class="flex flex-col rounded-b-lg border border-t-0 border-slate-200">
						<span class="px-4 pt-3 text-sm font-medium tracking-[0.01em] text-slate-700"
							>Result:</span
						>
						<div
							class="flex flex-col whitespace-pre-wrap rounded-[inherit] bg-white px-4 py-3 font-mono text-sm text-slate-800 [overflow-wrap:anywhere]"
						>
							{#if !toolresponse.content}
								<span class="italic">blank</span>
							{:else if typeof toolresponse.content === 'object'}
								<JsonView json={toolresponse.content} />
							{:else}
								{toolresponse.content}
							{/if}
						</div>
					</div>
				{/if}
			{:else if toolresponse && displayType === 'image'}
				<div class="flex flex-col rounded-b-lg border border-t-0 border-slate-200">
					<img src={displayedContent} alt="" class="w-full object-contain object-[0]" />
				</div>
			{:else if toolresponse && displayType === 'webpage'}
				<div class="flex h-full flex-col rounded-b-lg border border-t-0 border-slate-200">
					<iframe
						title="Webpage"
						srcdoc={displayedContent}
						class="h-full min-h-[50vh] w-full"
						frameborder="0"
						allowfullscreen
					/>
				</div>
			{/if}
		</div>
	{/if}
</div>
