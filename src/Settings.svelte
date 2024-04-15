<script>
	import { fade, scale } from 'svelte/transition';
	import { cubicIn } from 'svelte/easing';
	import { marked } from 'marked';
	import markedKatex from 'marked-katex-extension';

	marked.use(
		markedKatex({
			throwOnError: false
		})
	);

	export let open = false;
	export let trigger = '';

	// Settings:
	export let raw = false;
	export let grammar = '';

	let source = '';
</script>

<svelte:window
	on:click={(event) => {
		if (
			open &&
			!event.target.closest('[role="dialog"]') &&
			!event.target.closest(`[data-trigger="${trigger}"]`)
		) {
			open = false;
		}
	}}
/>

{#if open}
	<div
		transition:fade={{ duration: 150 }}
		aria-hidden="true"
		class="fixed inset-0 w-screen h-screen bg-black opacity-[0.1]"
	/>
	<div
		role="dialog"
		transition:scale={{ opacity: 0, start: 0.98, duration: 150, easing: cubicIn }}
		class="fixed w-[600px] h-[400px] top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 border rounded-xl bg-white"
	>
		<label class="flex flex-col gap-1.5 select-none">
			Raw:
			<input bind:checked={raw} type="checkbox" class="rounded-md w-6 h-6" />
		</label>

		<p>Grammar:</p>
		<textarea rows={2} bind:value={grammar} />

		<br>

		<textarea bind:value={source} />
		<br>
		<div>
			{@html marked.parse(source)}
		</div>
	</div>
{/if}
