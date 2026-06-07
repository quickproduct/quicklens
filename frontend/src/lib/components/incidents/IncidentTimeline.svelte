<script lang="ts">
	import { CheckCircle2, CircleDot, Clock } from 'lucide-svelte';
	import type { IncidentEvent } from '$api/operations';

	let { events = [] } = $props<{ events: IncidentEvent[] }>();

	function formatDateTime(value: string): string {
		try {
			return new Date(value).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return value;
		}
	}
</script>

<div class="timeline">
	{#each events as event}
		<div class="timeline-row">
			<span class="marker">
				{#if event.event_type === 'resolved'}
					<CheckCircle2 size={15} />
				{:else}
					<CircleDot size={15} />
				{/if}
			</span>
			<div>
				<strong>{event.message || event.event_type}</strong>
				<span><Clock size={13} /> {formatDateTime(event.created_at)}</span>
			</div>
		</div>
	{:else}
		<div class="empty-timeline">No incident activity yet.</div>
	{/each}
</div>

<style>
	.timeline {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.timeline-row {
		display: grid;
		grid-template-columns: 24px minmax(0, 1fr);
		gap: 0.625rem;
	}

	.marker {
		width: 24px;
		height: 24px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--ql-accent-subtle);
		color: var(--ql-accent);
	}

	.timeline-row div {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--ql-border);
	}

	.timeline-row strong {
		font-size: 0.875rem;
	}

	.timeline-row span:last-child {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		color: var(--ql-text-muted);
		font-size: 0.78rem;
	}

	.empty-timeline {
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
		color: var(--ql-text-muted);
		font-size: 0.85rem;
	}
</style>
