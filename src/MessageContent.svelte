<script>
	import Markdown from './svelte-marked/markdown/Markdown.svelte';

	export let message;
</script>

{#if message.error}
	<span class="text-slate-600">{message.error}</span>
{:else if message.content}
	<div
		class="markdown prose prose-slate flex w-full max-w-none flex-col break-words prose-p:whitespace-pre-wrap prose-p:text-slate-800 prose-a:[overflow-wrap:anywhere] prose-code:[overflow-wrap:anywhere] prose-pre:my-4 prose-pre:whitespace-pre-wrap prose-pre:[overflow-wrap:anywhere] prose-pre:border prose-pre:border-slate-200 prose-pre:bg-white prose-pre:text-slate-800 prose-img:mb-2"
	>
		{#if message.contentParts}
			{#each message.contentParts as part}
				<img
					src={part.image_url.url}
					alt=""
					class="max-h-[400px] w-min rounded-lg object-contain object-[0]"
				/>
			{/each}
		{/if}
		<Markdown source={message.content} {message} />
	</div>
{:else if message.generatedImageUrl}
	<img
		src={message.generatedImageUrl}
		alt=""
		class="max-h-[400px] w-min rounded-lg object-contain object-[0]"
	/>
{/if}
