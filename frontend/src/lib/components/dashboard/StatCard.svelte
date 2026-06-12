<script lang="ts">
	import { ArrowUpRight, ArrowDownRight } from 'lucide-svelte';
	import Card from '$components/shared/Card.svelte';

	let {
		title = '',
		value = '',
		icon: IconComponent = null,
		trend = '',
		trendUp = true,
		loading = false
	}: {
		title: string;
		value: string | number;
		icon?: any;
		trend?: string;
		trendUp?: boolean;
		loading?: boolean;
	} = $props();
</script>

<Card gradient padding="p-5">
	<div class="mb-2 flex items-start justify-between">
		<div class="flex flex-col gap-1">
			{#if loading}
				<div class="h-5 w-24 animate-pulse rounded bg-ql-surface-2"></div>
				<div class="mt-1 h-8 w-32 animate-pulse rounded bg-ql-surface-2"></div>
			{:else}
				<span class="text-sm font-medium text-ql-text-muted">{title}</span>
				<span class="text-[1.75rem] font-bold tracking-tight text-ql-text">{value}</span>
			{/if}
		</div>
		<div class="hidden sm:flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-ql-border bg-ql-surface-2 text-ql-accent">
			{#if loading}
				<div class="h-6 w-6 animate-pulse rounded-full bg-ql-border"></div>
			{:else if IconComponent}
				<IconComponent size={24} class="stat-icon" />
			{/if}
		</div>
	</div>
	{#if loading}
		<div class="mt-1 flex items-center gap-1.5">
			<div class="h-3 w-12 animate-pulse rounded bg-ql-surface-2"></div>
			<div class="h-3 w-20 animate-pulse rounded bg-ql-surface-2"></div>
		</div>
	{:else if trend}
		<div class="flex items-center gap-1.5 text-xs">
			<span 
				class="flex items-center gap-0.5 font-semibold {trendUp ? 'text-ql-success' : 'text-ql-danger'}" 
				title="Change compared to previous day" 
				aria-label="Trend: {trendUp ? 'Up' : 'Down'} {trend}"
			>
				{#if trendUp}
					<ArrowUpRight size={14} />
				{:else}
					<ArrowDownRight size={14} />
				{/if}
				{trend}
			</span>
			<span class="text-ql-text-muted" title="Compared to previous day" aria-label="vs yesterday">vs yesterday</span>
		</div>
	{/if}
</Card>
