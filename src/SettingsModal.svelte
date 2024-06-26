<script>
	import { createEventDispatcher, tick } from 'svelte';
	import { v4 as uuidv4 } from 'uuid';
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
	import Icon from './Icon.svelte';
	import {
		feEdit3,
		feHardDrive,
		feHelpCircle,
		feKey,
		fePlus,
		feRefreshCw,
		feTool,
	} from './feather.js';
	import ClientToolSetting from './ClientToolSetting.svelte';
	import ClientTool from './ClientTool.svelte';

	const dispatch = createEventDispatcher();

	export let open = false;
	export let trigger = '';

	let activeTab =
		$openaiAPIKey === '' && $groqAPIKey === '' && $openrouterAPIKey === '' ? 'api-keys' : 'tools';

	let addClientToolOpen = false;
	let loadClientTool = null;

	let flashRefreshToolSchema;
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
							<Icon icon={feKey} class="h-3 w-3 text-slate-700" />
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
							<Icon icon={feEdit3} class="h-3 w-3 text-slate-700" />
							Custom instructions
						</button>
					</li>
					<li>
						<button
							class="{activeTab === 'remote-server'
								? 'bg-gray-100/70'
								: ' hover:bg-gray-100/70'} flex w-full items-center gap-x-2.5 rounded-lg px-4 py-2.5 text-left text-[13px] font-medium text-slate-700 transition-colors"
							on:click={() => (activeTab = 'remote-server')}
						>
							<Icon icon={feHardDrive} class="h-3 w-3 text-slate-700" />
							Remote server
						</button>
					</li>
					<li>
						<button
							class="{activeTab === 'tools'
								? 'bg-gray-100/70'
								: ' hover:bg-gray-100/70'} flex w-full items-center gap-x-2.5 rounded-lg px-4 py-2.5 text-left text-[13px] font-medium text-slate-700 transition-colors"
							on:click={() => (activeTab = 'tools')}
						>
							<Icon icon={feTool} class="h-3 w-3 text-slate-700" />
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
			{:else if activeTab === 'remote-server'}
				<label class="flex flex-col text-[10px] uppercase tracking-wide">
					<span class="mb-2 ml-[3px] flex items-center">Server address </span>
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
			{:else if activeTab === 'tools'}
				<label class="my-2 flex items-center gap-x-3 text-sm text-slate-700">
					<input
						type="checkbox"
						bind:checked={$config.explicitToolView}
						class="h-5 w-5 rounded border-0 !border-slate-300 accent-slate-800 focus:outline-none focus:outline-0 focus:ring-0"
					/>
					Use explicit view for tool calls
				</label>

				<div class="mt-1 flex flex-col">
					<span class="mb-3 ml-[3px] flex items-center text-[10px] uppercase tracking-wide"
						>Tool schema
						<span class="group relative">
							<Icon icon={feHelpCircle} class="ml-2 h-3 w-3 text-slate-800" />
							<div
								class="pointer-events-none absolute left-1/2 top-[calc(100%+8px)] w-[280px] -translate-x-1/2 rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-opacity group-hover:opacity-100"
							>
								If you want to use remote tool calls or local models, you will need to have the
								lluminous server running on your machine.
							</div>
						</span>
					</span>
					<p class="ml-[3px] text-sm text-slate-700">
						{$toolSchema.length} tools defined
						{#if $toolSchema.length > 0}
							<button
								class="cursor-pointer text-slate-800 hover:text-slate-900"
								on:click={() => ($toolSchema = $toolSchema.filter((t) => t.clientDefinition))}
							>
								(Clear)
							</button>
						{/if}
					</p>
					<Button
						variant="outline"
						class="self-start mt-3"
						bind:flash={flashRefreshToolSchema}
						on:click={async () => {
							try {
								const schema = await (
									await fetch(`${$remoteServer.address}/tool_schema`, {
										method: 'GET',
										headers: {
											Authorization: `Basic ${$remoteServer.password}`,
										},
									})
								).text();
								const clientToolsSchema = $toolSchema.filter((t) => t.clientDefinition);
								$toolSchema = JSON.parse(schema).concat(clientToolsSchema);
								flashRefreshToolSchema('success');
							} catch (e) {
								flashRefreshToolSchema('error');
								console.error(e);
							}
						}}
					>
						<Icon icon={feRefreshCw} class="mr-2 h-3 w-3 text-slate-700" />
						Sync tools from server
					</Button>
				</div>

				<div>
					<p class="mt-1 flex flex-col text-[10px] uppercase tracking-wide">
						<span class="ml-[3px] flex items-center"
							>Client-side tools
							<span class="group relative">
								<Icon icon={feHelpCircle} class="ml-2 h-3 w-3 text-slate-800" />
								<div
									class="pointer-events-none absolute left-1/2 top-[calc(100%+8px)] w-[280px] -translate-x-1/2 rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-opacity group-hover:opacity-100"
								>
									Allows you to define custom tools that run directly in your browser (JavaScript
									only).
								</div>
							</span>
						</span>
					</p>
					<div class="mt-2 flex flex-col gap-y-2">
						{#each $toolSchema.filter((t) => t.clientDefinition) as tool}
							<ClientTool
								definition={tool.clientDefinition}
								on:edit={async () => {
									addClientToolOpen = true;
									await tick();
									loadClientTool(tool.clientDefinition);
								}}
								on:delete={() => {
									dispatch('disableTool', tool.clientDefinition.name);
									$toolSchema = $toolSchema.filter(
										(t) => t.clientDefinition && t.clientDefinition.id !== tool.clientDefinition.id
									);
								}}
							/>
						{/each}
						<Button
							class="self-start"
							on:click={async () => {
								addClientToolOpen = true;
								await tick();

								loadClientTool({
									name: '',
									description: '',
									arguments: [
										{
											name: '',
											type: 'string',
											description: '',
										},
									],
									body: '',
								});
							}}
						>
							<Icon icon={fePlus} strokeWidth={3} class="mr-2 h-3 w-3 text-slate-700" />
							Add new tool
						</Button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</Modal>

<Modal bind:open={addClientToolOpen} class="!px-8 transition-[all] md:!w-[800px]">
	<ClientToolSetting
		bind:load={loadClientTool}
		on:save={({ detail }) => {
			const i = $toolSchema.findIndex(
				(tool) => tool.clientDefinition && tool.clientDefinition.id === detail.clientDefinition.id
			);
			if (i > -1) {
				$toolSchema[i] = detail;
			} else {
				$toolSchema = [
					...$toolSchema,
					{ ...detail, clientDefinition: { id: uuidv4(), ...detail.clientDefinition } },
				];
			}
			addClientToolOpen = false;
		}}
	/>
</Modal>
