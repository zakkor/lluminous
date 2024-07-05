<script>
	import { createEventDispatcher, tick } from 'svelte';
	import { scale, slide } from 'svelte/transition';
	import { cubicIn } from 'svelte/easing';
	import { formatModelName } from './convo.js';
	import { toolSchema } from './stores.js';
	import CompanyLogo from './CompanyLogo.svelte';
	import { feCheckCircle, feChevronDown, feImage, feLoader, feTool } from './feather.js';
	import Icon from './Icon.svelte';

	const dispatch = createEventDispatcher();

	export let convo;
	export let models = [];
	export let loadedModel = null;
	export let loadingModel = false;
	let className = '';
	export { className as class };

	let success = false;
	export function modelFinishedLoading() {
		success = true;
		setTimeout(() => {
			success = false;
		}, 1000);
	}

	let id = 'tool-selector';
	let open = false;
	let toolsOpen = false;
	let query = '';

	let collapsed = {};

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
				: 'min-w-[180px] pr-7'} {loadingModel ? 'animate-pulse' : ''} {success
				? 'border-green-200 bg-green-50'
				: ''} flex w-full max-w-[180px] items-center gap-2 rounded-lg border border-slate-300 py-2 pl-3 text-left transition-colors hover:border-slate-400 sm:min-w-[290px]"
			on:click={async () => {
				open = !open;
				toolsOpen = false;
				if (open && innerWidth > 640) {
					await tick();
					inputEl.focus();
				}
			}}
		>
			<CompanyLogo model={convo.model} />
			<div class="flex items-center gap-x-1.5">
				{#if loadingModel}
					<Icon icon={feLoader} class="h-3 shrink-0 animate-spin text-slate-800" />
				{/if}
				{#if convo.model.provider === 'Local' && loadedModel && loadedModel.id === convo.model.id}
					<Icon icon={feCheckCircle} class="h-3 shrink-0 text-slate-800" />
				{/if}
				<p class="line-clamp-1 text-xs text-slate-800">
					{formatModelName(convo.model)}
				</p>
				{#if convo.model.modality === 'image-generation'}
					<Icon icon={feImage} class="mt-px h-3 w-3 text-slate-800" />
				{/if}
			</div>
			<Icon
				icon={feChevronDown}
				class="{$toolSchema.length > 0
					? 'hidden sm:inline'
					: ''} pointer-events-none absolute right-3 top-1/2 h-4 w-4 shrink-0 -translate-y-1/2 text-slate-600"
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
				<Icon icon={feTool} class="m-auto h-3 w-3 fill-slate-700 text-slate-700" />
				{#if convo.tools?.length > 0}
					<span
						class="absolute -bottom-1 -right-1.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-slate-800 text-[10px] text-white"
					>
						{convo.tools.length}
					</span>
				{/if}
			</button>

			{#if toolsOpen && $toolSchema.length > 0}
				<div
					class="absolute left-1/2 top-[calc(100%+6px)] z-10 flex w-[max-content] -translate-x-1/2 rounded-lg"
				>
					<div
						transition:scale={{ opacity: 0, start: 0.98, duration: 100, easing: cubicIn }}
						class="flex h-auto w-auto min-w-[250px] flex-col self-start overflow-y-auto rounded-lg border border-slate-300 bg-white"
					>
						<div class="w-full px-3 pb-2 pt-2.5">
							<div class="mb-3 flex items-center justify-between">
								<h3 class="text-sm font-medium text-slate-800">Tools</h3>
								<button
									class="text-xs text-gray-600 hover:text-gray-800"
									on:click={() => {
										dispatch('clearTools');
									}}
								>
									Clear all
								</button>
							</div>
							<input
								type="text"
								value=""
								placeholder="Search for tools"
								class="w-full rounded-lg border border-slate-300 px-3 py-2 text-xs text-slate-800 transition-colors placeholder:text-gray-500 focus:border-slate-400 focus:outline-none"
								on:input={() => {}}
							/>
						</div>
						<ul class="max-h-[400px] overflow-y-auto pb-1.5 scrollbar-ultraslim">
							{#each $toolSchema as group}
								<div class="relative w-full">
									<label
										class="flex w-full items-center gap-x-3 whitespace-nowrap py-2 pl-4 pr-3 text-left text-xs transition-colors hover:bg-gray-100"
									>
										<input
											type="checkbox"
											checked={group.schema.every((t) => convo.tools.includes(t.function.name))}
											on:change={(event) => {
												const names = group.schema.map((t) => t.function.name);
												if (event.target.checked) {
													dispatch('setTools', names);
												} else {
													dispatch('unsetTools', names);
												}
											}}
											class="h-4 w-4 rounded border-0 !border-slate-300 accent-slate-800 focus:outline-none focus:outline-0 focus:ring-0"
										/>
										<p class="w-full text-xs font-semibold text-slate-800">{group.name}</p>
									</label>

									<button
										on:click={() => (collapsed[group.name] = !collapsed[group.name])}
										class="absolute right-1 top-1/2 flex h-8 w-8 -translate-y-1/2 rounded-full transition-colors hover:bg-gray-100"
									>
										<Icon
											icon={feChevronDown}
											class="{collapsed[group.name]
												? 'rotate-180'
												: ''} m-auto h-4 w-4 text-slate-600 transition-transform"
										/>
									</button>
								</div>

								{#if !collapsed[group.name]}
									<div transition:slide={{ duration: 300 }}>
										{#each group.schema as schema, i}
											<li class="flex w-full">
												<label
													class="flex w-full items-center gap-x-3 whitespace-nowrap py-2 pl-6 pr-3 text-left text-xs transition-colors hover:bg-gray-100"
												>
													<input
														type="checkbox"
														checked={convo.tools.includes(schema.function.name)}
														on:change={(event) => {
															if (event.target.checked) {
																dispatch('setTools', [schema.function.name]);
															} else {
																dispatch('unsetTools', [schema.function.name]);
															}
														}}
														class="h-4 w-4 rounded border-0 !border-slate-300 accent-slate-800 focus:outline-none focus:outline-0 focus:ring-0"
													/>
													<p class="w-full text-xs font-medium text-slate-800">
														{schema.function.name}
													</p>
												</label>
											</li>
										{/each}
									</div>
								{/if}
							{/each}
						</ul>
					</div>
				</div>
			{/if}
		</div>
	{/if}
	{#if open}
		<div
			class="pointer-events-none absolute left-1/2 top-[calc(100%+6px)] z-10 flex w-[max-content] -translate-x-1/2 rounded-lg"
		>
			<div
				transition:scale={{ opacity: 0, start: 0.98, duration: 100, easing: cubicIn }}
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
							dispatch('change', filteredModels[0]);
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
									dispatch('change', model);
									open = false;
									query = '';
								}}
							>
								{#if model.provider === 'Local' && loadedModel && loadedModel.id === model.id}
									<Icon icon={feCheckCircle} class="h-3 text-slate-800" />
								{/if}
								<CompanyLogo {model} />
								{formatModelName(model)}
								{#if model.modality === 'image-generation'}
									<Icon icon={feImage} class="mt-px h-3 w-3 text-slate-800" />
								{/if}
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
