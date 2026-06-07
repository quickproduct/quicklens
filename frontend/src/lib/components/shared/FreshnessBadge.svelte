<script lang="ts">
	import { Clock, Wifi, WifiOff } from 'lucide-svelte';

	let { label = 'No data loaded', connected = false } = $props<{
		label: string;
		connected: boolean;
	}>();

	let stale = $derived(label.toLowerCase().includes('stale') || label === 'No data loaded');
</script>

<span class="freshness" class:stale title={connected ? 'Live data channel connected' : 'Live data channel offline'}>
	{#if connected}
		<Wifi size={14} />
	{:else}
		<WifiOff size={14} />
	{/if}
	<Clock size={14} />
	{label}
</span>

<style>
	.freshness {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.375rem 0.625rem;
		border: 1px solid rgba(34, 197, 94, 0.28);
		border-radius: 6px;
		background: rgba(34, 197, 94, 0.08);
		color: var(--ql-healthy);
		font-size: 0.8rem;
		font-weight: 600;
		white-space: nowrap;
	}

	.freshness.stale {
		border-color: rgba(245, 158, 11, 0.3);
		background: rgba(245, 158, 11, 0.08);
		color: var(--ql-warning);
	}
</style>
