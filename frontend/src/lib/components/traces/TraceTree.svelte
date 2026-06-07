<script lang="ts">
	import Badge from '../shared/Badge.svelte';
	import { ChevronDown, ChevronRight, Cpu, Search, Database, Terminal, User } from 'lucide-svelte';
	import TraceTree from './TraceTree.svelte';

	let {
		spans = [] as any[],
		selectedSpanId = '',
		onSelect = (id: string) => {},
		depth = 0
	}: {
		spans: any[];
		selectedSpanId: string;
		onSelect: (id: string) => void;
		depth?: number;
	} = $props();

	// Local state for tracking collapsed state of spans by ID
	let collapsedSpans = $state(new Set<string>());

	function toggleCollapse(spanId: string, event: MouseEvent) {
		event.stopPropagation();
		if (collapsedSpans.has(spanId)) {
			collapsedSpans.delete(spanId);
		} else {
			collapsedSpans.add(spanId);
		}
		// Svelte 5 triggers reactivity on re-assignment
		collapsedSpans = new Set(collapsedSpans);
	}

	function getSpanIcon(type: string) {
		switch (type.toLowerCase()) {
			case 'llm':
				return Cpu;
			case 'retrieval':
				return Database;
			case 'embedding':
				return Search;
			case 'tool':
				return Terminal;
			case 'agent':
				return User;
			default:
				return Cpu;
		}
	}
</script>

<div class="trace-tree-container" class:root-level={depth === 0}>
	{#each spans as span (span.id)}
		{@const Icon = getSpanIcon(span.type)}
		{@const isSelected = span.id === selectedSpanId}
		{@const hasChildren = span.children && span.children.length > 0}
		{@const isCollapsed = collapsedSpans.has(span.id)}

		<div class="tree-node-wrapper">
			<!-- Row representing this span -->
			<div
				class="tree-node-row span-type-{span.type.toLowerCase()}"
				class:selected={isSelected}
				style="padding-left: {depth * 1.5}rem"
				onclick={() => onSelect(span.id)}
				role="button"
				tabindex="0"
				onkeydown={(e) => e.key === 'Enter' && onSelect(span.id)}
			>
				<!-- Connecting lines guidelines -->
				{#if depth > 0}
					<div class="guideline-connector" style="left: {(depth - 0.5) * 1.5}rem"></div>
				{/if}

				<!-- Expand/Collapse toggle -->
				<div class="toggle-space">
					{#if hasChildren}
						<button
							class="collapse-btn"
							onclick={(e) => toggleCollapse(span.id, e)}
							aria-label={isCollapsed ? 'Expand' : 'Collapse'}
						>
							{#if isCollapsed}
								<ChevronRight size={14} />
							{:else}
								<ChevronDown size={14} />
							{/if}
						</button>
					{/if}
				</div>

				<!-- Span Details -->
				<div class="span-info">
					<span class="span-icon-wrapper">
						<Icon size={14} />
					</span>
					<span class="span-name">{span.name}</span>
					<span class="span-type-pill">{span.type}</span>
					{#if span.model_id}
						<span class="span-model-name">{span.model_id}</span>
					{/if}
				</div>

				<!-- Metrics -->
				<div class="span-metrics">
					{#if span.total_tokens > 0}
						<span class="span-tokens" title="Tokens">{span.total_tokens} t</span>
					{/if}
					<span class="span-duration">{span.duration_ms}ms</span>
					<Badge
						text={span.status === 'error' ? 'Error' : 'OK'}
						variant={span.status === 'error' ? 'danger' : 'success'}
					/>
				</div>
			</div>

			<!-- Recursive Children rendering -->
			{#if hasChildren && !isCollapsed}
				<TraceTree
					spans={span.children}
					{selectedSpanId}
					{onSelect}
					depth={depth + 1}
				/>

			{/if}
		</div>
	{/each}
</div>

<style>
	.trace-tree-container {
		display: flex;
		flex-direction: column;
		width: 100%;
	}

	.root-level {
		background-color: var(--ql-surface);
		border: 1px solid var(--ql-border);
		border-radius: 0.5rem;
		overflow: hidden;
	}

	.tree-node-wrapper {
		display: flex;
		flex-direction: column;
		position: relative;
	}

	.tree-node-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.625rem 0.875rem;
		border-bottom: 1px solid var(--ql-border);
		cursor: pointer;
		position: relative;
		transition: background-color 0.15s;
		user-select: none;
	}

	.tree-node-row:hover {
		background-color: rgba(255, 255, 255, 0.02);
	}

	.tree-node-row.selected {
		background-color: rgba(16, 185, 129, 0.08);
	}

	.guideline-connector {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1px;
		background-color: var(--ql-border);
		pointer-events: none;
	}

	.toggle-space {
		width: 1.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10;
	}

	.collapse-btn {
		background: none;
		border: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.125rem;
		border-radius: 0.25rem;
	}

	.collapse-btn:hover {
		background-color: var(--ql-surface-2);
		color: var(--ql-text);
	}

	.span-info {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-grow: 1;
		min-width: 0;
	}

	.span-icon-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 1.5rem;
		height: 1.5rem;
		border-radius: 0.25rem;
	}

	.span-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--ql-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.span-type-pill {
		font-size: 0.65rem;
		font-weight: 600;
		text-transform: uppercase;
		padding: 0.05rem 0.3rem;
		border-radius: 0.125rem;
		letter-spacing: 0.025em;
	}

	.span-model-name {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.span-metrics {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-shrink: 0;
		margin-left: 1rem;
	}

	.span-tokens {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
	}

	.span-duration {
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--ql-text-muted);
		min-width: 3.5rem;
		text-align: right;
	}

	/* Color-coding by type */
	.span-type-llm .span-icon-wrapper {
		background-color: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}
	.span-type-llm .span-type-pill {
		background-color: rgba(16, 185, 129, 0.15);
		color: #10b981;
	}

	.span-type-retrieval .span-icon-wrapper {
		background-color: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	.span-type-retrieval .span-type-pill {
		background-color: rgba(59, 130, 246, 0.15);
		color: #3b82f6;
	}

	.span-type-embedding .span-icon-wrapper {
		background-color: rgba(168, 85, 247, 0.1);
		color: #a855f7;
	}
	.span-type-embedding .span-type-pill {
		background-color: rgba(168, 85, 247, 0.15);
		color: #a855f7;
	}

	.span-type-tool .span-icon-wrapper {
		background-color: rgba(249, 115, 22, 0.1);
		color: #f97316;
	}
	.span-type-tool .span-type-pill {
		background-color: rgba(249, 115, 22, 0.15);
		color: #f97316;
	}

	.span-type-agent .span-icon-wrapper {
		background-color: rgba(236, 72, 153, 0.1);
		color: #ec4899;
	}
	.span-type-agent .span-type-pill {
		background-color: rgba(236, 72, 153, 0.15);
		color: #ec4899;
	}
</style>
