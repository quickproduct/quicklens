<script lang="ts">
	import { onMount } from 'svelte';
	import { getTraces, deleteTrace, type TraceListItem } from '$api/traces';
	import { getModels, type Model } from '$api/models';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import EmptyState from '$components/shared/EmptyState.svelte';
	import { GitBranch, Search, Trash2, Calendar, ShieldAlert, ChevronLeft, ChevronRight } from 'lucide-svelte';

	let traces = $state([] as TraceListItem[]);
	let modelsList = $state([] as Model[]);
	let totalTraces = $state(0);
	let currentPage = $state(1);
	let totalPages = $state(1);
	let limit = 20;

	// Filters
	let searchQuery = $state('');
	let selectedModel = $state('');
	let selectedStatus = $state('');
	let startDate = $state('');
	let endDate = $state('');
	let loading = $state(true);
	let filterDebounceTimer: ReturnType<typeof setTimeout> | null = null;

	async function loadFiltersData() {
		try {
			modelsList = await getModels();
		} catch (err) {
			console.error('Failed to load filter models:', err);
		}
	}

	async function loadTraces(page = 1) {
		loading = true;
		currentPage = page;
		try {
			const res = await getTraces({
				page,
				limit,
				model: selectedModel,
				status: selectedStatus,
				search: searchQuery,
				start_date: startDate ? new Date(startDate).toISOString() : undefined,
				end_date: endDate ? new Date(endDate).toISOString() : undefined
			});
			traces = res.items || [];
			totalTraces = res.total || 0;
			totalPages = res.pages || 1;
		} catch (err) {
			console.error(err);
			addToast('Failed to load traces.', 'error');
		} finally {
			loading = false;
		}
	}

	function handleFilterChange() {
		loadTraces(1);
	}

	function handleSearchInput() {
		if (filterDebounceTimer) clearTimeout(filterDebounceTimer);
		filterDebounceTimer = setTimeout(() => loadTraces(1), 350);
	}

	async function handleDelete(id: string, event: MouseEvent) {
		event.stopPropagation();
		if (!confirm('Are you sure you want to delete this trace and all its spans?')) return;
		try {
			await deleteTrace(id);
			addToast('Trace deleted successfully.', 'success');
			loadTraces(currentPage);
		} catch (err) {
			console.error(err);
			addToast('Failed to delete trace.', 'error');
		}
	}

	onMount(() => {
		loadFiltersData();
		loadTraces(1);
	});

	function formatDateTime(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return isoString;
		}
	}

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.0000';
		return `$${cost.toFixed(4)}`;
	}
</script>

<div class="traces-page">
	<!-- Header -->
	<div class="page-header">
		<h2 class="page-title">Traces</h2>
		<p class="page-subtitle">Track and debug downstream LLM calls, timings, token usage, and spans.</p>
	</div>

	<!-- Filter Panel -->
	<div class="filter-panel ql-card">
		<div class="filter-inputs">
			<div class="filter-group flex-grow">
				<label for="trace-search" class="filter-label">Search</label>
				<div class="search-input-wrapper">
					<Search size={16} />
					<input
						type="text"
						id="trace-search"
						class="search-input"
						placeholder="Search by trace name..."
						bind:value={searchQuery}
						oninput={handleSearchInput}
					/>
				</div>
			</div>

			<div class="filter-group">
				<label for="model-select" class="filter-label">Model</label>
				<select id="model-select" class="ql-input" bind:value={selectedModel} onchange={handleFilterChange}>
					<option value="">All Models</option>
					{#each modelsList as m}
						<option value={m.model_id}>{m.name}</option>
					{/each}
				</select>
			</div>

			<div class="filter-group">
				<label for="status-select" class="filter-label">Status</label>
				<select id="status-select" class="ql-input" bind:value={selectedStatus} onchange={handleFilterChange}>
					<option value="">All Statuses</option>
					<option value="ok">Success</option>
					<option value="error">Error</option>
				</select>
			</div>

			<div class="filter-group">
				<label for="date-start" class="filter-label">From</label>
				<input
					type="date"
					id="date-start"
					class="ql-input"
					bind:value={startDate}
					onchange={handleFilterChange}
				/>
			</div>

			<div class="filter-group">
				<label for="date-end" class="filter-label">To</label>
				<input
					type="date"
					id="date-end"
					class="ql-input"
					bind:value={endDate}
					onchange={handleFilterChange}
				/>
			</div>
		</div>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading traces...</span>
		</div>
	{:else if traces.length === 0}
		<EmptyState
			title="No Traces Found"
			description="No traces have been captured matching the active filters. Ensure your SDK is integrated or proxy is active."
			icon={GitBranch}
		/>
	{:else}
		<div class="traces-list-card ql-card">
			<div class="table-wrapper">
				<table class="traces-table">
					<thead>
						<tr>
							<th>Trace Name</th>
							<th>Status</th>
							<th>Associated Model</th>
							<th>Spans</th>
							<th class="text-right">Duration</th>
							<th class="text-right">Tokens</th>
							<th class="text-right">Cost</th>
							<th>Ingested At</th>
							<th class="text-center">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each traces as trace (trace.id)}
							<tr onclick={() => window.location.href = `/traces/${trace.id}`} class="clickable-row">
								<td class="trace-name font-semibold" title={trace.name}>
									{trace.name}
								</td>
								<td>
									<Badge
										text={trace.status === 'error' ? 'Error' : 'Success'}
										variant={trace.status === 'error' ? 'danger' : 'success'}
									/>
								</td>
								<td class="font-mono text-xs">
									{trace.model || 'N/A'}
								</td>
								<td class="span-count-cell">
									{trace.span_count || 0} spans
								</td>
								<td class="text-right font-medium">{trace.duration_ms}ms</td>
								<td class="text-right">{trace.total_tokens?.toLocaleString() || 0}</td>
								<td class="text-right text-accent font-semibold">{formatCost(trace.cost)}</td>
								<td class="text-muted text-xs">{formatDateTime(trace.created_at)}</td>
								<td class="text-center" onclick={(e) => e.stopPropagation()}>
									<button class="delete-btn" onclick={(e) => handleDelete(trace.id, e)} title="Delete Trace">
										<Trash2 size={15} />
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="pagination-controls">
					<span class="pagination-info">
						Showing {(currentPage - 1) * limit + 1} - {Math.min(currentPage * limit, totalTraces)} of {totalTraces}
					</span>
					<div class="pagination-buttons">
						<button
							class="ql-btn-secondary pag-btn"
							disabled={currentPage === 1}
							onclick={() => loadTraces(currentPage - 1)}
						>
							<ChevronLeft size={16} /> Prev
						</button>
						<span class="page-number">Page {currentPage} of {totalPages}</span>
						<button
							class="ql-btn-secondary pag-btn"
							disabled={currentPage === totalPages}
							onclick={() => loadTraces(currentPage + 1)}
						>
							Next <ChevronRight size={16} />
						</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.traces-page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.page-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--ql-text);
		letter-spacing: -0.025em;
	}

	.page-subtitle {
		font-size: 0.875rem;
		color: var(--ql-text-muted);
	}

	.filter-panel {
		padding: 1rem;
	}

	.filter-inputs {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.flex-grow {
		flex-grow: 1;
	}

	.filter-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.search-input-wrapper {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background-color: var(--ql-surface-2);
		padding: 0.45rem 0.75rem;
		border-radius: 0.375rem;
		border: 1px solid var(--ql-border);
		color: var(--ql-text-muted);
	}

	.search-input {
		border: none;
		background: none;
		color: var(--ql-text);
		font-size: 0.875rem;
		width: 100%;
	}

	.search-input:focus {
		outline: none;
	}

	.filter-group select,
	.filter-group input[type='date'] {
		padding: 0.45rem 0.5rem;
		font-size: 0.85rem;
		border-color: var(--ql-border);
		min-width: 130px;
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

	.traces-list-card {
		padding: 0;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.traces-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.traces-table th,
	.traces-table td {
		padding: 0.75rem 1rem;
		text-align: left;
		border-bottom: 1px solid var(--ql-border);
	}

	.traces-table th {
		color: var(--ql-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		font-size: 0.7rem;
		letter-spacing: 0.05em;
		background-color: rgba(26, 29, 46, 0.4);
	}

	.traces-table tr:last-child td {
		border-bottom: none;
	}

	.clickable-row {
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.clickable-row:hover {
		background-color: rgba(255, 255, 255, 0.02);
	}

	.trace-name {
		max-width: 200px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.span-count-cell {
		color: var(--ql-text-muted);
	}

	.delete-btn {
		background: none;
		border: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.25rem;
		border-radius: 0.25rem;
		transition: all 0.2s;
		margin: 0 auto;
	}

	.delete-btn:hover {
		color: var(--ql-danger, #ef4444);
		background-color: var(--ql-surface-2);
	}

	.text-right {
		text-align: right;
	}

	.text-center {
		text-align: center;
	}

	.text-muted {
		color: var(--ql-text-muted);
	}

	.text-accent {
		color: var(--ql-accent);
	}

	.font-semibold {
		font-weight: 500;
	}

	.font-medium {
		font-weight: 500;
	}

	/* Pagination styles */
	.pagination-controls {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border-top: 1px solid var(--ql-border);
		flex-wrap: wrap;
		gap: 1rem;
	}

	.pagination-info {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
	}

	.pagination-buttons {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.pag-btn {
		font-size: 0.75rem;
		padding: 0.35rem 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.page-number {
		font-size: 0.825rem;
		color: var(--ql-text);
		font-weight: 500;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
