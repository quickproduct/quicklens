<script lang="ts">
	import { Box, CheckCircle2, Server, XCircle } from 'lucide-svelte';
	import type { ModelUsageSummary } from '$api/dashboard';

	let { models = [], modelsOnline = 0, modelsTotal = 0 } = $props<{
		models: ModelUsageSummary[];
		modelsOnline: number;
		modelsTotal: number;
	}>();
</script>

<section class="status-card" aria-labelledby="status-grid-title">
	<div class="section-header">
		<div>
			<h3 id="status-grid-title">Product Status Grid</h3>
			<p>{modelsOnline} of {modelsTotal} models online</p>
		</div>
		<span class="overall" class:degraded={modelsTotal > 0 && modelsOnline < modelsTotal}>
			{#if modelsTotal === 0 || modelsOnline < modelsTotal}
				<XCircle size={15} />
				Degraded
			{:else}
				<CheckCircle2 size={15} />
				Operational
			{/if}
		</span>
	</div>

	<div class="status-grid">
		{#each models.slice(0, 6) as model}
			<div class="status-item">
				<Server size={16} />
				<div>
					<strong>{model.model_name}</strong>
					<span>{model.provider} · {model.request_count.toLocaleString()} requests</span>
				</div>
			</div>
		{:else}
			<div class="empty-inline">
				<Box size={18} />
				No model traffic has been recorded.
			</div>
		{/each}
	</div>
</section>

<style>
	.status-card {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.section-header,
	.overall,
	.status-item,
	.empty-inline {
		display: flex;
		align-items: center;
	}

	.section-header {
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 0.875rem;
	}

	h3,
	p {
		margin: 0;
	}

	h3 {
		font-size: 0.95rem;
	}

	p,
	.status-item span,
	.empty-inline {
		color: var(--ql-text-muted);
		font-size: 0.8rem;
	}

	.overall {
		gap: 0.35rem;
		padding: 0.35rem 0.5rem;
		border-radius: 6px;
		background: rgba(34, 197, 94, 0.1);
		color: var(--ql-healthy);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.overall.degraded {
		background: rgba(245, 158, 11, 0.1);
		color: var(--ql-warning);
	}

	.status-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
	}

	.status-item {
		gap: 0.625rem;
		min-width: 0;
		padding: 0.625rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-bg);
	}

	.status-item svg {
		color: var(--ql-accent);
		flex-shrink: 0;
	}

	.status-item div {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.status-item strong,
	.status-item span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.empty-inline {
		grid-column: 1 / -1;
		gap: 0.5rem;
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}

	@media (max-width: 720px) {
		.status-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
