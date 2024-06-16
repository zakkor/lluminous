<script>
	import { createEventDispatcher } from 'svelte';
	import Icon from './Icon.svelte';
	import {
		faBolt,
		faCircleQuestion,
		faHammer,
		faKey,
		faPencil,
	} from '@fortawesome/free-solid-svg-icons';
	import {
		config,
		groqAPIKey,
		mistralAPIKey,
		openaiAPIKey,
		openrouterAPIKey,
		params,
		remoteServer,
		toolSchema,
	} from './stores.js';
	import Button from './Button.svelte';
	import Modal from './Modal.svelte';

	const dispatch = createEventDispatcher();

	export let open = false;
	export let trigger = '';

	let activeTab =
		$openaiAPIKey === '' && $groqAPIKey === '' && $openrouterAPIKey === '' ? 'api-keys' : 'tools';
</script>

<Modal bind:open {trigger}>
	<div class="flex flex-col gap-x-8 gap-y-4 sm:flex-row">
		<div class="flex flex-col">
			<h1 class="mb-4 ml-3 text-xl font-semibold sm:mb-5">Settings</h1>

			<nav class="sm:w-[190px]">
				<ul class="flex flex-col gap-1">
					<li>
						<button
							class="{activeTab === 'api-keys'
								? 'bg-gray-100/70'
								: 'hover:bg-gray-100/70'} flex w-full items-center gap-x-2.5 rounded-lg px-4 py-2.5 text-left text-[13px] font-medium text-slate-700 transition-colors"
							on:click={() => (activeTab = 'api-keys')}
						>
							<Icon icon={faKey} class="h-3 w-3 text-slate-700" />
							API keys
						</button>
					</li>
					<li>
						<button
							class="{activeTab === 'custom-instructions'
								? 'bg-gray-100/70'
								: 'hover:bg-gray-100/70'} flex w-full items-center gap-x-2.5 rounded-lg px-4 py-2.5 text-left text-[13px] font-medium text-slate-700 transition-colors"
							on:click={() => (activeTab = 'custom-instructions')}
						>
							<Icon icon={faPencil} class="h-3 w-3 text-slate-700" />
							Custom instructions
						</button>
					</li>
					<li>
						<button
							class="{activeTab === 'tools'
								? 'bg-gray-100/70'
								: ' hover:bg-gray-100/70'} flex w-full items-center gap-x-2.5 rounded-lg px-4 py-2.5 text-left text-[13px] font-medium text-slate-700 transition-colors"
							on:click={() => (activeTab = 'tools')}
						>
							<Icon icon={faHammer} class="h-3 w-3 text-slate-700" />
							Tool calling
						</button>
					</li>
				</ul>
			</nav>
		</div>
		<div class="flex flex-col gap-y-4 pt-1 sm:w-[400px]">
			{#if activeTab === 'api-keys'}
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">OpenAI API Key</span>
					<input
						type="text"
						bind:value={$openaiAPIKey}
						placeholder="Enter your API key"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
						on:change={() => {
							if ($openaiAPIKey.length === 56 || $openaiAPIKey.length === 0) {
								dispatch('fetchModels');
							}
						}}
					/></label
				>
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">OpenRouter API Key</span>
					<input
						type="text"
						bind:value={$openrouterAPIKey}
						placeholder="Enter your API key"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
						on:change={() => {
							if ($openrouterAPIKey.length === 73 || $openrouterAPIKey.length === 0) {
								dispatch('fetchModels');
							}
						}}
					/></label
				>
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">Groq API Key</span>
					<input
						type="text"
						bind:value={$groqAPIKey}
						placeholder="Enter your API key"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
						on:change={() => {
							if ($groqAPIKey.length === 56 || $groqAPIKey.length === 0) {
								dispatch('fetchModels');
							}
						}}
					/></label
				>
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">Mistral API Key</span>
					<input
						type="text"
						bind:value={$mistralAPIKey}
						placeholder="Enter your API key"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
						on:change={() => {
							if ($mistralAPIKey.length === 32 || $mistralAPIKey.length === 0) {
								dispatch('fetchModels');
							}
						}}
					/></label
				>

				<p class="ml-1 text-center text-xs leading-relaxed text-slate-800 sm:text-left">
					Your API keys are stored entirely locally, on your device, in your browser. They are not
					sent to or stored on any remote server.
				</p>
			{:else if activeTab === 'custom-instructions'}
				<label class="mt-1 flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">Custom instructions</span>
					<textarea
						bind:value={$params.customInstructions}
						rows={10}
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm transition-colors scrollbar-slim focus:border-slate-400 focus:outline-none"
					/></label
				>
			{:else if activeTab === 'tools'}
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px] flex items-center"
						>Server address
						<span class="group relative">
							<Icon icon={faCircleQuestion} class="ml-2 h-3 w-3 text-slate-800" />
							<div
								class="pointer-events-none absolute left-0 top-[calc(100%+8px)] w-[280px] rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-opacity group-hover:opacity-100"
							>
								If you want to use tool calls or local models, you will need to have the lluminous
								server running on your machine.
							</div>
						</span>
					</span>
					<input
						type="text"
						bind:value={$remoteServer.address}
						placeholder="Enter server address (e.g. http://localhost:8081)"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
					/></label
				>

				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">Server password</span>
					<input
						type="password"
						bind:value={$remoteServer.password}
						placeholder="Enter server password"
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
					/></label
				>

				<label class="my-2 flex items-center gap-x-3 text-sm tracking-[-0.25px] text-slate-800">
					<input
						type="checkbox"
						bind:checked={$config.compactToolsView}
						class="h-5 w-5 rounded border-0 !border-slate-300 accent-slate-800 focus:outline-none focus:outline-0 focus:ring-0"
					/>
					Use compact view for tools:
				</label>

				<label class="mt-1 flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px]">Tool schema</span>
					<textarea
						value={JSON.stringify($toolSchema)}
						on:change={(event) => {
							const schema = event.target.value.length > 0 ? event.target.value : '[]';
							$toolSchema = JSON.parse(schema);
						}}
						rows={10}
						class="rounded-lg border border-slate-300 px-3 py-2 text-sm transition-colors scrollbar-slim focus:border-slate-400 focus:outline-none"
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
					<Icon icon={faBolt} class="mr-2 h-3 w-3 text-slate-700" />
					Sync tools from server
				</Button>
			{/if}
		</div>
	</div>
</Modal>
