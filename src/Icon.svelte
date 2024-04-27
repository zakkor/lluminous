<script>
	let className = null;
	export { className as class };
	export let icon;

	$: i = (icon && icon.icon) || [0, 0, '', [], ''];
</script>

{#if i[4]}
	<svg
		class={className}
		viewBox="0 0 {i[0]} {i[1]}"
		aria-hidden="true"
		role="img"
		xmlns="http://www.w3.org/2000/svg"
	>
		{#if typeof i[4] == 'string'}
			<path d={i[4]} fill="currentColor" />
		{:else}
			<path d={i[4][0]} fill="currentColor" fill-opacity="1.0" />
			<path d={i[4][1]} fill="currentColor" fill-opacity="0.4" />
		{/if}
	</svg>
{/if}
