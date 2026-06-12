<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getModels, createModel, deleteModel, discoverModels, type Model } from '$api/models';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import EmptyState from '$components/shared/EmptyState.svelte';
	import { createWebSocket } from '$lib/websocket/client';
	import { Plus, Search, Trash2, Cpu, RefreshCw, X, Link, Server } from 'lucide-svelte';

	let modelsList = $state([] as Model[]);
	let searchQuery = $state('');
	let loading = $state(true);
	let wsClient: any = null;

	// Modal State
	let isModalOpen = $state(false);
	let name = $state('');
	let provider = $state('openai');
	let modelID = $state('');
	let endpoint = $state('');
	let submitting = $state(false);

	let filteredModels = $derived(
		modelsList.filter(
			(m) =>
				m.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				m.provider.toLowerCase().includes(searchQuery.toLowerCase()) ||
				m.model_id.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	async function loadModels() {
		try {
			modelsList = await getModels();
		} catch (err) {
			console.error(err);
			addToast('Failed to load models list.', 'error');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadModels();

		// WebSocket listener for live model status/polling updates
		wsClient = createWebSocket('models', (msg: any) => {
			loadModels();
		});
	});

	onDestroy(() => {
		if (wsClient) wsClient.close();
	});

	async function handleDiscover() {
		addToast('Discovering Ollama models...', 'info');
		try {
			await discoverModels();
			addToast('Ollama models discovery triggered!', 'success');
			loadModels();
		} catch (err) {
			console.error(err);
			addToast('Discovery failed. Make sure Ollama is running.', 'error');
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Are you sure you want to delete this model?')) return;
		try {
			await deleteModel(id);
			addToast('Model deleted successfully.', 'success');
			loadModels();
		} catch (err) {
			console.error(err);
			addToast('Failed to delete model.', 'error');
		}
	}

	async function handleCreateModel(e: SubmitEvent) {
		e.preventDefault();
		if (!name || !modelID) return;

		submitting = true;
		try {
			await createModel({
				name,
				provider,
				model_id: modelID,
				endpoint: endpoint || undefined
			});
			addToast('Model registered successfully!', 'success');
			isModalOpen = false;
			name = '';
			modelID = '';
			endpoint = '';
			loadModels();
		} catch (err) {
			console.error(err);
			addToast('Failed to register model.', 'error');
		} finally {
			submitting = false;
		}
	}
</script>

<div class="models-page">
	<!-- Header -->
	<div class="page-header-row">
		<div class="header-titles">
			<h2 class="page-title">Model Registry</h2>
			<p class="page-subtitle">Manage LLM configurations, poll statuses, and configure local models.</p>
		</div>

		<div class="action-buttons">
			<button class="ql-btn-secondary discover-btn" onclick={handleDiscover}>
				<RefreshCw size={16} /> Discover Ollama
			</button>
			<button class="ql-btn-primary add-btn" onclick={() => (isModalOpen = true)}>
				<Plus size={16} /> Register Model
			</button>
		</div>
	</div>

	<!-- Filter Bar -->
	<div class="filter-search-bar ql-card">
		<div class="search-input-wrapper">
			<Search size={18} class="search-icon" />
			<input
				type="text"
				class="ql-input search-input"
				placeholder="Search models by name, ID, or provider..."
				bind:value={searchQuery}
			/>
		</div>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading models...</span>
		</div>
	{:else if filteredModels.length === 0}
		<EmptyState
			title="No Models Registered"
			description="No models match your search criteria or are registered in QuickLens yet."
			icon={Cpu}
			actionText="Register Your First Model"
			onAction={() => (isModalOpen = true)}
		/>
	{:else}
		<div class="models-grid">
			{#each filteredModels as model (model.id)}
				<div class="ql-card model-card">
					<div class="model-header">
						<div class="model-info">
							<h3 class="model-title">{model.name}</h3>
							<span class="provider-pill provider-{model.provider.toLowerCase()}">
								{model.provider}
							</span>
						</div>
						<div class="model-header-actions">
							{#if model.status === 'online'}
								<span class="status-badge online">Online</span>
							{:else}
								<span class="status-badge offline">Offline</span>
							{/if}
							<button
								class="delete-btn"
								onclick={() => handleDelete(model.id)}
								title="Delete Model"
							>
								<Trash2 size={15} />
							</button>
						</div>
					</div>

					<!-- Details -->
					<div class="model-body">
						<div class="detail-row">
							<span class="detail-label">Model ID</span>
							<span class="detail-val font-mono">{model.model_id}</span>
						</div>
						{#if model.endpoint}
							<div class="detail-row">
								<span class="detail-label">Endpoint</span>
								<span class="detail-val font-mono text-xs" title={model.endpoint}>
									{model.endpoint}
								</span>
							</div>
						{/if}
					</div>

					<!-- Metrics -->
					<div class="model-stats">
						<div class="stat-box">
							<span class="stat-num">{model.total_requests?.toLocaleString() || 0}</span>
							<span class="stat-lbl">Requests</span>
						</div>
						<div class="stat-box">
							<span class="stat-num">{model.total_tokens?.toLocaleString() || 0}</span>
							<span class="stat-lbl">Tokens</span>
						</div>
						<div class="stat-box">
							<span class="stat-num">{model.avg_latency_ms ? `${model.avg_latency_ms.toFixed(0)}ms` : '0ms'}</span>
							<span class="stat-lbl">Latency</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Register Model Modal -->
	{#if isModalOpen}
		<div class="modal-backdrop" onclick={() => (isModalOpen = false)} role="presentation">
			<div class="modal-container ql-card" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
				<div class="modal-header">
					<h3 class="modal-title">Register Custom Model</h3>
					<button class="close-btn" onclick={() => (isModalOpen = false)}>
						<X size={18} />
					</button>
				</div>

				<form onsubmit={handleCreateModel} class="modal-form">
					<div class="form-group">
						<label for="reg-name" class="form-label">Display Name</label>
						<input
							type="text"
							id="reg-name"
							class="ql-input"
							placeholder="e.g. My ChatGPT"
							bind:value={name}
							required
						/>
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="reg-provider" class="form-label">Provider</label>
							<select id="reg-provider" class="ql-input" bind:value={provider}>
								<option value="openai">OpenAI</option>
								<option value="anthropic">Anthropic</option>
								<option value="google">Google Gemini</option>
								<option value="ollama">Ollama (Local)</option>
								<option value="mistral">Mistral</option>
								<option value="cohere">Cohere</option>
								<option value="custom">Custom Upstream</option>
							</select>
						</div>

						<div class="form-group">
							<label for="reg-id" class="form-label">Model ID / Identifier</label>
							<input
								type="text"
								id="reg-id"
								class="ql-input"
								placeholder="e.g. gpt-4o"
								bind:value={modelID}
								required
							/>
						</div>
					</div>

					<div class="form-group">
						<label for="reg-endpoint" class="form-label">Endpoint URL (Optional)</label>
						<input
							type="text"
							id="reg-endpoint"
							class="ql-input"
							placeholder="e.g. http://localhost:11434"
							bind:value={endpoint}
						/>
					</div>

					<div class="form-actions">
						<button type="button" class="ql-btn-secondary" onclick={() => (isModalOpen = false)}>
							Cancel
						</button>
						<button type="submit" class="ql-btn-primary" disabled={submitting}>
							{submitting ? 'Registering...' : 'Register Model'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>

<style>
	.models-page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-titles {
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

	.action-buttons {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.discover-btn {
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.add-btn {
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
		background-color: var(--ql-accent);
	}

	.filter-search-bar {
		padding: 0.75rem 1rem;
	}

	.search-input-wrapper {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background-color: var(--ql-surface-2);
		padding: 0.5rem 0.75rem;
		border-radius: 0.375rem;
		border: 1px solid var(--ql-border);
	}

	.search-icon {
		color: var(--ql-text-muted);
	}

	.search-input {
		border: none;
		padding: 0;
		background: none;
		flex-grow: 1;
		font-size: 0.875rem;
		color: var(--ql-text);
	}

	.search-input:focus {
		outline: none;
		box-shadow: none;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 250px;
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

	.models-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.25rem;
	}

	.model-card {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		transition: transform 0.2s, border-color 0.2s;
	}

	.model-card:hover {
		border-color: rgba(16, 185, 129, 0.4);
		transform: translateY(-2px);
	}

	.model-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		border-bottom: 1px solid var(--ql-border);
		padding-bottom: 0.75rem;
	}

	.model-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		min-width: 0;
	}

	.model-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
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
		width: max-content;
	}

	.model-header-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.status-badge {
		font-size: 0.7rem;
		font-weight: 600;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
	}

	.status-badge.online {
		background-color: rgba(34, 197, 94, 0.1);
		color: #4ade80;
	}

	.status-badge.offline {
		background-color: rgba(148, 163, 184, 0.1);
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
	}

	.delete-btn:hover {
		color: var(--ql-danger, #ef4444);
		background-color: var(--ql-surface-2);
	}

	.model-body {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		font-size: 0.8rem;
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.5rem;
	}

	.detail-label {
		color: var(--ql-text-muted);
	}

	.detail-val {
		color: var(--ql-text);
		max-width: 70%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.model-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.5rem;
		border-top: 1px solid var(--ql-border);
		padding-top: 0.75rem;
		text-align: center;
	}

	.stat-box {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.stat-num {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--ql-text);
	}

	.stat-lbl {
		font-size: 0.65rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	/* Modal styling */
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: rgba(15, 17, 23, 0.7);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-container {
		width: 100%;
		max-width: 500px;
		display: flex;
		flex-direction: column;
		padding: 1.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
		border-color: var(--ql-border);
		animation: scale-up 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid var(--ql-border);
		padding-bottom: 0.75rem;
		margin-bottom: 1.25rem;
	}

	.modal-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.close-btn {
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
	}

	.close-btn:hover {
		color: var(--ql-text);
		background-color: var(--ql-surface-2);
	}

	.modal-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--ql-text-muted);
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		border-top: 1px solid var(--ql-border);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@keyframes scale-up {
		from {
			transform: scale(0.95);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}
</style>
