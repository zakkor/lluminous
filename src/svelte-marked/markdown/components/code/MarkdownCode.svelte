<script lang="ts">
	import { flash } from '../../../../actions.js';
	import { feCopy, feFile } from '../../../../feather.js';
	import FilePreview from '../../../../FilePreview.svelte';
	import Icon from '../../../../Icon.svelte';
	import type { MarkdownOptions, Renderers } from '../../markedConfiguration';
	import type { Tokens } from 'marked';

	export let token: Tokens.Code;
	export const options: MarkdownOptions = undefined;
	export const renderers: Renderers = undefined;
	export let message;

	let showCode = message && message.role === 'assistant';
	let attrs: Record<string, any> = {};
	$: if (token.lang) {
		attrs = Object.fromEntries(
			token.lang
				.split(' ')
				.filter((s) => s.indexOf('=') !== -1)
				.map((s) => {
					const [key, value] = s.split('=');
					return [key, value.slice(1, -1)];
				})
		);
	}
</script>

{#if attrs.filename && !showCode}
	<FilePreview filename={attrs.filename} element="button" on:click={() => (showCode = true)} />
{:else if attrs.filename}
	<div class="flex flex-col">
		<div class="flex items-center rounded-t-lg border bg-white py-2 pl-4 pr-2">
			<Icon icon={feFile} class="mr-3 h-4 w-4 text-slate-700" />
			<p class="!my-0 font-mono text-sm font-semibold !leading-none tracking-tight !text-slate-700">
				{attrs.filename}
			</p>
			<div class="ml-auto flex gap-x-3">
				<button
					class="rounded-md border p-1.5 hover:bg-gray-100"
					use:flash
					on:click={(event) => {
						event.currentTarget.dispatchEvent(new CustomEvent('flashSuccess'));
						navigator.clipboard.writeText(token.text);
					}}
				>
					<Icon icon={feCopy} class="h-2.5 w-2.5 text-slate-700" />
				</button>
			</div>
		</div>
		<pre class="!mt-0 !rounded-t-none !border-t-0"><code class={`lang-${token.lang}`}
				>{token.text}</code
			></pre>
	</div>
{:else}
	<pre><code class={`lang-${token.lang}`}>{token.text}</code></pre>
{/if}
