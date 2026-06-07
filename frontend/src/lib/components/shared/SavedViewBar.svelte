<script lang="ts">
	import { Bookmark, Filter } from 'lucide-svelte';
	import type { SavedView } from '$api/operations';

	let { views = [], activeScope = 'overview' } = $props<{
		views: SavedView[];
		activeScope?: string;
	}>();
</script>

<div class="saved-view-bar" aria-label="Saved views">
	<span class="label"><Bookmark size={15} /> Saved Views</span>
	<a class="view-chip active" href={`/${activeScope === 'overview' ? '' : activeScope}`}>Default</a>
	{#each views.slice(0, 4) as view}
		<a class="view-chip" href={`/${view.scope}?view=${view.id}`}>
			<Filter size={13} />
			{view.name}
		</a>
	{:else}
		<span class="empty-chip">No shared views yet</span>
	{/each}
</div>

<style>
	.saved-view-bar {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		overflow-x: auto;
		padding-bottom: 0.125rem;
	}

	.label,
	.view-chip,
	.empty-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		white-space: nowrap;
		font-size: 0.8rem;
	}

	.label,
	.empty-chip {
		color: var(--ql-text-muted);
	}

	.view-chip {
		padding: 0.375rem 0.625rem;
		border: 1px solid var(--ql-border);
		border-radius: 999px;
		color: var(--ql-text);
		text-decoration: none;
		background: var(--ql-surface);
	}

	.view-chip.active {
		color: var(--ql-accent);
		border-color: rgba(16, 185, 129, 0.35);
		background: var(--ql-accent-subtle);
	}
</style>
