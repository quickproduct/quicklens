<script lang="ts">
	import { BarChart3, Cpu, MessageSquare, Star } from 'lucide-svelte';
</script>

<div class="monitoring-page">
	<section class="page-header">
		<div>
			<h2>Monitoring</h2>
			<p>LLM operations grouped by runtime inventory, prompt quality, feedback, and live logs.</p>
		</div>
	</section>

	<div class="monitoring-grid">
		<a class="monitoring-card" href="/models">
			<Cpu size={22} />
			<div>
				<h3>Model Registry</h3>
				<p>Manage providers, endpoints, online status, latency, token volume, and local discovery.</p>
			</div>
		</a>
		<a class="monitoring-card" href="/prompts">
			<MessageSquare size={22} />
			<div>
				<h3>Prompt Operations</h3>
				<p>Review prompt versions, compare changes, and manage reusable prompt assets.</p>
			</div>
		</a>
		<a class="monitoring-card" href="/evaluations">
			<Star size={22} />
			<div>
				<h3>Evaluations</h3>
				<p>Track feedback and quality signals from spans, traces, and manual review.</p>
			</div>
		</a>
		<a class="monitoring-card" href="/logs">
			<BarChart3 size={22} />
			<div>
				<h3>Live Logs</h3>
				<p>Search operational logs, failures, and recent model activity.</p>
			</div>
		</a>
	</div>
</div>

<style>
	.monitoring-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	h2,
	h3,
	p {
		margin: 0;
	}

	h2 {
		font-size: 1.35rem;
	}

	.page-header p,
	.monitoring-card p {
		color: var(--ql-text-muted);
		font-size: 0.875rem;
	}

	.monitoring-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
	}

	.monitoring-card {
		display: flex;
		align-items: flex-start;
		gap: 0.875rem;
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
		color: var(--ql-text);
		text-decoration: none;
	}

	.monitoring-card:hover,
	.monitoring-card:focus-visible {
		border-color: var(--ql-accent);
		outline: none;
	}

	.monitoring-card svg {
		color: var(--ql-accent);
		flex-shrink: 0;
	}

	.monitoring-card div {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	@media (max-width: 720px) {
		.monitoring-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
