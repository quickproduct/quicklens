<script lang="ts">
	import { goto } from '$app/navigation';
	import { Search, Zap } from 'lucide-svelte';

	let { open = false, onClose } = $props<{
		open: boolean;
		onClose: () => void;
	}>();

	let query = $state('');

	const actions = [
		{ label: 'Open Overview', href: '/', keywords: 'dashboard health overview' },
		{ label: 'Open Monitoring', href: '/monitoring', keywords: 'models prompts evaluations llm ops' },
		{ label: 'Open Traces', href: '/traces', keywords: 'debug spans latency tokens' },
		{ label: 'Open Alerts', href: '/alerts', keywords: 'alert rules acknowledge severity' },
		{ label: 'Open Incidents', href: '/incidents', keywords: 'incident response timeline resolve' },
		{ label: 'Open Analytics', href: '/analytics', keywords: 'slo trends cost tokens burn rate' },
		{ label: 'Open Reports', href: '/reports', keywords: 'exports scheduled dashboards' },
		{ label: 'Open Settings', href: '/settings', keywords: 'rbac profile audit tenants' }
	];

	let filtered = $derived(
		actions.filter((action) => {
			const target = `${action.label} ${action.keywords}`.toLowerCase();
			return target.includes(query.toLowerCase());
		})
	);

	function openAction(href: string) {
		onClose();
		goto(href);
	}
</script>

{#if open}
	<div class="palette-backdrop" role="presentation" onclick={onClose} onkeydown={(e) => e.key === 'Escape' && onClose()}>
		<div
			class="palette"
			role="dialog"
			aria-modal="true"
			aria-label="Command palette"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<div class="palette-search">
				<Search size={18} />
				<input
					bind:value={query}
					placeholder="Search pages, actions, and workflows..."
					onkeydown={(e) => {
						if (e.key === 'Escape') onClose();
						if (e.key === 'Enter' && filtered[0]) openAction(filtered[0].href);
					}}
				/>
			</div>
			<div class="palette-list">
				{#each filtered as action}
					<button type="button" onclick={() => openAction(action.href)}>
						<Zap size={15} />
						<span>{action.label}</span>
					</button>
				{:else}
					<div class="palette-empty">No matching actions.</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style>
	.palette-backdrop {
		position: fixed;
		inset: 0;
		z-index: 90;
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 12vh;
		background: rgba(0, 0, 0, 0.55);
	}

	.palette {
		width: min(640px, calc(100vw - 32px));
		border: 1px solid var(--ql-border);
		border-radius: 10px;
		background: var(--ql-surface);
		box-shadow: var(--ql-shadow-lg);
		overflow: hidden;
	}

	.palette-search {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.875rem 1rem;
		border-bottom: 1px solid var(--ql-border);
	}

	.palette-search svg {
		color: var(--ql-text-muted);
	}

	.palette-search input {
		width: 100%;
		border: 0;
		background: transparent;
		color: var(--ql-text);
		font: inherit;
		outline: none;
	}

	.palette-list {
		display: flex;
		flex-direction: column;
		padding: 0.5rem;
	}

	.palette-list button {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		width: 100%;
		padding: 0.75rem;
		border: 1px solid transparent;
		border-radius: 6px;
		background: transparent;
		color: var(--ql-text);
		cursor: pointer;
		text-align: left;
		font: inherit;
	}

	.palette-list button:hover,
	.palette-list button:focus-visible {
		border-color: var(--ql-border);
		background: var(--ql-accent-subtle);
		outline: none;
	}

	.palette-empty {
		padding: 1rem;
		color: var(--ql-text-muted);
		font-size: 0.875rem;
	}
</style>
