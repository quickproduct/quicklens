<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		padding = 'p-4',
		gradient = false
	}: {
		children: Snippet;
		class?: string;
		padding?: string;
		gradient?: boolean;
	} = $props();
</script>

<div class="relative flex flex-col overflow-hidden rounded-ql-md border border-ql-border bg-ql-surface shadow-sm {padding} {className}">
	{#if gradient}
		<div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-[rgba(16,185,129,0.03)] to-transparent"></div>
	{/if}
	<div class="relative z-10 flex h-full flex-col">
		{@render children()}
	</div>
</div>
