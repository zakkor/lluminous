<script>
	export let model;
</script>

{#if model && model.provider}
	{#if model.id.startsWith('openai')}
		<img
			src="https://www.svgrepo.com/show/306500/openai.svg"
			loading="lazy"
			class="h-4 w-4 rounded"
			alt=""
		/>
	{:else if model.id.startsWith('anthropic')}
		<img
			src="https://asset.brandfetch.io/idmJWF3N06/ida_TmGBgC.jpeg"
			loading="lazy"
			class="h-4 w-4 rounded"
			alt=""
		/>
	{:else if model.id.startsWith('meta-llama')}
		<img src="/logos/meta.png" loading="lazy" class="h-3.5 w-3.5 rounded" alt="" />
	{:else if model.provider === 'Groq'}
		<img
			src="https://asset.brandfetch.io/idxygbEPCQ/idzCyF-I44.png"
			loading="lazy"
			class="h-3.5 w-3.5 rounded"
			alt=""
		/>
	{/if}
{/if}
