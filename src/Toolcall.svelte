<script>
	import { fade, slide } from 'svelte/transition';
	import JsonView from './svelte-json-view/JsonView.svelte';
	import Icon from './Icon.svelte';
	import { feCheck, feChevronDown, feLoader } from './feather.js';

	export let toolcall;
	export let toolresponse;
	export let collapsable = true;
	let className = '';
	export { className as class };

	$: finished = toolcall.finished || toolresponse;
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
					class="{finished ? '' : 'animate-spin'} h-4 w-4 text-slate-700"
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
		{/if}
	</svelte:element>
	{#if !collapsable || toolcall.expanded}
		<div transition:slide={{ duration: 300 }}>
			<div
				class="{toolresponse
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
			{#if toolresponse}
				<div class="h-px w-full border-t border-dashed border-slate-300" />
				<div class="flex flex-col rounded-b-lg border border-t-0 border-slate-200">
					<span class="px-4 pt-3 text-sm font-medium tracking-[0.01em] text-slate-700">Result:</span
					>
					<div
						class="whitespace-pre-wrap break-all rounded-[inherit] bg-white px-4 py-3 font-mono text-sm text-slate-800"
					>
						{#if toolresponse.content}
							{toolresponse.content}
						{:else}
							<span class="italic">blank</span>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
