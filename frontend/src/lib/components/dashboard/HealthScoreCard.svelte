<script lang="ts">
	import { Activity, AlertTriangle, ShieldCheck } from 'lucide-svelte';

	let { score = 0, criticalAlerts = 0, activeIncidents = 0 } = $props<{
		score: number;
		criticalAlerts: number;
		activeIncidents: number;
	}>();

	let state = $derived(score >= 90 ? 'healthy' : score >= 70 ? 'warning' : 'critical');
	let label = $derived(state === 'healthy' ? 'Healthy' : state === 'warning' ? 'Watch' : 'Critical');
</script>

<section class="health-card {state}" aria-label="Operational health score">
	<div class="health-main">
		<div class="score-ring">
			{#if state === 'healthy'}
				<ShieldCheck size={24} />
			{:else}
				<AlertTriangle size={24} />
			{/if}
		</div>
		<div>
			<p class="eyebrow">Health Score</p>
			<div class="score-row">
				<strong>{score}</strong>
				<span>{label}</span>
			</div>
		</div>
	</div>
	<div class="health-meta">
		<span><AlertTriangle size={14} /> {criticalAlerts} critical</span>
		<span><Activity size={14} /> {activeIncidents} incidents</span>
	</div>
</section>

<style>
	.health-card {
		display: flex;
		align-items: stretch;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.health-card.healthy { border-left: 4px solid var(--ql-healthy); }
	.health-card.warning { border-left: 4px solid var(--ql-warning); }
	.health-card.critical { border-left: 4px solid var(--ql-critical); }

	.health-main,
	.health-meta,
	.score-row,
	.health-meta span {
		display: flex;
		align-items: center;
	}

	.health-main {
		gap: 0.875rem;
	}

	.score-ring {
		width: 44px;
		height: 44px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		background: var(--ql-surface-2);
		color: var(--ql-accent);
	}

	.eyebrow {
		margin: 0;
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		font-weight: 700;
	}

	.score-row {
		gap: 0.75rem;
	}

	.score-row strong {
		font-size: 2rem;
		line-height: 1;
	}

	.score-row span {
		font-size: 0.875rem;
		color: var(--ql-text-muted);
	}

	.health-meta {
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 0.5rem;
	}

	.health-meta span {
		gap: 0.35rem;
		padding: 0.375rem 0.5rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		color: var(--ql-text-muted);
		font-size: 0.8rem;
	}

	@media (max-width: 720px) {
		.health-card {
			flex-direction: column;
		}

		.health-meta {
			justify-content: flex-start;
		}
	}
</style>
