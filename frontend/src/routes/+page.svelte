<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getDashboard, type DashboardResponse } from '$api/dashboard';
	import StatCard from '$components/dashboard/StatCard.svelte';
	import TokenChart from '$components/dashboard/TokenChart.svelte';
	import CostChart from '$components/dashboard/CostChart.svelte';
	import Badge from '$components/shared/Badge.svelte';
	import { createWebSocket } from '$lib/websocket/client';
	import { GitBranch, Cpu, Coins, Clock, Activity, ArrowRight, Server } from 'lucide-svelte';

	let data = $state(null as DashboardResponse | null);
	let loading = $state(true);
	let wsClient: any = null;

	async function loadDashboard() {
		try {
			data = await getDashboard();
		} catch (err) {
			console.error('Failed to load dashboard data:', err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDashboard();

		// Subscribe to traces WebSocket channel for live updates
		wsClient = createWebSocket('traces', (msg: any) => {
			// Reload dashboard data on new trace ingestion
			if (msg && (msg.type === 'trace_ingested' || msg.type === 'update')) {
				loadDashboard();
			}
		});
	});

	onDestroy(() => {
		if (wsClient) {
			wsClient.close();
		}
	});

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.00';
		return `$${cost.toFixed(4)}`;
	}

	function formatLatency(ms: number): string {
		if (ms === 0 || !ms) return '0ms';
		return `${ms.toFixed(0)}ms`;
	}
</script>

<div class="dashboard-page">
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading Dashboard...</span>
		</div>
	{:else if data}
		<!-- Stats Row -->
		<div class="stats-row">
			<StatCard
				title="Traces Today"
				value={data.total_traces_today}
				icon={GitBranch}
			/>
			<StatCard
				title="Tokens Processed"
				value={data.total_tokens_today.toLocaleString()}
				icon={Cpu}
			/>
			<StatCard
				title="Cost Today"
				value={formatCost(data.total_cost_today)}
				icon={Coins}
			/>
			<StatCard
				title="Avg Latency"
				value={formatLatency(data.avg_latency_ms)}
				icon={Clock}
			/>
			<StatCard
				title="Active Models"
				value="{data.models_online} / {data.models_total}"
				icon={Activity}
			/>
		</div>

		<!-- Charts Grid -->
		<div class="charts-grid">
			<div class="chart-col token-chart-col">
				<TokenChart data={data.token_time_series} />
			</div>
			<div class="chart-col cost-chart-col">
				<CostChart data={data.cost_time_series} />
			</div>
		</div>

		<!-- Tables Grid -->
		<div class="tables-grid">
			<!-- Top Models -->
			<div class="ql-card table-card top-models-card">
				<h3 class="card-title">Top Models</h3>
				{#if data.top_models.length === 0}
					<div class="empty-table-state">No model usage recorded today.</div>
				{:else}
					<div class="table-wrapper">
						<table class="dashboard-table">
							<thead>
								<tr>
									<th>Model</th>
									<th>Provider</th>
									<th class="text-right">Requests</th>
									<th class="text-right">Tokens</th>
								</tr>
							</thead>
							<tbody>
								{#each data.top_models as model}
									<tr>
										<td class="font-semibold">{model.model_name}</td>
										<td>
											<span class="provider-pill">{model.provider}</span>
										</td>
										<td class="text-right">{model.request_count.toLocaleString()}</td>
										<td class="text-right">{model.token_count.toLocaleString()}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>

			<!-- Recent Traces -->
			<div class="ql-card table-card recent-traces-card">
				<div class="card-header-row">
					<h3 class="card-title">Recent Traces</h3>
					<a href="/traces" class="view-all-link">
						View All <ArrowRight size={14} />
					</a>
				</div>
				{#if data.recent_traces.length === 0}
					<div class="empty-table-state">No traces ingested yet.</div>
				{:else}
					<div class="table-wrapper">
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
								{#each data.recent_traces as trace}
									<tr onclick={() => window.location.href = `/traces/${trace.id}`} class="clickable-row">
										<td class="trace-name-cell font-semibold">{trace.name}</td>
										<td class="font-mono text-xs">{trace.model_name || 'N/A'}</td>
										<td>
											<Badge
												text={trace.status === 'error' ? 'Error' : 'Success'}
												variant={trace.status === 'error' ? 'danger' : 'success'}
											/>
										</td>
										<td class="text-right">{trace.total_duration_ms}ms</td>
										<td class="text-right">{trace.total_tokens}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.dashboard-page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: 1rem;
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

	.stats-row {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 1rem;
	}

	@media (max-width: 1024px) {
		.stats-row {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	@media (max-width: 640px) {
		.stats-row {
			grid-template-columns: repeat(1, 1fr);
		}
	}

	.charts-grid {
		display: grid;
		grid-template-columns: 3.5fr 2.5fr;
		gap: 1.5rem;
	}

	@media (max-width: 768px) {
		.charts-grid {
			grid-template-columns: 1fr;
		}
	}

	.chart-col {
		min-width: 0;
	}

	.tables-grid {
		display: grid;
		grid-template-columns: 2fr 3fr;
		gap: 1.5rem;
	}

	@media (max-width: 768px) {
		.tables-grid {
			grid-template-columns: 1fr;
		}
	}

	.table-card {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		min-width: 0;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.card-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.view-all-link {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.825rem;
		color: var(--ql-accent);
		text-decoration: none;
		font-weight: 500;
		transition: color 0.2s;
	}

	.view-all-link:hover {
		color: var(--ql-accent-hover);
	}

	.empty-table-state {
		text-align: center;
		padding: 3rem 1.5rem;
		color: var(--ql-text-muted);
		font-size: 0.875rem;
		border: 1px dashed var(--ql-border);
		border-radius: 0.375rem;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.dashboard-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.dashboard-table th,
	.dashboard-table td {
		padding: 0.625rem 0.75rem;
		text-align: left;
		border-bottom: 1px solid var(--ql-border);
	}

	.dashboard-table th {
		color: var(--ql-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		font-size: 0.7rem;
		letter-spacing: 0.05em;
	}

	.dashboard-table tr:last-child td {
		border-bottom: none;
	}

	.clickable-row {
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.clickable-row:hover {
		background-color: rgba(255, 255, 255, 0.02);
	}

	.trace-name-cell {
		max-width: 150px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.provider-pill {
		font-size: 0.65rem;
		font-weight: 700;
		text-transform: uppercase;
		padding: 0.1rem 0.35rem;
		border-radius: 0.25rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		color: var(--ql-text);
	}

	.text-right {
		text-align: right;
	}

	.font-semibold {
		font-weight: 500;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
