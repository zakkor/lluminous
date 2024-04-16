<script>
	import { createEventDispatcher, onMount } from 'svelte';
	import { detectFormat, isModelLocal, promptFormats } from './convo.js';
	import { conversationTemplates } from './templates.js';
	import { openrouterAPIKey } from './stores.js';

	const dispatch = createEventDispatcher();

	export let convo;
	export let settingsOpen = false;

	let models = [];
	let loading = false;

	async function loadModel(newModel) {
		loading = true;

		const local = isModelLocal(newModel);
		if (local) {
			// For local models, we need to tell the server to load them:
			await fetch(`http://localhost:8081/model`, {
				method: 'POST',
				body: JSON.stringify({
					model: newModel,
				}),
			});
		}
		setModel(newModel);

		loading = false;
	}

	let autodetectedFormat = false;

	function setModel(newModel) {
		$convo.model = newModel;
		const { detected, local } = detectFormat(newModel);
		if (detected) {
			$convo.tmpl = detected;
			autodetectedFormat = true;
		} else {
			autodetectedFormat = false;
		}
		$convo.local = local;
		dispatch('rerender');
	}

	let conversationTemplate = null;

	async function loadConversationTemplate(name) {
		const template = conversationTemplates[name];
		if (template) {
			if ($convo.model !== template.model) {
				await loadModel(template.model);
			}

			// Create new instances of Message for every template message:
			$convo.messages = await Promise.all(
				template.messages.map(async (msg) => {
					if (msg.contentFn) {
						const functionSchema = await (
							await fetch('http://localhost:8081/tool_schema', { method: 'GET' })
						).text();
						return { role: msg.role, content: msg.contentFn(functionSchema) };
					}

					return msg;
				})
			);
		}

		conversationTemplate = null;
		dispatch('rerender');
	}

	onMount(async () => {
		fetch('https://openrouter.ai/api/v1/models', { method: 'GET' })
			.then((response) => response.json())
			.then((json) => {
				models = models.concat(json.data).map((m) => {
					return {
						id: m.id,
						name: m.name,
						local: false,
					};
				});
			})
			.catch((error) => {
				console.error('Error:', error);
			});

		try {
			const modelsData = await (
				await fetch('http://localhost:8081/models', { method: 'GET' })
			).json();
			models = modelsData.models.map((m) => {
				return {
					id: m,
					name: m,
					local: true,
				};
			});

			// Get the model that is currently loaded on the llama.cpp server, if any:
			const modelData = await (
				await fetch('http://localhost:8081/model', { method: 'GET' })
			).json();
			setModel(modelData.model || null);

			// NOTE: For dev only, because HMR wipes the state.
			if ($convo.model === 'openhermes-2.5-mistral-7b.Q5_K_M.gguf') {
				// loadConversationTemplate('OpenHermes-2.5');
				loadConversationTemplate('Markdown');
			} else if ($convo.model === 'Hermes-2-Pro-Mistral-7B.Q6_K.gguf') {
				loadConversationTemplate('Hermes 2 Pro Function Calling');
			}
		} catch (error) {
			console.warn('Local llama.cpp server is not running, running in external mode only.');
		}
	});
</script>

<aside
	data-sidebar="settings"
	class="{settingsOpen
		? ''
		: 'translate-x-full'} fixed right-0 top-0 z-[100] flex h-full w-[230px] flex-col gap-2 border-l border-slate-200 bg-white px-3 py-4 transition-transform ease-in-out xl:static xl:translate-x-0"
>
	<div class="flex flex-col gap-1.5">
		{#if models.length > 0}
			<div class="relative">
				<select
					class="{loading ? 'opacity-50' : ''} w-full rounded-lg border-slate-300 text-sm"
					value={$convo.model}
					disabled={loading}
					on:change={(event) => {
						loadModel(event.target.value);
					}}
				>
					<option value={null} disabled>No model loaded</option>
					{#each models as model}
						<option value={model.id}>{model.name}</option>
					{/each}
				</select>

				{#if loading}
					<div role="status" class="absolute -left-8 top-1 h-5 w-5">
						<svg
							aria-hidden="true"
							class="inline h-5 w-5 animate-spin fill-gray-600 text-gray-200 dark:fill-gray-300 dark:text-gray-600"
							viewBox="0 0 100 101"
							fill="none"
							xmlns="http://www.w3.org/2000/svg"
						>
							<path
								d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z"
								fill="currentColor"
							/>
							<path
								d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z"
								fill="currentFill"
							/>
						</svg>
						<span class="sr-only">Loading...</span>
					</div>
				{/if}
			</div>
		{/if}

		{#if $convo.local}
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
		{/if}
	</div>

	<select
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
	</select>

	<label class="mt-3 flex flex-col text-[10px] uppercase tracking-wide">
		<span class="mb-1.5 ml-[3px]">OpenRouter API Key</span>
		<input
			type="text"
			bind:value={$openrouterAPIKey}
			class="rounded-md border border-slate-300 text-sm"
		/></label
	>
</aside>
