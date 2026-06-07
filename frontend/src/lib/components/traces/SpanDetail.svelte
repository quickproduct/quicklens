<script lang="ts">
	import Badge from '../shared/Badge.svelte';
	import { getSpanEvaluations, createEvaluation } from '$api/evaluations';
	import { addToast } from '$stores/ui';
	import { Cpu, Clock, Coins, Database, Plus, ThumbsUp, ThumbsDown } from 'lucide-svelte';

	let { span }: { span: any } = $props();

	let evaluations = $state([] as any[]);
	let activeTab = $state('io'); // 'io', 'metadata', 'evals'
	let loadingEvals = $state(false);

	// Feedback Form State
	let scoreType = $state('sentiment');
	let scoreValue = $state(1); // 1 = Positive/Thumbs Up, 0 = Negative/Thumbs Down, or a float
	let feedbackText = $state('');
	let evaluator = $state('user');
	let submittingEval = $state(false);

	$effect(() => {
		if (span && span.id) {
			loadEvaluations();
		}
	});

	async function loadEvaluations() {
		loadingEvals = true;
		try {
			evaluations = await getSpanEvaluations(span.id);
		} catch (err) {
			console.error('Failed to load evaluations:', err);
		} finally {
			loadingEvals = false;
		}
	}

	async function submitFeedback(value: number) {
		submittingEval = true;
		try {
			await createEvaluation({
				span_id: span.id,
				score_type: 'thumbs',
				score_value: value,
				evaluator: 'quicklens-ui',
				feedback_text: feedbackText
			});
			addToast('Feedback submitted successfully!', 'success');
			feedbackText = '';
			loadEvaluations();
		} catch (err) {
			console.error('Failed to submit feedback:', err);
			addToast('Failed to submit feedback.', 'error');
		} finally {
			submittingEval = false;
		}
	}

	function tryFormatJson(str: string): string {
		if (!str) return '';
		try {
			const parsed = JSON.parse(str);
			return JSON.stringify(parsed, null, 2);
		} catch {
			return str;
		}
	}

	function formatCost(cost: number): string {
		if (cost === 0 || !cost) return '$0.0000';
		return `$${cost.toFixed(4)}`;
	}
</script>

{#if span}
	<div class="span-detail-panel">
		<!-- Header -->
		<div class="panel-header">
			<div class="panel-header-title">
				<h3 class="span-title">{span.name}</h3>
				<div class="badges-row">
					<Badge text={span.type} variant="info" />
					<Badge
						text={span.status === 'error' ? 'Error' : 'Success'}
						variant={span.status === 'error' ? 'danger' : 'success'}
					/>
				</div>
			</div>
			{#if span.model_id}
				<div class="span-model-details">
					<span class="model-label">{span.provider} / {span.model_id}</span>
				</div>
			{/if}
		</div>

		<!-- Metrics Grid -->
		<div class="metrics-grid">
			<div class="metric-card">
				<Clock size={16} class="metric-icon" />
				<div class="metric-info">
					<span class="metric-label">Duration</span>
					<span class="metric-value">{span.duration_ms}ms</span>
				</div>
			</div>
			<div class="metric-card">
				<Cpu size={16} class="metric-icon" />
				<div class="metric-info">
					<span class="metric-label">Tokens</span>
					<span class="metric-value">{span.total_tokens || 0}</span>
					{#if span.prompt_tokens > 0}
						<span class="metric-sub">{span.prompt_tokens}p / {span.completion_tokens}c</span>
					{/if}
				</div>
			</div>
			<div class="metric-card">
				<Coins size={16} class="metric-icon" />
				<div class="metric-info">
					<span class="metric-label">Cost</span>
					<span class="metric-value">{formatCost(span.cost)}</span>
				</div>
			</div>
		</div>

		<!-- Tabs -->
		<div class="tabs-header">
			<button
				class="tab-btn"
				class:active={activeTab === 'io'}
				onclick={() => (activeTab = 'io')}
			>
				Input & Output
			</button>
			<button
				class="tab-btn"
				class:active={activeTab === 'metadata'}
				onclick={() => (activeTab = 'metadata')}
			>
				Metadata
			</button>
			<button
				class="tab-btn"
				class:active={activeTab === 'evals'}
				onclick={() => (activeTab = 'evals')}
			>
				Evaluations ({evaluations.length})
			</button>
		</div>

		<!-- Tab Content -->
		<div class="tab-content">
			{#if activeTab === 'io'}
				<div class="io-tab">
					{#if span.error_message}
						<div class="error-banner">
							<span class="error-title">Error Message</span>
							<pre class="error-text">{span.error_message}</pre>
						</div>
					{/if}

					<div class="code-section">
						<span class="section-title">Input</span>
						<pre class="monospace-code">{tryFormatJson(span.input)}</pre>
					</div>

					<div class="code-section">
						<span class="section-title">Output</span>
						<pre class="monospace-code">{tryFormatJson(span.output)}</pre>
					</div>
				</div>
			{:else if activeTab === 'metadata'}
				<div class="metadata-tab">
					{#if span.metadata && Object.keys(span.metadata).length > 0}
						<table class="metadata-table">
							<thead>
								<tr>
									<th>Key</th>
									<th>Value</th>
								</tr>
							</thead>
							<tbody>
								{#each Object.entries(span.metadata) as [key, val]}
									<tr>
										<td class="meta-key">{key}</td>
										<td class="meta-val">{typeof val === 'object' ? JSON.stringify(val) : val}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					{:else}
						<p class="no-meta">No metadata available for this span.</p>
					{/if}
				</div>
			{:else if activeTab === 'evals'}
				<div class="evals-tab">
					<!-- Feedback submission -->
					<div class="feedback-form">
						<h4 class="form-title">Submit Evaluation</h4>
						<textarea
							class="ql-input feedback-textarea"
							placeholder="Add optional comments or feedback text..."
							bind:value={feedbackText}
						></textarea>
						<div class="feedback-actions">
							<button
								class="ql-btn-secondary feedback-btn thumbs-down-btn"
								disabled={submittingEval}
								onclick={() => submitFeedback(0)}
							>
								<ThumbsDown size={16} /> Dislike
							</button>
							<button
								class="ql-btn-primary feedback-btn thumbs-up-btn"
								disabled={submittingEval}
								onclick={() => submitFeedback(1)}
							>
								<ThumbsUp size={16} /> Like
							</button>
						</div>
					</div>

					<!-- Existing evaluations -->
					<div class="evals-list-wrapper">
						<h4 class="form-title">Evaluations History</h4>
						{#if loadingEvals}
							<div class="loading-spinner">Loading evaluations...</div>
						{:else if evaluations.length > 0}
							<div class="evals-list">
								{#each evaluations as ev (ev.id)}
									<div class="eval-card">
										<div class="eval-header">
											<span class="eval-type">{ev.score_type}</span>
											<span class="eval-score" class:score-positive={ev.score_value > 0} class:score-negative={ev.score_value === 0}>
												{ev.score_value === 1 ? '👍' : ev.score_value === 0 ? '👎' : ev.score_value}
											</span>
										</div>
										{#if ev.feedback_text}
											<p class="eval-comment">{ev.feedback_text}</p>
										{/if}
										<div class="eval-meta">
											<span>By {ev.evaluator}</span>
											<span>{new Date(ev.created_at).toLocaleDateString()}</span>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<p class="no-meta">No evaluations recorded yet.</p>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	</div>
{:else}
	<div class="no-span-selected">
		<Cpu size={32} class="no-selected-icon" />
		<p>Select a span in the tree to view its input, output, metadata, and evaluations.</p>
	</div>
{/if}

<style>
	.span-detail-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		gap: 1.25rem;
	}

	.panel-header {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		border-bottom: 1px solid var(--ql-border);
		padding-bottom: 0.75rem;
	}

	.panel-header-title {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.span-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--ql-text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 70%;
	}

	.badges-row {
		display: flex;
		gap: 0.35rem;
	}

	.span-model-details {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.75rem;
	}

	.metric-card {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
	}

	.metric-icon {
		color: var(--ql-accent);
		flex-shrink: 0;
	}

	.metric-info {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.metric-label {
		font-size: 0.65rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.metric-value {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.metric-sub {
		font-size: 0.65rem;
		color: var(--ql-text-muted);
		white-space: nowrap;
	}

	.tabs-header {
		display: flex;
		border-bottom: 1px solid var(--ql-border);
	}

	.tab-btn {
		background: none;
		border: none;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--ql-text-muted);
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition: all 0.2s;
	}

	.tab-btn:hover {
		color: var(--ql-text);
	}

	.tab-btn.active {
		color: var(--ql-accent);
		border-bottom-color: var(--ql-accent);
	}

	.tab-content {
		flex-grow: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
	}

	.io-tab {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.error-banner {
		background-color: rgba(239, 68, 68, 0.08);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: 0.375rem;
		padding: 0.75rem;
	}

	.error-title {
		font-size: 0.75rem;
		font-weight: 600;
		color: #fca5a5;
		text-transform: uppercase;
		display: block;
		margin-bottom: 0.25rem;
	}

	.error-text {
		font-size: 0.825rem;
		color: #fca5a5;
		white-space: pre-wrap;
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
	}

	.code-section {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.section-title {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.monospace-code {
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		font-size: 0.8rem;
		padding: 0.75rem;
		background-color: #0f1117;
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		color: var(--ql-text);
		white-space: pre-wrap;
		overflow-x: auto;
		max-height: 250px;
	}

	.metadata-tab {
		display: flex;
		flex-direction: column;
	}

	.metadata-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.825rem;
	}

	.metadata-table th,
	.metadata-table td {
		padding: 0.5rem 0.75rem;
		text-align: left;
		border-bottom: 1px solid var(--ql-border);
	}

	.metadata-table th {
		color: var(--ql-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		font-size: 0.7rem;
	}

	.meta-key {
		color: var(--ql-text-muted);
		font-weight: 500;
		width: 35%;
	}

	.meta-val {
		color: var(--ql-text);
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		word-break: break-all;
	}

	.no-meta {
		text-align: center;
		color: var(--ql-text-muted);
		font-size: 0.875rem;
		padding: 2rem 0;
	}

	.evals-tab {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.feedback-form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		padding: 1rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		border-radius: 0.5rem;
	}

	.form-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.feedback-textarea {
		min-height: 4rem;
		font-size: 0.825rem;
		resize: vertical;
	}

	.feedback-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}

	.feedback-btn {
		font-size: 0.75rem;
		padding: 0.375rem 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.thumbs-down-btn:hover {
		border-color: var(--ql-danger, #ef4444);
		color: #ef4444;
	}

	.thumbs-up-btn {
		background-color: var(--ql-accent);
	}

	.evals-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.eval-card {
		padding: 0.75rem;
		background-color: rgba(255, 255, 255, 0.01);
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.eval-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.eval-type {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.eval-score {
		font-size: 1rem;
	}

	.eval-comment {
		font-size: 0.825rem;
		color: var(--ql-text);
	}

	.eval-meta {
		display: flex;
		justify-content: space-between;
		font-size: 0.65rem;
		color: var(--ql-text-muted);
	}

	.no-span-selected {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		height: 100%;
		padding: 3rem 1.5rem;
		color: var(--ql-text-muted);
	}

	.no-selected-icon {
		margin-bottom: 1rem;
		opacity: 0.3;
	}

	.loading-spinner {
		text-align: center;
		font-size: 0.875rem;
		color: var(--ql-text-muted);
		padding: 1rem 0;
	}
</style>
