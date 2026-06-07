<script lang="ts">
	import Badge from '../shared/Badge.svelte';
	import { Cpu, CheckCircle, AlertTriangle, Clock } from 'lucide-svelte';

	let { models = [] }: { models: any[] } = $props();

	function formatBytes(bytes: number): string {
		if (bytes === 0 || !bytes) return 'N/A';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatTime(timeStr: string | null): string {
		if (!timeStr) return 'Never';
		try {
			const date = new Date(timeStr);
			return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} catch {
			return 'Invalid date';
		}
	}
</script>

<div class="model-grid">
	{#each models as model (model.id)}
		<div class="ql-card model-card">
			<div class="model-card-header">
				<div class="model-meta">
					<span class="model-name">{model.name}</span>
					<div class="provider-badge-wrapper">
						<span class="provider-badge provider-{model.provider.toLowerCase()}">
							{model.provider}
						</span>
					</div>
				</div>
				<div class="model-status">
					{#if model.status === 'online'}
						<div class="status-indicator online">
							<span class="dot"></span> Online
						</div>
					{:else}
						<div class="status-indicator offline">
							<span class="dot"></span> Offline
						</div>
					{/if}
				</div>
			</div>

			<div class="model-stats">
				<div class="stat-item">
					<span class="label">Tokens</span>
					<span class="value">{model.total_tokens?.toLocaleString() || 0}</span>
				</div>
				<div class="stat-item">
					<span class="label">Requests</span>
					<span class="value">{model.total_requests?.toLocaleString() || 0}</span>
				</div>
				<div class="stat-item">
					<span class="label">Avg Latency</span>
					<span class="value">{model.avg_latency_ms ? `${model.avg_latency_ms.toFixed(0)}ms` : '0ms'}</span>
				</div>
			</div>

			<div class="model-footer">
				<div class="footer-item" title="Size">
					<Cpu size={14} />
					<span>{model.quantization || 'Unknown'} ({formatBytes(model.size_bytes)})</span>
				</div>
				<div class="footer-item" title="Last Seen">
					<Clock size={14} />
					<span>Seen: {formatTime(model.last_seen_at)}</span>
				</div>
			</div>
		</div>
	{/each}
</div>

<style>
	.model-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
		width: 100%;
	}

	.model-card {
		padding: 1rem;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		gap: 1rem;
	}

	.model-card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}

	.model-meta {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.model-name {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--ql-text);
		word-break: break-all;
	}

	.provider-badge-wrapper {
		display: flex;
		gap: 0.25rem;
	}

	.provider-badge {
		font-size: 0.65rem;
		font-weight: 700;
		text-transform: uppercase;
		padding: 0.1rem 0.35rem;
		border-radius: 0.25rem;
		letter-spacing: 0.05em;
	}

	.provider-ollama {
		background-color: rgba(255, 255, 255, 0.1);
		color: #e2e8f0;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.provider-openai {
		background-color: rgba(16, 163, 127, 0.1);
		color: #10a37f;
		border: 1px solid rgba(16, 163, 127, 0.2);
	}

	.provider-anthropic {
		background-color: rgba(241, 151, 120, 0.1);
		color: #f19778;
		border: 1px solid rgba(241, 151, 120, 0.2);
	}

	.status-indicator {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.status-indicator .dot {
		width: 6px;
		height: 6px;
		border-radius: 9999px;
	}

	.status-indicator.online {
		color: #4ade80;
	}

	.status-indicator.online .dot {
		background-color: var(--ql-success, #22c55e);
		box-shadow: 0 0 8px var(--ql-success, #22c55e);
	}

	.status-indicator.offline {
		color: var(--ql-text-muted);
	}

	.status-indicator.offline .dot {
		background-color: #64748b;
	}

	.model-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.5rem;
		background-color: rgba(15, 17, 23, 0.5);
		padding: 0.5rem;
		border-radius: 0.375rem;
		border: 1px solid var(--ql-border);
		text-align: center;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.stat-item .label {
		font-size: 0.65rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.stat-item .value {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.model-footer {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		border-top: 1px solid var(--ql-border);
		padding-top: 0.5rem;
		font-size: 0.75rem;
		color: var(--ql-text-muted);
	}

	.footer-item {
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}
</style>
