<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import CostChart from '$components/dashboard/CostChart.svelte';
	import StatCard from '$components/dashboard/StatCard.svelte';
	import StatusGrid from '$components/dashboard/StatusGrid.svelte';
	import TokenChart from '$components/dashboard/TokenChart.svelte';
	import FreshnessBadge from '$components/shared/FreshnessBadge.svelte';
	import { dashboardAutoRefresh, dashboardData, dashboardError, dashboardFreshnessLabel, dashboardLoading, refreshDashboard, scheduleDashboardRefresh } from '$stores/dashboard';
	import { wsConnected } from '$stores/ui';
	import { createWebSocket } from '$lib/websocket/client';
	import { Activity, Coins, Cpu, GitBranch, RefreshCw, ShieldCheck } from 'lucide-svelte';

	let wsClient: ReturnType<typeof createWebSocket> | null = null;

	onMount(() => {
		refreshDashboard();

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

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.0000';
		return `$${cost.toFixed(4)}`;
	}

	function formatLatency(ms: number): string {
		if (ms === 0 || !ms) return '0ms';
		return `${ms.toFixed(0)}ms`;
	}

	function formatPercent(val: number): string {
		if (val === undefined || isNaN(val)) return '100.0%';
		return `${val.toFixed(1)}%`;
	}
</script>

<div class="flex flex-col gap-5 animate-fade-in">
	<!-- Top Bar -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h2 class="page-title text-gradient">LLM Observability</h2>
			<p class="text-sm text-ql-text-muted">Real-time metrics, logs, and token tracing for your language model workflows.</p>
		</div>
		<div class="flex flex-wrap items-center justify-start gap-2.5 sm:justify-end">
			<label class="flex items-center gap-1.5 text-xs text-ql-text-muted cursor-pointer select-none">
				<input type="checkbox" bind:checked={$dashboardAutoRefresh} class="rounded border-ql-border bg-ql-surface text-ql-accent focus:ring-0" />
				Auto-refresh
			</label>
			<FreshnessBadge label={$dashboardFreshnessLabel} connected={$wsConnected} />
			<button class="ql-btn-secondary flex items-center gap-1.5" type="button" onclick={() => refreshDashboard()} disabled={$dashboardLoading}>
				<span class="inline-flex" class:animate-spin={$dashboardLoading}>
					<RefreshCw size={14} />
				</span>
				Refresh
			</button>
		</div>
	</div>

	{#if $dashboardLoading && !$dashboardData}
		<div class="flex min-h-[400px] flex-col items-center justify-center text-ql-text-muted gap-3">
			<div class="h-8 w-8 animate-spin rounded-full border-2 border-ql-accent border-t-transparent"></div>
			<span>Loading system dashboard...</span>
		</div>
	{:else if $dashboardError}
		<div class="flex min-h-[400px] items-center justify-center text-ql-danger">
			{$dashboardError}
		</div>
	{:else if $dashboardData}
		<!-- Metrics Overview -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5 stagger-children">
			<StatCard title="Total Runs Today" value={$dashboardData.total_traces_today} icon={GitBranch} />
			<StatCard title="Tokens Processed" value={$dashboardData.total_tokens_today.toLocaleString()} icon={Cpu} />
			<StatCard title="Total Spend Today" value={formatCost($dashboardData.total_cost_today)} icon={Coins} />
			<StatCard title="Success Rate" value={formatPercent($dashboardData.success_rate_today)} icon={ShieldCheck} />
			<StatCard title="Avg Latency" value={formatLatency($dashboardData.avg_latency_ms)} icon={Activity} />
		</div>

		<!-- Charts & Inventory -->
		<div class="grid grid-cols-1 items-start gap-5 lg:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
			<!-- Throughput Analytics -->
			<div class="grid grid-cols-1 gap-5">
				<div class="grid grid-cols-1 gap-5 md:grid-cols-2">
					<TokenChart data={$dashboardData.token_time_series} />
					<CostChart data={$dashboardData.cost_time_series} />
				</div>
			</div>

			<!-- Models Status Grid -->
			<StatusGrid
				models={$dashboardData.top_models}
				modelsOnline={$dashboardData.models_online}
				modelsTotal={$dashboardData.models_total}
			/>
		</div>

		<!-- Recent Activity table -->
		<div class="ql-card p-5">
			<div class="mb-4 flex items-center justify-between">
				<div>
					<h3 class="m-0 text-md font-semibold text-ql-text">Recent Requests</h3>
					<p class="text-xs text-ql-text-muted m-0 mt-0.5">Click any request row to inspect full trace inputs, outputs, and JSON payloads.</p>
				</div>
				<a href="/traces" class="ql-btn-ghost ql-btn-sm text-xs font-bold no-underline">View all logs</a>
			</div>
			<div class="overflow-x-auto ql-scrollbar">
				<table class="w-full border-collapse text-sm">
					<thead>
						<tr>
							<th class="border-b border-ql-border pb-3 text-left text-xs font-semibold uppercase text-ql-text-muted">Request Name</th>
							<th class="border-b border-ql-border pb-3 text-left text-xs font-semibold uppercase text-ql-text-muted">Model ID</th>
							<th class="border-b border-ql-border pb-3 text-left text-xs font-semibold uppercase text-ql-text-muted">Provider</th>
							<th class="border-b border-ql-border pb-3 text-left text-xs font-semibold uppercase text-ql-text-muted">Timing</th>
							<th class="border-b border-ql-border pb-3 text-right text-xs font-semibold uppercase text-ql-text-muted">Tokens</th>
							<th class="border-b border-ql-border pb-3 text-right text-xs font-semibold uppercase text-ql-text-muted">Cost</th>
							<th class="border-b border-ql-border pb-3 text-center text-xs font-semibold uppercase text-ql-text-muted">Status</th>
						</tr>
					</thead>
					<tbody>
						{#each $dashboardData.recent_traces as trace}
							<tr onclick={() => goto(`/traces/${trace.id}`)} class="cursor-pointer hover:bg-ql-accent-subtle group transition-colors">
								<td class="border-b border-ql-border py-3 text-left font-medium text-ql-text group-hover:text-ql-accent">{trace.name}</td>
								<td class="border-b border-ql-border py-3 text-left font-mono text-xs">{trace.model_name || 'N/A'}</td>
								<td class="border-b border-ql-border py-3 text-left text-xs text-ql-text-muted capitalize">{trace.provider || 'unknown'}</td>
								<td class="border-b border-ql-border py-3 text-left text-xs text-ql-text-muted">{trace.total_duration_ms}ms</td>
								<td class="border-b border-ql-border py-3 text-right font-mono text-xs">{trace.total_tokens}</td>
								<td class="border-b border-ql-border py-3 text-right font-mono text-xs text-ql-accent">{formatCost(trace.total_cost)}</td>
								<td class="border-b border-ql-border py-3 text-center">
									<span class="ql-badge" class:ql-badge-success={trace.status !== 'error'} class:ql-badge-danger={trace.status === 'error'}>
										{trace.status === 'error' ? 'Error' : 'Success'}
									</span>
								</td>
							</tr>
						{:else}
							<tr>
								<td colspan="7" class="py-10 text-center text-ql-text-muted">
									No traces ingested yet. Visit the integration settings to connect your LLM.
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<style>
	.page-title {
		font-size: 1.5rem;
		font-weight: 800;
		letter-spacing: -0.025em;
		margin: 0;
	}
</style>
