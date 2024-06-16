<script>
	import Icon from './Icon.svelte';
	import { feTerminal } from './feather.js';

	export let i;
	export let toolcall;
	export let toolresponse;
</script>

<button
	data-trigger="toolcall"
	class="{toolresponse
		? ''
		: 'animate-pulse'} group/toolcall relative self-start rounded-full bg-gray-100/80 p-2 transition-colors hover:bg-gray-200"
	on:click
>
	<Icon icon={feTerminal} class="h-4 w-4 text-slate-800" />
	<span
		class="absolute -top-1 right-0 inline-flex h-3 w-3 items-center justify-center rounded-full bg-slate-800 text-[6px] font-semibold text-white"
	>
		{i + 1}
	</span>
	{#if toolcall}
		<div
			class="pointer-events-none absolute bottom-[calc(100%+8px)] left-1/2 w-min -translate-x-1/2 rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-opacity group-hover/toolcall:opacity-100"
		>
			{toolcall.name}
		</div>
	{/if}
</button>
