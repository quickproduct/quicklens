<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { acknowledgeAlert } from '$api/alerts';
	import type { AlertSummary } from '$api/dashboard';
	import { createIncident, getAuditLogs, getSavedViews, type AuditLog, type SavedView } from '$api/operations';
	import AlertInbox from '$components/dashboard/AlertInbox.svelte';
	import AuditTimeline from '$components/dashboard/AuditTimeline.svelte';
	import CostChart from '$components/dashboard/CostChart.svelte';
	import HealthScoreCard from '$components/dashboard/HealthScoreCard.svelte';
	import StatCard from '$components/dashboard/StatCard.svelte';
	import StatusGrid from '$components/dashboard/StatusGrid.svelte';
	import TokenChart from '$components/dashboard/TokenChart.svelte';
	import FreshnessBadge from '$components/shared/FreshnessBadge.svelte';
	import SavedViewBar from '$components/shared/SavedViewBar.svelte';
	import { dashboardAutoRefresh, dashboardData, dashboardError, dashboardFreshnessLabel, dashboardLoading, refreshDashboard, scheduleDashboardRefresh } from '$stores/dashboard';
	import { addToast, wsConnected } from '$stores/ui';
	import { createWebSocket } from '$lib/websocket/client';
	import { Activity, Bell, Coins, Cpu, GitBranch, RefreshCw, ShieldCheck, Siren, Target } from 'lucide-svelte';

	let wsClient: ReturnType<typeof createWebSocket> | null = null;
	let auditLogs = $state([] as AuditLog[]);
	let savedViews = $state([] as SavedView[]);

	onMount(() => {
		refreshDashboard();
		loadSupportingData();

		wsClient = createWebSocket('traces', (msg: any) => {
			if (!$dashboardAutoRefresh) return;
			if (msg && (msg.type === 'trace_ingested' || msg.type === 'update')) {
				scheduleDashboardRefresh(750);
			}
		});
	});

	onDestroy(() => {
		wsClient?.close();
	});

	async function loadSupportingData() {
		try {
			const [views, logs] = await Promise.all([getSavedViews('overview'), getAuditLogs()]);
			savedViews = views;
			auditLogs = logs;
		} catch (err) {
			console.error(err);
		}
	}

	async function handleAcknowledge(id: string) {
		try {
			await acknowledgeAlert(id);
			addToast('Alert acknowledged.', 'success');
			refreshDashboard();
		} catch (err) {
			console.error(err);
			addToast('Failed to acknowledge alert.', 'error');
		}
	}

	async function handleDeclareIncident(alert: AlertSummary) {
		try {
			await createIncident({
				title: alert.message,
				severity: alert.severity,
				alert_id: alert.id,
				model_id: alert.model_id,
				service_id: alert.service_id,
				runbook_url: alert.runbook_url,
				summary: 'Declared from active alert on the overview dashboard.'
			});
			addToast('Incident declared.', 'success');
			refreshDashboard();
			loadSupportingData();
		} catch (err) {
			console.error(err);
			addToast('Failed to declare incident.', 'error');
		}
	}

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.00';
		return `$${cost.toFixed(4)}`;
	}

	function formatLatency(ms: number): string {
		if (ms === 0 || !ms) return '0ms';
		return `${ms.toFixed(0)}ms`;
	}
</script>

<div class="overview-page">
	<div class="overview-toolbar">
		<SavedViewBar views={savedViews} activeScope="overview" />
		<div class="toolbar-actions">
			<label class="auto-refresh">
				<input type="checkbox" bind:checked={$dashboardAutoRefresh} />
				Auto-refresh
			</label>
			<FreshnessBadge label={$dashboardFreshnessLabel} connected={$wsConnected} />
			<button class="ql-btn-secondary refresh-btn" type="button" onclick={() => refreshDashboard()} disabled={$dashboardLoading}>
				<RefreshCw size={15} />
				Refresh
			</button>
		</div>
	</div>

	{#if $dashboardLoading && !$dashboardData}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading operations overview...</span>
		</div>
	{:else if $dashboardError}
		<div class="error-state">{$dashboardError}</div>
	{:else if $dashboardData}
		<section class="hero-grid">
			<HealthScoreCard
				score={$dashboardData.health_score}
				criticalAlerts={$dashboardData.critical_alert_count}
				activeIncidents={$dashboardData.active_incident_count}
			/>
			<div class="slo-card">
				<div class="slo-icon"><Target size={22} /></div>
				<div>
					<p class="eyebrow">SLO Burn Summary</p>
					<strong>{$dashboardData.slo_burn_summary.attainment_percent.toFixed(2)}%</strong>
					<span>Target {$dashboardData.slo_burn_summary.target_percent.toFixed(1)}% · {$dashboardData.slo_burn_summary.error_budget_remaining.toFixed(0)}% budget left</span>
				</div>
			</div>
			<div class="incident-card">
				<Siren size={22} />
				<div>
					<p class="eyebrow">Active Incidents</p>
					<strong>{$dashboardData.active_incident_count}</strong>
					<a href="/incidents">Open response queue</a>
				</div>
			</div>
		</section>

		<div class="stats-row">
			<StatCard title="Traces Today" value={$dashboardData.total_traces_today} icon={GitBranch} />
			<StatCard title="Tokens Processed" value={$dashboardData.total_tokens_today.toLocaleString()} icon={Cpu} />
			<StatCard title="Cost Today" value={formatCost($dashboardData.total_cost_today)} icon={Coins} />
			<StatCard title="Avg Latency" value={formatLatency($dashboardData.avg_latency_ms)} icon={Activity} />
			<StatCard title="Critical Alerts" value={$dashboardData.critical_alert_count} icon={Bell} />
		</div>

		<section class="operations-grid">
			<div class="main-column">
				<StatusGrid
					models={$dashboardData.top_models}
					modelsOnline={$dashboardData.models_online}
					modelsTotal={$dashboardData.models_total}
				/>
				<div class="charts-grid">
					<TokenChart data={$dashboardData.token_time_series} />
					<CostChart data={$dashboardData.cost_time_series} />
				</div>
			</div>
			<AlertInbox
				alerts={$dashboardData.active_alerts}
				onAcknowledge={handleAcknowledge}
				onDeclareIncident={handleDeclareIncident}
			/>
		</section>

		<section class="lower-grid">
			<div class="ql-card table-card">
				<div class="card-header-row">
					<h3 class="card-title">Recent Traces</h3>
					<a href="/traces" class="view-all-link">View all</a>
				</div>
				<div class="responsive-table">
					<table class="dashboard-table">
						<thead>
							<tr>
								<th>Name</th>
								<th>Model</th>
								<th>Status</th>
								<th class="text-right">Duration</th>
								<th class="text-right">Tokens</th>
							</tr>
						</thead>
						<tbody>
							{#each $dashboardData.recent_traces as trace}
								<tr onclick={() => window.location.href = `/traces/${trace.id}`} class="clickable-row">
									<td>{trace.name}</td>
									<td class="font-mono text-xs">{trace.model_name || 'N/A'}</td>
									<td>{trace.status === 'error' ? 'Error' : 'Success'}</td>
									<td class="text-right">{trace.total_duration_ms}ms</td>
									<td class="text-right">{trace.total_tokens}</td>
								</tr>
							{:else}
								<tr><td colspan="5">No traces ingested yet.</td></tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
			<AuditTimeline logs={auditLogs} />
		</section>

		<section class="collab-strip">
			<ShieldCheck size={18} />
			<span>Enterprise readiness queue: RBAC review, tenant switcher, comments, approvals, exports, and scheduled reports.</span>
		</section>
	{/if}
</div>

<style>
	.overview-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.overview-toolbar,
	.toolbar-actions,
	.auto-refresh,
	.slo-card,
	.incident-card,
	.collab-strip {
		display: flex;
		align-items: center;
	}

	.overview-toolbar {
		justify-content: space-between;
		gap: 1rem;
	}

	.toolbar-actions {
		gap: 0.625rem;
		flex-wrap: wrap;
		justify-content: flex-end;
	}

	.auto-refresh {
		gap: 0.4rem;
		color: var(--ql-text-muted);
		font-size: 0.8rem;
	}

	.refresh-btn {
		gap: 0.4rem;
	}

	.hero-grid {
		display: grid;
		grid-template-columns: minmax(0, 2fr) minmax(220px, 1fr) minmax(220px, 1fr);
		gap: 1rem;
	}

	.slo-card,
	.incident-card {
		gap: 0.875rem;
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.slo-icon,
	.incident-card > svg {
		width: 44px;
		height: 44px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		background: var(--ql-surface-2);
		color: var(--ql-accent);
		flex-shrink: 0;
	}

	.eyebrow {
		margin: 0;
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		color: var(--ql-text-muted);
	}

	.slo-card strong,
	.incident-card strong {
		display: block;
		font-size: 1.55rem;
		line-height: 1.15;
	}

	.slo-card span,
	.incident-card a {
		color: var(--ql-text-muted);
		font-size: 0.8rem;
	}

	.incident-card a {
		color: var(--ql-accent);
		text-decoration: none;
	}

	.stats-row {
		display: grid;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 1rem;
	}

	.operations-grid {
		display: grid;
		grid-template-columns: minmax(0, 1.5fr) minmax(340px, 0.85fr);
		gap: 1rem;
		align-items: start;
	}

	.main-column {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		min-width: 0;
	}

	.charts-grid,
	.lower-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
	}

	.table-card {
		padding: 1rem;
	}

	.card-header-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.card-title {
		margin: 0;
		font-size: 0.95rem;
	}

	.view-all-link {
		color: var(--ql-accent);
		text-decoration: none;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.responsive-table {
		overflow-x: auto;
	}

	.dashboard-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}

	.dashboard-table th,
	.dashboard-table td {
		padding: 0.625rem;
		border-bottom: 1px solid var(--ql-border);
		text-align: left;
	}

	.dashboard-table th {
		color: var(--ql-text-muted);
		font-size: 0.72rem;
		text-transform: uppercase;
	}

	.clickable-row {
		cursor: pointer;
	}

	.clickable-row:hover {
		background: var(--ql-accent-subtle);
	}

	.text-right {
		text-align: right;
	}

	.collab-strip {
		gap: 0.625rem;
		padding: 0.875rem 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-bg);
		color: var(--ql-text-muted);
		font-size: 0.85rem;
	}

	.loading-container,
	.error-state {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 360px;
		color: var(--ql-text-muted);
	}

	.spinner {
		width: 2rem;
		height: 2rem;
		border: 3px solid rgba(16, 185, 129, 0.1);
		border-top-color: var(--ql-accent);
		border-radius: 9999px;
		animation: spin 0.8s linear infinite;
	}

	@media (max-width: 1180px) {
		.hero-grid,
		.operations-grid,
		.lower-grid {
			grid-template-columns: 1fr;
		}

		.stats-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 640px) {
		.overview-toolbar {
			align-items: flex-start;
			flex-direction: column;
		}

		.toolbar-actions {
			justify-content: flex-start;
		}

		.stats-row {
			grid-template-columns: 1fr;
		}
	}
</style>
