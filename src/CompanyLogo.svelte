<script>
	export let model;
	export let size = 'h-3.5 w-3.5';
	export let rounded = null;
</script>

{#if model && model.provider}
	{#if model.id.startsWith('openai')}
		<img src="/logos/openai.ico" loading="lazy" class="{size} {rounded || 'rounded-sm'}" alt="" />
	{:else if model.id.startsWith('anthropic')}
		<img
			src="/logos/anthropic.jpeg"
			loading="lazy"
			class="{size} {rounded || 'rounded-sm'}"
			alt=""
		/>
	{:else if model.id.startsWith('meta-llama')}
		<img src="/logos/meta.png" loading="lazy" class={size} alt="" />
	{:else if model.id.startsWith('mistralai')}
		<img src="/logos/mistral.png" loading="lazy" class={size} alt="" />
	{:else if model.id.startsWith('cohere')}
		<img src="/logos/cohere.png" loading="lazy" class={size} alt="" />
	{:else if model.provider === 'Groq'}
		<img src="/logos/groq.png" loading="lazy" class="{size} {rounded || 'rounded-sm'}" alt="" />
	{:else if model.id.startsWith('nous')}
		<img src="/logos/nous.png" loading="lazy" class={size} alt="" />
	{/if}
{/if}
