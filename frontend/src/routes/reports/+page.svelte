<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuditLogs, getDashboardLayouts, getNotificationRules, getSavedViews, type AuditLog, type SavedView } from '$api/operations';
	import { CalendarClock, Download, FileText, LayoutDashboard, Send } from 'lucide-svelte';

	let savedViews = $state([] as SavedView[]);
	let auditLogs = $state([] as AuditLog[]);
	let notificationRules = $state([] as Record<string, unknown>[]);
	let layouts = $state([] as Record<string, unknown>[]);

	onMount(async () => {
		try {
			const [views, logs, rules, dashboardLayouts] = await Promise.all([
				getSavedViews(),
				getAuditLogs(),
				getNotificationRules(),
				getDashboardLayouts()
			]);
			savedViews = views;
			auditLogs = logs;
			notificationRules = rules;
			layouts = dashboardLayouts;
		} catch (err) {
			console.error(err);
		}
	});
</script>

<div class="reports-page">
	<div class="page-header">
		<div>
			<h2>Reports</h2>
			<p>Export operational evidence, schedule stakeholder summaries, and manage shared dashboard views.</p>
		</div>
		<button class="ql-btn-secondary" type="button">
			<Download size={15} /> Export CSV
		</button>
	</div>

	<div class="report-grid">
		<section class="report-card">
			<FileText size={20} />
			<div>
				<h3>Executive Weekly Report</h3>
				<p>Health score, incidents, SLO burn, trace volume, cost, and model reliability.</p>
				<span>Ready to configure</span>
			</div>
		</section>
		<section class="report-card">
			<CalendarClock size={20} />
			<div>
				<h3>Scheduled Reports</h3>
				<p>{notificationRules.length} notification routes available for delivery.</p>
				<span>Email, Slack, webhook, and status-page targets can be layered in.</span>
			</div>
		</section>
		<section class="report-card">
			<LayoutDashboard size={20} />
			<div>
				<h3>Dashboard Layouts</h3>
				<p>{layouts.length} saved dashboard layouts.</p>
				<span>Default layouts support operations-center screens.</span>
			</div>
		</section>
		<section class="report-card">
			<Send size={20} />
			<div>
				<h3>Shared Views</h3>
				<p>{savedViews.length} shared views across dashboard scopes.</p>
				<span>Use saved filters for teams, tenants, models, and environments.</span>
			</div>
		</section>
	</div>

	<section class="evidence-card">
		<h3>Recent Operational Evidence</h3>
		<div class="evidence-list">
			{#each auditLogs.slice(0, 8) as log}
				<div class="evidence-row">
					<strong>{log.action}</strong>
					<span>{log.resource} · {new Date(log.created_at).toLocaleString()}</span>
				</div>
			{:else}
				<div class="empty-state">No audit evidence has been collected yet.</div>
			{/each}
		</div>
	</section>
</div>

<style>
	.reports-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.page-header,
	.report-card {
		display: flex;
		align-items: center;
	}

	.page-header {
		justify-content: space-between;
		gap: 1rem;
	}

	h2,
	h3,
	p {
		margin: 0;
	}

	.page-header p,
	.report-card p,
	.report-card span,
	.evidence-row span,
	.empty-state {
		color: var(--ql-text-muted);
		font-size: 0.875rem;
	}

	.report-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
	}

	.report-card,
	.evidence-card {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.report-card {
		align-items: flex-start;
		gap: 0.875rem;
	}

	.report-card svg {
		color: var(--ql-accent);
		flex-shrink: 0;
	}

	.report-card div,
	.evidence-list,
	.evidence-row {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.evidence-card h3 {
		margin-bottom: 0.875rem;
	}

	.evidence-row {
		padding: 0.7rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-bg);
	}

	.empty-state {
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}

	@media (max-width: 780px) {
		.report-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
