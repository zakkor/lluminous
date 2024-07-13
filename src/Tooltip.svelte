<script>
	import Icon from './Icon.svelte';
	import { feHelpCircle } from './feather.js';

	export let icon = feHelpCircle;

	let className = '';
	export { className as class };
</script>

<span class="{className} group relative">
	<Icon {icon} class="h-3 w-3 text-slate-800" />
	<div
		class="pointer-events-none absolute left-1/2 top-[calc(100%+8px)] flex w-[300px] -translate-x-1/2 scale-90 rounded-lg bg-black px-3 py-2 text-xs font-normal normal-case tracking-normal text-white opacity-0 transition-[opacity,transform] ease-in-out group-hover:scale-100 group-hover:opacity-100"
	>
		<slot />
	</div>
</span>
