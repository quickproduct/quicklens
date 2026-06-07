<script lang="ts">
	import { AlertCircle, Check, ExternalLink } from 'lucide-svelte';
	import type { AlertSummary } from '$api/dashboard';
	import Badge from '$components/shared/Badge.svelte';

	let { alerts = [], onAcknowledge, onDeclareIncident } = $props<{
		alerts: AlertSummary[];
		onAcknowledge?: (id: string) => void;
		onDeclareIncident?: (alert: AlertSummary) => void;
	}>();

	function formatDateTime(value: string): string {
		try {
			return new Date(value).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return value;
		}
	}
</script>

<section class="alert-inbox" aria-labelledby="alert-inbox-title">
	<div class="section-header">
		<div>
			<h3 id="alert-inbox-title">Alert Center</h3>
			<p>Critical issues first, with direct incident escalation.</p>
		</div>
		<a href="/alerts">View all</a>
	</div>

	<div class="alert-list">
		{#each alerts.slice(0, 5) as alert}
			<article class="alert-row {alert.severity}">
				<div class="alert-copy">
					<div class="alert-meta">
						<Badge
							text={alert.severity}
							variant={alert.severity === 'critical' ? 'danger' : alert.severity === 'warning' ? 'warning' : 'info'}
						/>
						<span>{formatDateTime(alert.created_at)}</span>
					</div>
					<p>{alert.message}</p>
					<div class="alert-context">
						<span>{alert.status || 'open'}</span>
						<span>{alert.model_id || alert.service_id || 'Unassigned service'}</span>
					</div>
				</div>
				<div class="alert-actions">
					{#if onAcknowledge}
						<button type="button" class="icon-action" onclick={() => onAcknowledge?.(alert.id)} title="Acknowledge alert">
							<Check size={15} />
						</button>
					{/if}
					{#if onDeclareIncident}
						<button type="button" class="icon-action primary" onclick={() => onDeclareIncident?.(alert)} title="Declare incident">
							<ExternalLink size={15} />
						</button>
					{/if}
				</div>
			</article>
		{:else}
			<div class="empty-alerts">
				<AlertCircle size={18} />
				No active alerts. The current operating window is clear.
			</div>
		{/each}
	</div>
</section>

<style>
	.alert-inbox {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.section-header,
	.alert-row,
	.alert-meta,
	.alert-context,
	.alert-actions,
	.empty-alerts {
		display: flex;
		align-items: center;
	}

	.section-header {
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 0.75rem;
	}

	h3,
	p {
		margin: 0;
	}

	h3 {
		font-size: 0.95rem;
	}

	.section-header p,
	.alert-meta span,
	.alert-context,
	.empty-alerts {
		color: var(--ql-text-muted);
		font-size: 0.8rem;
	}

	.section-header a {
		color: var(--ql-accent);
		text-decoration: none;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.alert-list {
		display: flex;
		flex-direction: column;
		gap: 0.625rem;
	}

	.alert-row {
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.75rem;
		border: 1px solid var(--ql-border);
		border-left-width: 4px;
		border-radius: 6px;
		background: var(--ql-bg);
	}

	.alert-row.critical { border-left-color: var(--ql-critical); }
	.alert-row.warning { border-left-color: var(--ql-warning); }
	.alert-row.info { border-left-color: var(--ql-info); }

	.alert-copy {
		min-width: 0;
	}

	.alert-copy p {
		margin-top: 0.35rem;
		font-size: 0.875rem;
		line-height: 1.35;
	}

	.alert-meta,
	.alert-context {
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.alert-context {
		margin-top: 0.5rem;
	}

	.alert-context span {
		padding: 0.2rem 0.4rem;
		border: 1px solid var(--ql-border);
		border-radius: 999px;
	}

	.alert-actions {
		gap: 0.35rem;
		flex-shrink: 0;
	}

	.icon-action {
		width: 30px;
		height: 30px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-surface-2);
		color: var(--ql-text-muted);
		cursor: pointer;
	}

	.icon-action:hover,
	.icon-action:focus-visible {
		color: var(--ql-text);
		border-color: var(--ql-accent);
		outline: none;
	}

	.icon-action.primary {
		color: var(--ql-accent);
	}

	.empty-alerts {
		gap: 0.5rem;
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}
</style>
