<script>
	export let filename;
	export let element = 'div';
	let className = 'w-full px-6 py-6 my-auto';
	export { className as class };
	export let outerClass = 'aspect-[4/5] max-h-[160px] relative';
	export let filenameClass = 'line-clamp-2 leading-relaxed';
	export let badgeClass = '';
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<svelte:element
	this={element}
	class="{outerClass} group/file bg-white flex flex-col gap-2 self-start rounded-lg border border-slate-200 transition-colors hover:border-slate-300"
	on:click
>
	<div class="absolute right-[-1px] top-[-1px] bg-white">
		<svg
			viewBox="0 0 16 16"
			fill="#000000"
			xmlns="http://www.w3.org/2000/svg"
			class="block h-5 w-5 flex-shrink-0 drop-shadow"
			><path
				d="M0 0L16 16H0V0Z"
				class="fill-white stroke-slate-200 transition-colors group-hover/file:stroke-slate-300"
				stroke-width="0.5"
			></path><path
				d="M-1 -1L17 17"
				class="stroke-slate-200 transition-colors group-hover/file:stroke-slate-300"
				stroke-width="0.5"
			></path></svg
		>
	</div>
	<div class="{className} flex flex-col gap-2">
		<p class="{filenameClass} font-mono text-sm font-semibold">
			{filename.slice(0, filename.indexOf('.'))}
		</p>
		<p
			class="{badgeClass} !m-0 !mx-auto whitespace-nowrap rounded-md bg-black px-1 leading-none py-1 font-mono text-[8px] font-semibold uppercase tracking-widest !text-white"
		>
			{filename.slice(filename.indexOf('.') + 1)}
		</p>
	</div>
</svelte:element>
