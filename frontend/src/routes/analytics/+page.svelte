<script lang="ts">
	import { onMount } from 'svelte';
	import { getDashboard, type DashboardResponse } from '$api/dashboard';
	import { getSLODefinitions, type SLODefinition } from '$api/operations';
	import CostChart from '$components/dashboard/CostChart.svelte';
	import TokenChart from '$components/dashboard/TokenChart.svelte';
	import Badge from '$components/shared/Badge.svelte';
	import { addToast } from '$stores/ui';
	import { Activity, BarChart3, Target, TrendingUp } from 'lucide-svelte';

	let dashboard = $state(null as DashboardResponse | null);
	let slos = $state([] as SLODefinition[]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const [dashboardRes, sloRes] = await Promise.all([getDashboard(), getSLODefinitions()]);
			dashboard = dashboardRes;
			slos = sloRes;
		} catch (err) {
			console.error(err);
			addToast('Failed to load analytics.', 'error');
		} finally {
			loading = false;
		}
	});
</script>

<div class="analytics-page">
	<div class="page-header">
		<div>
			<h2>Analytics & Trends</h2>
			<p>Decision views for throughput, cost, latency, SLO burn, and health forecasting.</p>
		</div>
	</div>

	{#if loading}
		<div class="empty-state">Loading analytics...</div>
	{:else if dashboard}
		<div class="analytics-grid">
			<section class="metric-card">
				<Target size={20} />
				<div>
					<span>SLO Attainment</span>
					<strong>{dashboard.slo_burn_summary.attainment_percent.toFixed(2)}%</strong>
					<p>{dashboard.slo_burn_summary.error_budget_remaining.toFixed(0)}% error budget remaining</p>
				</div>
			</section>
			<section class="metric-card">
				<Activity size={20} />
				<div>
					<span>Health Score</span>
					<strong>{dashboard.health_score}</strong>
					<p>{dashboard.critical_alert_count} critical alerts affecting score</p>
				</div>
			</section>
			<section class="metric-card">
				<TrendingUp size={20} />
				<div>
					<span>Burn Rate</span>
					<strong>{dashboard.slo_burn_summary.burn_rate.toFixed(0)}%</strong>
					<p>Derived from current trace success window</p>
				</div>
			</section>
		</div>

		<div class="charts-grid">
			<TokenChart data={dashboard.token_time_series} />
			<CostChart data={dashboard.cost_time_series} />
		</div>

		<section class="slo-card">
			<div class="section-header">
				<h3><BarChart3 size={17} /> Service Levels</h3>
			</div>
			<div class="responsive-table">
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Target</th>
							<th>Window</th>
							<th>Events</th>
							<th>Budget</th>
							<th>Status</th>
						</tr>
					</thead>
					<tbody>
						{#each slos as slo}
							<tr>
								<td>{slo.name}</td>
								<td>{slo.target_percent.toFixed(1)}%</td>
								<td>{slo.period_days} days</td>
								<td>{slo.good_events.toLocaleString()} / {slo.total_events.toLocaleString()}</td>
								<td>{slo.error_budget_remaining.toFixed(0)}%</td>
								<td>
									<Badge
										text={slo.status}
										variant={slo.status === 'healthy' ? 'success' : slo.status === 'warning' ? 'warning' : 'danger'}
									/>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>
	{/if}
</div>

<style>
	.analytics-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	h2,
	h3,
	p {
		margin: 0;
	}

	.page-header p,
	.metric-card span,
	.metric-card p,
	.empty-state {
		color: var(--ql-text-muted);
		font-size: 0.875rem;
	}

	.analytics-grid,
	.charts-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.charts-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.metric-card,
	.slo-card {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.metric-card {
		display: flex;
		align-items: flex-start;
		gap: 0.875rem;
	}

	.metric-card svg {
		color: var(--ql-accent);
	}

	.metric-card div {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.metric-card strong {
		font-size: 1.6rem;
	}

	.section-header h3 {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.875rem;
	}

	.responsive-table {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	th,
	td {
		padding: 0.7rem;
		border-bottom: 1px solid var(--ql-border);
		text-align: left;
	}

	th {
		color: var(--ql-text-muted);
		font-size: 0.72rem;
		text-transform: uppercase;
	}

	.empty-state {
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}

	@media (max-width: 980px) {
		.analytics-grid,
		.charts-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
