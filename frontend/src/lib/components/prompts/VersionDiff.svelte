<script lang="ts">
	import { X } from 'lucide-svelte';

	let {
		isOpen = false,
		promptName = '',
		versionA = 1,
		versionB = 2,
		contentA = '',
		contentB = '',
		versions = [] as any[],
		onVersionChange = (a: number, b: number) => {},
		onClose = () => {}
	}: {
		isOpen: boolean;
		promptName: string;
		versionA: number;
		versionB: number;
		contentA: string;
		contentB: string;
		versions: any[];
		onVersionChange: (a: number, b: number) => void;
		onClose: () => void;
	} = $props();

	// Split contents by lines for comparison
	let linesA = $derived(contentA.split('\n'));
	let linesB = $derived(contentB.split('\n'));
	let maxLines = $derived(Math.max(linesA.length, linesB.length));

	function handleVersionAChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		onVersionChange(parseInt(target.value), versionB);
	}

	function handleVersionBChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		onVersionChange(versionA, parseInt(target.value));
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" onclick={onClose} role="presentation">
		<div class="modal-container ql-card" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
			<!-- Header -->
			<div class="modal-header">
				<div class="header-titles">
					<h3 class="modal-title">Compare Versions — {promptName}</h3>
				</div>
				<button class="close-btn" onclick={onClose}>
					<X size={18} />
				</button>
			</div>

			<!-- Selectors -->
			<div class="diff-selectors">
				<div class="selector-group">
					<label for="version-a-select" class="selector-label">Base Version (Left)</label>
					<select id="version-a-select" class="ql-input version-select" value={versionA} onchange={handleVersionAChange}>
						{#each versions as v}
							<option value={v.version}>Version {v.version}</option>
						{/each}
					</select>
				</div>
				<div class="selector-divider">vs</div>
				<div class="selector-group">
					<label for="version-b-select" class="selector-label">Target Version (Right)</label>
					<select id="version-b-select" class="ql-input version-select" value={versionB} onchange={handleVersionBChange}>
						{#each versions as v}
							<option value={v.version}>Version {v.version}</option>
						{/each}
					</select>
				</div>
			</div>

			<!-- Diff Columns -->
			<div class="diff-view-area">
				<div class="diff-pane pane-left">
					<div class="pane-header">Version {versionA} (Deletions)</div>
					<div class="pane-code">
						{#each Array(maxLines) as _, i}
							{@const line = linesA[i]}
							{@const isDiff = line !== linesB[i]}
							<div class="code-line" class:diff-removed={isDiff && line !== undefined}>
								<span class="line-number">{i + 1}</span>
								<span class="line-text">{line !== undefined ? line : ''}</span>
							</div>
						{/each}
					</div>
				</div>

				<div class="diff-pane pane-right">
					<div class="pane-header">Version {versionB} (Additions)</div>
					<div class="pane-code">
						{#each Array(maxLines) as _, i}
							{@const line = linesB[i]}
							{@const isDiff = line !== linesA[i]}
							<div class="code-line" class:diff-added={isDiff && line !== undefined}>
								<span class="line-number">{i + 1}</span>
								<span class="line-text">{line !== undefined ? line : ''}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
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
		width: 90%;
		max-width: 1100px;
		height: 80vh;
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
		margin-bottom: 1rem;
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

	.diff-selectors {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.75rem 1rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		margin-bottom: 1.25rem;
	}

	.selector-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		flex-grow: 1;
	}

	.selector-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.version-select {
		padding: 0.375rem;
		font-size: 0.875rem;
		border-color: var(--ql-border);
	}

	.selector-divider {
		font-weight: 600;
		color: var(--ql-text-muted);
		margin-top: 1rem;
	}

	.diff-view-area {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		flex-grow: 1;
		min-height: 0;
	}

	.diff-pane {
		display: flex;
		flex-direction: column;
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		overflow: hidden;
		background-color: #0f1117;
	}

	.pane-header {
		padding: 0.5rem 0.75rem;
		background-color: var(--ql-surface-2);
		border-bottom: 1px solid var(--ql-border);
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
	}

	.pane-code {
		padding: 0.5rem 0;
		overflow-y: auto;
		flex-grow: 1;
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		font-size: 0.8rem;
		line-height: 1.5;
	}

	.code-line {
		display: flex;
		padding: 0 0.75rem;
		white-space: pre-wrap;
	}

	.line-number {
		width: 2rem;
		color: rgba(148, 163, 184, 0.4);
		text-align: right;
		margin-right: 1rem;
		user-select: none;
		border-right: 1px solid rgba(46, 51, 72, 0.3);
		padding-right: 0.5rem;
	}

	.line-text {
		color: var(--ql-text);
		word-break: break-all;
	}

	.diff-removed {
		background-color: rgba(239, 68, 68, 0.15);
	}
	.diff-removed .line-text {
		color: #fca5a5;
		text-decoration: line-through;
	}

	.diff-added {
		background-color: rgba(16, 185, 129, 0.15);
	}
	.diff-added .line-text {
		color: #86efac;
	}

	@keyframes scale-up {
		from {
			transform: scale(0.98);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}
</style>


