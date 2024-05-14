<script>
	import { faChevronDown, faHammer } from '@fortawesome/free-solid-svg-icons';
	import Icon from './Icon.svelte';
	import { toolSchema, tools } from './stores.js';
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import CompanyLogo from './CompanyLogo.svelte';
	import { formatModelName } from './convo.js';

	export let convo;
	export let models = [];
	let className = '';
	export { className as class };

	let id = 'tool-selector';

	let open = false;
	let toolsOpen = false;
	let query = '';

	$: filteredModels =
		query.length > 0
			? models.filter((model) => model.name.toLowerCase().includes(query.toLowerCase()))
			: models;

	let inputEl;
	let innerWidth;
</script>

<svelte:window
	bind:innerWidth
	on:click={(event) => {
		if (!event.target.closest(`#${id}`)) {
			open = false;
			toolsOpen = false;
			query = '';
		}
	}}
	on:touchstart={(event) => {
		if (!event.target.closest(`#${id}`)) {
			open = false;
			toolsOpen = false;
			query = '';
		}
	}}
/>

<div {id} class="{className} flex gap-1.5 sm:gap-x-2">
	<div class="relative">
		<button
			class="{$toolSchema.length > 0
				? 'min-w-[140px] pr-3 sm:pr-7'
				: 'min-w-[180px] pr-7'} flex w-full max-w-[180px] items-center gap-2 rounded-lg border border-slate-300 py-2 pl-3 text-left transition-colors hover:border-slate-400 sm:min-w-[240px]"
			on:click={async () => {
				open = !open;
				toolsOpen = false;
				if (open && innerWidth > 640) {
					await tick();
					inputEl.focus();
				}
			}}
		>
			<CompanyLogo model={$convo.model} />
			<p class="line-clamp-1 text-xs text-slate-800">
				{formatModelName($convo.model)}
			</p>
			<Icon
				icon={faChevronDown}
				class="{$toolSchema.length > 0
					? 'hidden sm:inline'
					: ''} pointer-events-none absolute right-3 top-1/2 h-2.5 w-2.5 shrink-0 -translate-y-1/2 text-slate-600"
			/>
		</button>
	</div>
	{#if $toolSchema.length > 0}
		<div class="sm:absolute sm:-right-2 sm:translate-x-full">
			<button
				on:click={() => {
					toolsOpen = !toolsOpen;
					open = false;
				}}
				class="flex h-[34px] w-[34px] rounded-lg border border-slate-300 transition-colors hover:border-slate-400"
			>
				<Icon icon={faHammer} class="m-auto h-3 w-3 text-slate-700" />
				{#if $tools.length > 0}
					<span
						class="absolute -bottom-1 -right-1.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-teal-300 text-[10px]"
					>
						{$tools.length}
					</span>
				{/if}
			</button>

			{#if toolsOpen && $toolSchema.length > 0}
				<div
					class="absolute left-1/2 top-[calc(100%+6px)] z-10 flex w-[max-content] -translate-x-1/2 rounded-lg"
				>
					<ul
						transition:fade={{ duration: 100 }}
						class="flex h-auto w-auto min-w-[250px] flex-col self-start overflow-y-auto rounded-lg border border-slate-300 bg-white py-1.5 scrollbar-none"
					>
						<!-- <p class="mb-2 ml-3 mt-1 text-xs font-medium text-slate-800">Tool use:</p> -->
						{#each $toolSchema as schema, i}
							<li class="flex w-full">
								<label
									class="flex w-full items-center gap-x-3 whitespace-nowrap px-3 py-2 text-left text-xs transition-colors hover:bg-gray-100"
								>
									<p class="w-full text-xs font-semibold text-slate-800">{schema.function.name}</p>
									<input
										type="checkbox"
										checked={$tools.includes(schema.function.name)}
										on:change={() => {
											if ($tools.includes(schema.function.name)) {
												tools.update((arr) => arr.filter((item) => item !== schema.function.name));
											} else {
												tools.update((arr) => [...arr, schema.function.name]);
											}
										}}
										class="h-5 w-5 rounded border-0 !border-slate-300 accent-slate-800 focus:outline-none focus:outline-0 focus:ring-0"
									/>
								</label>
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		</div>
	{/if}
	{#if open}
		<div
			class="pointer-events-none absolute left-1/2 top-[calc(100%+6px)] z-10 flex w-[max-content] -translate-x-1/2 rounded-lg"
		>
			<div
				transition:fade={{ duration: 100 }}
				class="pointer-events-auto min-w-[240px] rounded-lg border border-slate-300 bg-white"
			>
				<input
					bind:this={inputEl}
					type="text"
					placeholder="Search models..."
					class="w-full appearance-none rounded-t-lg border-0 border-b border-slate-300 px-4 py-2.5 text-sm text-slate-800 placeholder:text-slate-700 focus:outline-none focus:ring-0"
					bind:value={query}
					on:keydown={(event) => {
						if (event.key === 'Enter') {
							$convo.model = filteredModels[0];
							open = false;
							query = '';
						}
					}}
				/>
				<ul class="flex max-h-[400px] w-full flex-col overflow-y-auto scrollbar-none">
					{#each filteredModels as model, i}
						<li>
							<button
								class="flex w-full items-center gap-2 whitespace-nowrap px-3 py-2 text-left text-xs transition-colors hover:bg-gray-100"
								on:click={() => {
									$convo.model = model;
									open = false;
									query = '';
								}}
							>
								<CompanyLogo {model} />
								{formatModelName(model)}
							</button>
						</li>
					{:else}
						<div class="px-4 py-2.5 text-sm">No results.</div>
					{/each}
				</ul>
			</div>
		</div>
	{/if}
</div>
