<script>
	import { toolSchema } from './stores.js';

	console.log(`$toolSchema:`, $toolSchema);
</script>

<select class="max-w-[180px] rounded-lg border-slate-200 py-2 !pr-7 text-xs md:max-w-[250px]">
  <option value={null}>No tool</option>
	{#each $toolSchema as schema}
		<option value={schema.function.name}>{schema.function.name}</option>
	{/each}
</select>
