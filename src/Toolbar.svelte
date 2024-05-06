<script>
	import { createEventDispatcher } from 'svelte';
	// import { conversationTemplates } from './templates.js';
	import { groqAPIKey, openrouterAPIKey, params, remoteServer, toolSchema } from './stores.js';
	import Button from './Button.svelte';
	import { faSync } from '@fortawesome/free-solid-svg-icons';
	import Icon from './Icon.svelte';

	const dispatch = createEventDispatcher();

	// export let convo;
	export let settingsOpen = false;
</script>

<aside
	data-sidebar="settings"
	class="{settingsOpen
		? ''
		: 'translate-x-full'} fixed right-0 top-0 z-[100] flex h-full w-[230px] flex-col gap-2 border-l border-slate-200 bg-white px-3 py-4 transition-transform ease-in-out xl:static xl:translate-x-0"
>
	<label class="mb-6 flex flex-col text-[10px] uppercase tracking-wide">
		<div class="mb-2 ml-[3px] flex items-baseline">
			<span>Temperature</span>
			<span class="ml-auto">{$params.temperature}</span>
		</div>
		<input
			type="range"
			min={0}
			max={2}
			bind:value={$params.temperature}
			step={0.1}
			class="appearance-none overflow-hidden rounded-full border border-slate-300
				[&::-webkit-slider-runnable-track]:h-2.5 [&::-webkit-slider-runnable-track]:rounded-full [&::-webkit-slider-runnable-track]:bg-slate-100
				[&::-webkit-slider-thumb]:h-2.5 [&::-webkit-slider-thumb]:w-2.5 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full
			[&::-webkit-slider-thumb]:bg-slate-700 [&::-webkit-slider-thumb]:shadow-[-407px_0_0_404px_theme('colors.slate.200')]"
		/>
	</label>

	<!-- FIXME: -->
	<!-- {#if $convo.local}
		<select
			class="{autodetectedFormat
				? 'border-green-300/80 bg-green-50/75'
				: 'border-red-300 bg-red-50/75'} rounded-lg text-sm"
			value={$convo.tmpl}
			on:change={(event) => {
				$convo.tmpl = event.target.value;
			}}
		>
			<option value={null} disabled>Select a prompt format</option>
			{#each promptFormats as fmt}
				<option value={fmt}>{fmt}</option>
			{/each}
		</select>
	{/if} -->

	<!-- <select
		class="w-full rounded-lg border-slate-300 text-sm"
		value={conversationTemplate}
		on:change={async (event) => {
			loadConversationTemplate(event.target.value);
		}}
	>
		<option value={null} disabled>Load from template</option>
		{#each Object.keys(conversationTemplates) as name}
			<option value={name}>{name}</option>
		{/each}
	</select> -->

	<label class="flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">OpenRouter API Key</span>
		<input
			type="text"
			bind:value={$openrouterAPIKey}
			placeholder="Enter your API key"
			class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors focus:border-slate-400 focus:outline-none"
			on:change={() => {
				if ($openrouterAPIKey.length === 73 || $openrouterAPIKey.length === 0) {
					dispatch('fetchModels');
				}
			}}
		/></label
	>

	<label class="flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">Groq API Key</span>
		<input
			type="text"
			bind:value={$groqAPIKey}
			placeholder="Enter your API key"
			class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors focus:border-slate-400 focus:outline-none"
			on:change={() => {
				if ($groqAPIKey.length === 56 || $groqAPIKey.length === 0) {
					dispatch('fetchModels');
				}
			}}
		/></label
	>

	<label class="mt-6 flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">Server address</span>
		<input
			type="text"
			bind:value={$remoteServer.address}
			placeholder=""
			class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors focus:border-slate-400 focus:outline-none"
		/></label
	>

	<label class="flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">Server password</span>
		<input
			type="password"
			bind:value={$remoteServer.password}
			placeholder=""
			class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors focus:border-slate-400 focus:outline-none"
		/></label
	>

	<label class="mt-3 flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]"
			>Tool schema<span
				class="ml-2 self-end rounded bg-teal-300 px-1 py-0.5 text-[10px] uppercase text-black"
				>Beta</span
			></span
		>
		<textarea
			value={JSON.stringify($toolSchema)}
			on:change={(event) => {
				const schema = event.target.value.length > 0 ? event.target.value : '[]';
				$toolSchema = JSON.parse(schema);
			}}
			rows={10}
			class="rounded-lg border border-slate-300 px-3 py-2 text-sm transition-colors focus:border-slate-400 focus:outline-none"
		/></label
	>
	<Button
		variant="outline"
		class="self-start"
		on:click={async () => {
			const schema = await (
				await fetch(`${$remoteServer.address}/tool_schema`, {
					method: 'GET',
					headers: {
						Authorization: `Basic ${$remoteServer.password}`,
					},
				})
			).text();
			$toolSchema = JSON.parse(schema);
		}}
	>
		<Icon icon={faSync} class="mr-2 h-3 w-3 text-slate-700" />
		Sync
	</Button>
</aside>
