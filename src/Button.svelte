<script lang="ts">
	export let variant = 'secondary';
	let className = '';
	export { className as class };

	const variantClass = {
		outline:
			'border bg-white border-slate-300 hover:border-slate-400 transition-colors text-slate-600 px-4 py-2 rounded-lg',
	}[variant];
</script>

<button class="{className} {variantClass} flex items-center" on:click>
	<slot />
</button>
