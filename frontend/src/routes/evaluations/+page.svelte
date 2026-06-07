<script lang="ts">
	import { onMount } from 'svelte';
	import { getEvaluations, type Evaluation } from '$api/evaluations';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import EmptyState from '$components/shared/EmptyState.svelte';
	import { Star, ThumbsUp, ThumbsDown, MessageSquare, Calendar, Activity as ActivityIcon } from 'lucide-svelte';

	let evaluations = $state([] as Evaluation[]);
	let loading = $state(true);

	// Locally aggregated summary statistics
	let totalEvals = $derived(evaluations.length);
	let avgScore = $derived(
		totalEvals > 0 ? evaluations.reduce((sum, e) => sum + e.score_value, 0) / totalEvals : 0
	);
	let thumbsUpCount = $derived(evaluations.filter((e) => e.score_value === 1).length);
	let thumbsDownCount = $derived(evaluations.filter((e) => e.score_value === 0).length);

	async function loadEvaluations() {
		try {
			evaluations = await getEvaluations();
		} catch (err) {
			console.error(err);
			addToast('Failed to load evaluations list.', 'error');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadEvaluations();
	});

	function formatDateTime(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} catch {
			return isoString;
		}
	}
</script>

<div class="evaluations-page">
	<!-- Header -->
	<div class="page-header">
		<h2 class="page-title">Evaluations & Feedback</h2>
		<p class="page-subtitle">Inspect manual feedback, system evaluations, and quality scores collected on LLM traces.</p>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading evaluations...</span>
		</div>
	{:else if evaluations.length === 0}
		<EmptyState
			title="No Evaluations Yet"
			description="Submit likes/dislikes on trace spans to populate evaluations."
			icon={Star}
		/>
	{:else}
		<!-- Stats Grid -->
		<div class="stats-row">
			<div class="ql-card stat-card">
				<Star size={24} class="stat-icon info-icon" />
				<div class="stat-info">
					<span class="stat-title">Total Evaluations</span>
					<span class="stat-value">{totalEvals}</span>
				</div>
			</div>

			<div class="ql-card stat-card">
				<ThumbsUp size={24} class="stat-icon success-icon" />
				<div class="stat-info">
					<span class="stat-title">Likes (Thumbs Up)</span>
					<span class="stat-value">{thumbsUpCount}</span>
				</div>
			</div>

			<div class="ql-card stat-card">
				<ThumbsDown size={24} class="stat-icon danger-icon" />
				<div class="stat-info">
					<span class="stat-title">Dislikes (Thumbs Down)</span>
					<span class="stat-value">{thumbsDownCount}</span>
				</div>
			</div>

			<div class="ql-card stat-card">
				<ActivityIcon size={24} class="stat-icon warning-icon" />
				<div class="stat-info">
					<span class="stat-title">Average Rating</span>
					<span class="stat-value">{avgScore.toFixed(2)}</span>
				</div>
			</div>
		</div>

		<!-- Table Card -->
		<div class="ql-card table-card">
			<h3 class="card-title">Recent Feedback Entries</h3>
			<div class="table-wrapper">
				<table class="evals-table">
					<thead>
						<tr>
							<th>Span ID</th>
							<th>Metric / Type</th>
							<th>Value</th>
							<th>Feedback Comment</th>
							<th>Evaluator</th>
							<th>Submitted At</th>
						</tr>
					</thead>
					<tbody>
						{#each evaluations as ev (ev.id)}
							<tr>
								<td class="font-mono text-xs text-muted" title={ev.span_id}>
									{ev.span_id.substring(0, 8)}...
								</td>
								<td class="font-semibold text-uppercase">{ev.score_type}</td>
								<td>
									<span class="score-badge" class:like={ev.score_value === 1} class:dislike={ev.score_value === 0}>
										{ev.score_value === 1 ? '👍 Positive' : ev.score_value === 0 ? '👎 Negative' : ev.score_value}
									</span>
								</td>
								<td class="feedback-comment">
									{ev.feedback_text || '-'}
								</td>
								<td>
									<Badge text={ev.evaluator} variant="neutral" />
								</td>
								<td class="text-xs text-muted">
									{formatDateTime(ev.created_at)}
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
	.evaluations-page {
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

	.stats-row {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1rem;
	}

	@media (max-width: 768px) {
		.stats-row {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.stats-row {
			grid-template-columns: 1fr;
		}
	}

	.stat-card {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.25rem;
	}

	.stat-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 3rem;
		height: 3rem;
		border-radius: 0.5rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		flex-shrink: 0;
	}

	.info-icon { color: var(--ql-info, #3b82f6); }
	.success-icon { color: var(--ql-success, #22c55e); }
	.danger-icon { color: var(--ql-danger, #ef4444); }
	.warning-icon { color: var(--ql-warning, #f59e0b); }

	.stat-info {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.stat-title {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		font-weight: 600;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--ql-text);
	}

	.table-card {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.evals-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.evals-table th,
	.evals-table td {
		padding: 0.75rem 1rem;
		text-align: left;
		border-bottom: 1px solid var(--ql-border);
	}

	.evals-table th {
		color: var(--ql-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		font-size: 0.7rem;
		letter-spacing: 0.05em;
		background-color: rgba(26, 29, 46, 0.4);
	}

	.evals-table tr:last-child td {
		border-bottom: none;
	}

	.text-uppercase {
		text-transform: uppercase;
		font-size: 0.75rem;
	}

	.text-muted {
		color: var(--ql-text-muted);
	}

	.text-xs {
		font-size: 0.75rem;
	}

	.feedback-comment {
		max-width: 250px;
		word-break: break-word;
	}

	.score-badge {
		font-size: 0.75rem;
		font-weight: 600;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		background-color: var(--ql-surface-2);
		color: var(--ql-text);
	}

	.score-badge.like {
		background-color: rgba(34, 197, 94, 0.1);
		color: #4ade80;
	}

	.score-badge.dislike {
		background-color: rgba(239, 68, 68, 0.1);
		color: #fca5a5;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
