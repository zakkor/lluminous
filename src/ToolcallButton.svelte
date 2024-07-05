<script>
	import Icon from './Icon.svelte';
	import { feTerminal } from './feather.js';

	export let toolcall;
	export let toolresponse;
	export let active = false;
</script>

<button
	data-trigger="toolcall"
	class="{toolresponse ? '' : 'animate-pulse'} {active
		? 'bg-slate-800'
		: 'bg-gray-200 hover:bg-gray-300'} group/toolcall relative self-start rounded-full p-2 transition-colors duration-500"
	on:click
>
	<Icon icon={feTerminal} class="{active ? 'text-white' : 'text-slate-800'} h-4 w-4" />
	{#if toolcall}
		<div
			class="pointer-events-none absolute bottom-[calc(100%+8px)] left-1/2 w-min -translate-x-1/2 rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-opacity group-hover/toolcall:opacity-100"
		>
			{toolcall.name}
		</div>
	{/if}
</button>
