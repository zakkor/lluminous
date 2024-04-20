<script>
	// import { createEventDispatcher, onMount } from 'svelte';
	// import { conversationTemplates } from './templates.js';
	import { groqAPIKey, openrouterAPIKey, toolSchema } from './stores.js';
	import Button from './Button.svelte';
	import { faSync } from '@fortawesome/free-solid-svg-icons';
	import Icon from './Icon.svelte';

	// const dispatch = createEventDispatcher();

	// export let convo;
	export let settingsOpen = false;

	// let conversationTemplate = null;

	// async function loadConversationTemplate(name) {
	// 	const template = conversationTemplates[name];
	// 	if (template) {
	// 		if ($convo.model !== template.model) {
	// 			await loadModel(template.model);
	// 		}

	// 		// Create new instances of Message for every template message:
	// 		$convo.messages = await Promise.all(
	// 			template.messages.map(async (msg) => {
	// 				if (msg.contentFn) {
	// 					const functionSchema = await (
	// 						await fetch('http://localhost:8081/tool_schema', { method: 'GET' })
	// 					).text();
	// 					return { role: msg.role, content: msg.contentFn(functionSchema) };
	// 				}

	// 				return msg;
	// 			})
	// 		);
	// 	}

	// 	conversationTemplate = null;
	// 	dispatch('rerender');
	// }
</script>

<aside
	data-sidebar="settings"
	class="{settingsOpen
		? ''
		: 'translate-x-full'} fixed right-0 top-0 z-[100] flex h-full w-[230px] flex-col gap-2 border-l border-slate-200 bg-white px-3 py-4 transition-transform ease-in-out xl:static xl:translate-x-0"
>
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
			class="{!$openrouterAPIKey || $openrouterAPIKey.length === 0
				? 'bg-red-50 placeholder:text-slate-800'
				: ''} rounded-md border border-slate-300 text-sm"
		/></label
	>

	<label class="flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">Groq API Key</span>
		<input
			type="text"
			bind:value={$groqAPIKey}
			placeholder="Enter your API key"
			class="rounded-md border border-slate-300 text-sm"
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
			class="rounded-md border border-slate-300 text-sm"
		/></label
	>
	<Button
		variant="outline"
		class="self-start"
		on:click={async () => {
			const schema = await (
				await fetch('http://localhost:8081/tool_schema', { method: 'GET' })
			).text();
			$toolSchema = JSON.parse(schema);
		}}
	>
		<Icon icon={faSync} class="mr-2 h-3 w-3 text-slate-700" />
		Sync
	</Button>
</aside>
