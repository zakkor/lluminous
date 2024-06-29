<script lang="ts">
	export let variant = 'outline';
	let className = '';
	export { className as class };
	export let trigger = null;

	const variantClass = {
		outline:
			'border bg-white border-slate-300 hover:border-slate-400 transition-colors text-slate-700 text-sm px-3 py-1.5 rounded-lg',
		transparent: 'rounded-full p-3 transition-colors hover:bg-gray-100',
	}[variant];

	let flashClass = '';
	export function flash(kind = 'success') {
		flashClass = {
			success: '!bg-green-200 !border-green-300',
			error: '!bg-red-100 !border-red-300',
		}[kind];
		setTimeout(() => {
			flashClass = '';
		}, 1000);
	}
</script>

<button
	data-trigger={trigger}
	class="{className} {variantClass} {flashClass} flex items-center"
	on:click
>
	<slot />
</button>
