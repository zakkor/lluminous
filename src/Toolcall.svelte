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

	let toolresponseJson = null;
	function parseToolResponse(toolresponse) {
		try {
			return JSON.parse(toolresponse.content);
		} catch (error) {
			console.error(error);
			return null;
		}
	}
	let contentTypeDisplay = null;
	$: if (toolresponse) {
		contentTypeDisplay = null;
		toolresponseJson = parseToolResponse(toolresponse);
		if (toolresponseJson && toolresponseJson.content_type && toolresponseJson.content) {
			// Get display type for this type of content:
			if (toolresponseJson.content_type.startsWith('image/')) {
				contentTypeDisplay = 'image';
			} else if (toolresponseJson.content_type === 'text/html') {
				contentTypeDisplay = 'webpage';
			}
			// ...
		}
	}
</script>

<div class="{className} flex w-full flex-col bg-white">
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<svelte:element
		this={collapsable ? 'button' : 'div'}
		class="{!collapsable || toolcall.expanded ? '' : 'rounded-b-lg'} {collapsable
			? 'transition-colors hover:bg-gray-50'
			: ''} flex items-center gap-3 rounded-t-lg border border-slate-200 py-3 pl-4 pr-5 text-sm text-slate-700"
		on:click
	>
		{#key finished}
			<span in:fade={{ duration: 300 }}>
				<Icon
					icon={finished ? feCheck : feLoader}
					class="{finished ? '' : 'animate-spin'} h-5 w-5 text-slate-700"
				/>
			</span>
		{/key}
		<span>
			{finished ? 'Used' : 'Using'} tool:
			<code class="ml-1 font-semibold">{toolcall.name}</code>
		</span>
		{#if collapsable}
			<Icon
				icon={feChevronDown}
				class="{toolcall.expanded
					? 'rotate-180'
					: ''} ml-auto h-4 w-4 text-slate-700 transition-transform"
			/>
		{:else if closeButton}
			<button
				on:click={() => {
					dispatch('close');
				}}
				class="-my-1.5 -mr-2 ml-auto flex rounded-full p-3 transition-colors hover:bg-gray-100"
			>
				<Icon icon={feX} class="ml-auto h-4 w-4 text-slate-700" />
			</button>
		{/if}
	</svelte:element>
	{#if !collapsable || toolcall.expanded}
		<div transition:slide={{ duration: 300 }}>
			{#if contentTypeDisplay === null}
				<div
					class="{toolresponse
						? 'border-b-0'
						: 'rounded-b-lg'} whitespace-pre-wrap border border-t-0 border-slate-200 bg-white px-4 py-3 font-mono text-sm text-slate-800 [overflow-wrap:anywhere]"
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
					<div
						class="flex max-h-[80vh] flex-col overflow-y-auto rounded-b-lg border border-t-0 border-slate-200 scrollbar-ultraslim"
					>
						<span class="px-4 pt-3 text-sm font-medium tracking-[0.01em] text-slate-700"
							>Result:</span
						>
						<div
							class="whitespace-pre-wrap rounded-[inherit] bg-white px-4 py-3 font-mono text-sm text-slate-800 [overflow-wrap:anywhere]"
						>
							{#if toolresponse.content}
								{toolresponse.content}
							{:else}
								<span class="italic">blank</span>
							{/if}
						</div>
					</div>
				{/if}
			{:else if contentTypeDisplay === 'image'}
				<div
					class="flex max-h-[80vh] flex-col overflow-y-auto rounded-b-lg border border-t-0 border-slate-200 scrollbar-ultraslim"
				>
					<img src={toolresponseJson.content} alt="" class="w-full object-contain object-[0]" />
				</div>
			{:else if contentTypeDisplay === 'webpage'}
				<iframe title="Webpage">
					<h3 class="text-3xl font-semibold">Soon</h3>
				</iframe>
			{/if}
		</div>
	{/if}
</div>
