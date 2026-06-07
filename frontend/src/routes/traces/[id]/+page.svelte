<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { getTrace, type TraceResponse } from '$api/traces';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import TraceTree from '$components/traces/TraceTree.svelte';
	import TraceTimeline from '$components/traces/TraceTimeline.svelte';
	import SpanDetail from '$components/traces/SpanDetail.svelte';
	import { ArrowLeft, GitBranch, Calendar, Clock, Coins, Cpu } from 'lucide-svelte';

	let traceId = $state('');
	let trace = $state(null as TraceResponse | null);
	let selectedSpanId = $state('');
	let loading = $state(true);

	$effect(() => {
		const unsub = page.subscribe((p) => {
			if (p.params.id) {
				traceId = p.params.id;
			}
		});
		return unsub;
	});

	$effect(() => {
		if (traceId) {
			loadTrace();
		}
	});

	async function loadTrace() {
		loading = true;
		try {
			trace = await getTrace(traceId);
			// Auto-select the first span if available
			if (trace && trace.spans && trace.spans.length > 0) {
				selectedSpanId = trace.spans[0].id;
			}
		} catch (err) {
			console.error(err);
			addToast('Failed to load trace details.', 'error');
		} finally {
			loading = false;
		}
	}

	function findSpanById(spans: any[], id: string): any | null {
		for (const s of spans) {
			if (s.id === id) return s;
			if (s.children && s.children.length > 0) {
				const found = findSpanById(s.children, id);
				if (found) return found;
			}
		}
		return null;
	}

	let selectedSpan = $derived(trace ? findSpanById(trace.spans, selectedSpanId) : null);

	function formatDateTime(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' });
		} catch {
			return isoString;
		}
	}

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.0000';
		return `$${cost.toFixed(4)}`;
	}
</script>

<div class="trace-detail-page">
	<!-- Back link & Header -->
	<div class="trace-header-row">
		<div class="header-left">
			<a href="/traces" class="back-link">
				<ArrowLeft size={16} /> Back to Traces
			</a>
			{#if trace}
				<div class="title-section">
					<h2 class="trace-title">{trace.name}</h2>
					<Badge
						text={trace.status === 'error' ? 'Error' : 'Success'}
						variant={trace.status === 'error' ? 'danger' : 'success'}
					/>
				</div>
			{/if}
		</div>

		{#if trace}
			<div class="trace-summary-metrics font-mono text-xs">
				<div class="metric-summary-box">
					<Clock size={12} />
					<span>{trace.duration_ms}ms</span>
				</div>
				<div class="metric-summary-box">
					<Cpu size={12} />
					<span>{trace.total_tokens} tokens</span>
				</div>
				<div class="metric-summary-box">
					<Coins size={12} />
					<span>{formatCost(trace.cost)}</span>
				</div>
				<div class="metric-summary-box">
					<Calendar size={12} />
					<span>{formatDateTime(trace.created_at)}</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Content split panels -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading trace details...</span>
		</div>
	{:else if trace}
		<!-- Trace session alert info -->
		{#if trace.session_id}
			<div class="session-info-bar">
				<GitBranch size={16} />
				<span>Part of trace session: <span class="session-id-val font-mono">{trace.session_id}</span></span>
			</div>
		{/if}

		<div class="trace-panels-layout">
			<!-- Left panel: Tree & Timeline -->
			<div class="left-panel">
				<div class="tree-section">
					<h3 class="panel-section-title">Span Tree</h3>
					<TraceTree
						spans={trace.spans}
						{selectedSpanId}
						onSelect={(id) => (selectedSpanId = id)}
					/>
				</div>

				<div class="timeline-section">
					<TraceTimeline spans={trace.spans} totalDurationMs={trace.duration_ms} />
				</div>
			</div>

			<!-- Right panel: Span Details -->
			<div class="right-panel ql-card">
				<SpanDetail span={selectedSpan} />
			</div>
		</div>
	{/if}
</div>

<style>
	.trace-detail-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		height: 100%;
		min-height: 0;
	}

	.trace-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
		border-bottom: 1px solid var(--ql-border);
		padding-bottom: 1rem;
		flex-shrink: 0;
	}

	.header-left {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		min-width: 0;
	}

	.back-link {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.825rem;
		color: var(--ql-text-muted);
		text-decoration: none;
		transition: color 0.2s;
	}

	.back-link:hover {
		color: var(--ql-text);
	}

	.title-section {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		min-width: 0;
	}

	.trace-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--ql-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		letter-spacing: -0.025em;
	}

	.trace-summary-metrics {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.metric-summary-box {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.35rem 0.5rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		border-radius: 0.25rem;
		color: var(--ql-text-muted);
	}

	.session-info-bar {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 1rem;
		background-color: rgba(16, 185, 129, 0.04);
		border: 1px solid rgba(16, 185, 129, 0.15);
		border-radius: 0.375rem;
		color: var(--ql-accent);
		font-size: 0.825rem;
		flex-shrink: 0;
	}

	.session-id-val {
		font-weight: 600;
		text-decoration: underline;
	}

	.trace-panels-layout {
		display: grid;
		grid-template-columns: 3.5fr 2.5fr;
		gap: 1.25rem;
		flex-grow: 1;
		min-height: 0;
	}

	@media (max-width: 1024px) {
		.trace-panels-layout {
			grid-template-columns: 1fr;
		}
	}

	.left-panel {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		min-height: 0;
		overflow-y: auto;
	}

	.tree-section {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.panel-section-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.right-panel {
		padding: 1.25rem;
		overflow-y: auto;
		height: 100%;
		border-color: var(--ql-border);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
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

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
