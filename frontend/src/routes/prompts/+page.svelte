<script lang="ts">
	import { onMount } from 'svelte';
	import { getPrompts, createPrompt, deletePrompt, diffPromptVersions, type Prompt } from '$api/prompts';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import EmptyState from '$components/shared/EmptyState.svelte';
	import PromptEditor from '$components/prompts/PromptEditor.svelte';
	import VersionDiff from '$components/prompts/VersionDiff.svelte';
	import { MessageSquare, Plus, GitCompare, Trash2, Calendar, ChevronDown, ChevronUp } from 'lucide-svelte';

	let prompts = $state([] as Prompt[]);
	let loading = $state(true);

	// Expanded prompt states by ID
	let expandedPromptID = $state('');

	// Modal States
	let isEditorOpen = $state(false);
	let isDiffOpen = $state(false);
	let selectedPromptForEdit = $state(null as any);
	let diffPromptName = $state('');
	let diffPromptID = $state('');
	let diffVersionA = $state(1);
	let diffVersionB = $state(1);
	let diffContentA = $state('');
	let diffContentB = $state('');
	let diffVersionsList = $state([] as any[]);

	async function loadPrompts() {
		try {
			prompts = await getPrompts();
		} catch (err) {
			console.error(err);
			addToast('Failed to load prompts.', 'error');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadPrompts();
	});

	function toggleExpand(id: string) {
		expandedPromptID = expandedPromptID === id ? '' : id;
	}

	async function handleSavePrompt(data: any) {
		try {
			await createPrompt(data);
			addToast('Prompt saved successfully!', 'success');
			isEditorOpen = false;
			loadPrompts();
		} catch (err) {
			console.error(err);
			addToast('Failed to save prompt.', 'error');
		}
	}

	async function handleDelete(id: string, event: MouseEvent) {
		event.stopPropagation();
		if (!confirm('Are you sure you want to delete this prompt version?')) return;
		try {
			await deletePrompt(id);
			addToast('Prompt version deleted.', 'success');
			loadPrompts();
		} catch (err) {
			console.error(err);
			addToast('Failed to delete prompt.', 'error');
		}
	}

	async function openDiff(prompt: Prompt) {
		if (!prompt.versions || prompt.versions.length < 2) {
			addToast('Need at least 2 versions to compare.', 'warning');
			return;
		}

		diffPromptID = prompt.id;
		diffPromptName = prompt.name;
		diffVersionsList = prompt.versions;

		// Default to comparing the latest two versions
		diffVersionA = prompt.versions[1].version;
		diffVersionB = prompt.versions[0].version;

		await loadDiffContents(diffPromptID, diffVersionA, diffVersionB);
		isDiffOpen = true;
	}

	async function loadDiffContents(id: string, verA: number, verB: number) {
		try {
			// Svelte api expects versionA parameter as target version (the one to fetch contents for)
			// Wait, diffPromptVersions(id, version) compares the prompt of ID `id` (which is versionA) with `version` (which is versionB).
			// So, if we look up the target versions:
			// Let's find the prompt version matching versionA in the list
			const pA = prompts.find((p) => p.name === diffPromptName);
			if (!pA) return;

			// Fetch contents of version B
			const res = await diffPromptVersions(id, verA);
			diffContentA = res.content_b; // content of version B
			diffContentB = res.content_a; // content of version A
			diffVersionA = res.version_b;
			diffVersionB = res.version_a;
		} catch (err) {
			console.error(err);
			addToast('Failed to load version contents for diff.', 'error');
		}
	}

	async function handleVersionChange(verA: number, verB: number) {
		await loadDiffContents(diffPromptID, verA, verB);
	}
</script>

<div class="prompts-page">
	<!-- Header -->
	<div class="page-header-row">
		<div class="header-titles">
			<h2 class="page-title">Prompt Library</h2>
			<p class="page-subtitle">Draft, organize, and version control prompt templates for downstream applications.</p>
		</div>

		<button class="ql-btn-primary add-btn" onclick={() => {
			selectedPromptForEdit = null;
			isEditorOpen = true;
		}}>
			<Plus size={16} /> New Prompt
		</button>
	</div>

	<!-- Content -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading prompts...</span>
		</div>
	{:else if prompts.length === 0}
		<EmptyState
			title="No Prompts Registered"
			description="Prompt templates enable prompt engineering version control. Create your first template now."
			icon={MessageSquare}
			actionText="Create New Prompt"
			onAction={() => {
				selectedPromptForEdit = null;
				isEditorOpen = true;
			}}
		/>
	{:else}
		<div class="prompts-list">
			{#each prompts as prompt (prompt.id)}
				{@const isExpanded = expandedPromptID === prompt.id}
				<div class="prompt-card ql-card" class:expanded={isExpanded}>
					<!-- Expand Header -->
					<div class="prompt-card-header" onclick={() => toggleExpand(prompt.id)} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && toggleExpand(prompt.id)}>
						<div class="header-left">
							<span class="prompt-name">{prompt.name}</span>
							<div class="meta-row">
								<Badge text="v{prompt.version}" variant="info" />
								{#if prompt.model_id}
									<span class="model-id font-mono text-xs">{prompt.model_id}</span>
								{/if}
								{#if prompt.tags && prompt.tags.length > 0}
									<div class="tags-list">
										{#each prompt.tags as tag}
											<span class="tag-pill">#{tag}</span>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<div class="header-right">
							<button class="ql-btn-secondary diff-btn" onclick={(e) => {
								e.stopPropagation();
								openDiff(prompt);
							}} disabled={!prompt.versions || prompt.versions.length < 2}>
								<GitCompare size={14} /> Compare
							</button>
							<button class="ql-btn-secondary edit-btn" onclick={(e) => {
								e.stopPropagation();
								selectedPromptForEdit = prompt;
								isEditorOpen = true;
							}}>
								New Version
							</button>
							<button class="expand-toggle">
								{#if isExpanded}
									<ChevronUp size={16} />
								{:else}
									<ChevronDown size={16} />
								{/if}
							</button>
						</div>
					</div>

					<!-- Expanded Area -->
					{#if isExpanded}
						<div class="prompt-card-body">
							<div class="prompt-template-section">
								<span class="section-title">Template Content</span>
								<pre class="prompt-template-pre">{prompt.content}</pre>
							</div>

							<!-- Version History list -->
							{#if prompt.versions && prompt.versions.length > 0}
								<div class="version-history-section">
									<span class="section-title">Version History</span>
									<div class="version-list">
										{#each prompt.versions as v}
											<div class="version-row">
												<div class="ver-left">
													<Badge text="Version {v.version}" variant={v.version === prompt.version ? 'info' : 'neutral'} />
													<span class="ver-date font-mono text-xs">
														{new Date(v.created_at || '').toLocaleDateString()} {new Date(v.created_at || '').toLocaleTimeString()}
													</span>
												</div>
												{#if v.version !== prompt.version}
													<button class="delete-version-btn" onclick={(e) => handleDelete(v.id, e)} title="Delete Version">
														<Trash2 size={14} />
													</button>
												{/if}
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}

	<!-- Modals -->
	<PromptEditor
		isOpen={isEditorOpen}
		prompt={selectedPromptForEdit}
		onClose={() => (isEditorOpen = false)}
		onSave={handleSavePrompt}
	/>

	<VersionDiff
		isOpen={isDiffOpen}
		promptName={diffPromptName}
		versionA={diffVersionA}
		versionB={diffVersionB}
		contentA={diffContentA}
		contentB={diffContentB}
		versions={diffVersionsList}
		onVersionChange={handleVersionChange}
		onClose={() => (isDiffOpen = false)}
	/>
</div>

<style>
	.prompts-page {
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

	.add-btn {
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
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

	.prompts-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.prompt-card {
		padding: 0;
		display: flex;
		flex-direction: column;
		transition: border-color 0.2s;
	}

	.prompt-card.expanded {
		border-color: rgba(16, 185, 129, 0.3);
	}

	.prompt-card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.25rem;
		cursor: pointer;
		user-select: none;
	}

	.header-left {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.prompt-name {
		font-size: 1.05rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.meta-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.model-id {
		color: var(--ql-text-muted);
	}

	.tags-list {
		display: flex;
		gap: 0.35rem;
	}

	.tag-pill {
		font-size: 0.75rem;
		color: var(--ql-accent);
		font-weight: 500;
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.diff-btn {
		font-size: 0.75rem;
		padding: 0.35rem 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.edit-btn {
		font-size: 0.75rem;
		padding: 0.35rem 0.75rem;
	}

	.expand-toggle {
		background: none;
		border: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		padding: 0.25rem;
	}

	.prompt-card-body {
		padding: 1.25rem;
		border-top: 1px solid var(--ql-border);
		background-color: rgba(15, 17, 23, 0.2);
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.prompt-template-section {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.section-title {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.prompt-template-pre {
		background-color: #0f1117;
		padding: 1rem;
		border-radius: 0.375rem;
		border: 1px solid var(--ql-border);
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		font-size: 0.85rem;
		color: var(--ql-text);
		white-space: pre-wrap;
		line-height: 1.5;
	}

	.version-history-section {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.version-list {
		display: flex;
		flex-direction: column;
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		background-color: var(--ql-bg);
		overflow: hidden;
	}

	.version-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.625rem 0.875rem;
		border-bottom: 1px solid var(--ql-border);
	}

	.version-row:last-child {
		border-bottom: none;
	}

	.ver-left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.ver-date {
		color: var(--ql-text-muted);
	}

	.delete-version-btn {
		background: none;
		border: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 0.25rem;
		transition: all 0.2s;
	}

	.delete-version-btn:hover {
		color: var(--ql-danger, #ef4444);
		background-color: var(--ql-surface-2);
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
