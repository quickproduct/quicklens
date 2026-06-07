<script lang="ts">
	import { FileClock } from 'lucide-svelte';
	import type { AuditLog } from '$api/operations';

	let { logs = [] } = $props<{ logs: AuditLog[] }>();

	function formatDateTime(value: string): string {
		try {
			return new Date(value).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return value;
		}
	}
</script>

<section class="audit-card" aria-labelledby="audit-title">
	<div class="section-header">
		<h3 id="audit-title">Audit Timeline</h3>
		<a href="/settings#audit">Open audit log</a>
	</div>
	<div class="audit-list">
		{#each logs.slice(0, 6) as log}
			<div class="audit-row">
				<FileClock size={15} />
				<div>
					<strong>{log.action}</strong>
					<span>{log.resource}{log.resource_id ? ` · ${log.resource_id.slice(0, 8)}` : ''} · {formatDateTime(log.created_at)}</span>
				</div>
			</div>
		{:else}
			<div class="empty-audit">Audit events will appear after operational actions.</div>
		{/each}
	</div>
</section>

<style>
	.audit-card {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.section-header,
	.audit-row {
		display: flex;
		align-items: center;
	}

	.section-header {
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 0.75rem;
	}

	h3 {
		margin: 0;
		font-size: 0.95rem;
	}

	.section-header a {
		color: var(--ql-accent);
		text-decoration: none;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.audit-list {
		display: flex;
		flex-direction: column;
		gap: 0.625rem;
	}

	.audit-row {
		gap: 0.625rem;
		padding: 0.625rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-bg);
	}

	.audit-row svg {
		color: var(--ql-text-muted);
		flex-shrink: 0;
	}

	.audit-row div {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.audit-row strong {
		font-size: 0.85rem;
	}

	.audit-row span,
	.empty-audit {
		color: var(--ql-text-muted);
		font-size: 0.78rem;
	}

	.empty-audit {
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}
</style>
