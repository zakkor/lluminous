<script lang="ts">
	import type { Tokens } from 'marked';
	import type { MarkdownOptions, Renderers } from '../markedConfiguration';

	export let token: Tokens.HTML;
	export const options: MarkdownOptions = undefined;
	export const renderers: Renderers = undefined;
</script>

<span class="{token.block ? 'block' : 'inline'} whitespace-pre-wrap [overflow-wrap:anywhere]"
	>{token.text}</span
>
