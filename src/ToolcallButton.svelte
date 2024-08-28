<script>
	import Icon from './Icon.svelte';
	import { feTerminal } from './feather.js';
	import { onMount } from 'svelte';

	export let toolcall;
	export let toolresponse;
	export let active = false;

	let flash = false;

	onMount(() => {
		flash = true;
		setTimeout(() => {
			flash = false;
		}, 1000);
	});
</script>

<button
	data-trigger="toolcall"
	class="{toolresponse ? '' : 'animate-pulse'} {active
		? 'bg-slate-800'
		: 'bg-gray-200 hover:bg-gray-300'} {flash ? 'ring-2 ring-offset-1 ring-gray-400': ''} group/toolcall relative self-start rounded-full p-2 transition-[background-color,box-shadow] duration-500"
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
