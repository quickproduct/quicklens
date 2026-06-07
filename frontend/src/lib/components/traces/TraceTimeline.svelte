<script lang="ts">
	let { spans = [], totalDurationMs = 1 }: { spans: any[]; totalDurationMs: number } = $props();

	// Flatten the tree of spans to calculate absolute start offsets
	function getTraceStartTime(nodes: any[]): number {
		let minTime = Infinity;
		function walk(list: any[]) {
			for (const node of list) {
				if (node.started_at) {
					const t = new Date(node.started_at).getTime();
					if (t < minTime) minTime = t;
				}
				if (node.children) walk(node.children);
			}
		}
		walk(nodes);
		return minTime === Infinity ? Date.now() : minTime;
	}

	function flattenSpans(nodes: any[], traceStart: number, depth = 0): any[] {
		let result: any[] = [];
		for (const node of nodes) {
			const nodeStart = node.started_at ? new Date(node.started_at).getTime() : traceStart;
			const startOffset = Math.max(0, nodeStart - traceStart);

			result.push({
				id: node.id,
				name: node.name,
				type: node.type,
				duration_ms: node.duration_ms,
				start_offset_ms: startOffset,
				depth
			});

			if (node.children && node.children.length > 0) {
				result.push(...flattenSpans(node.children, traceStart, depth + 1));
			}
		}
		return result;
	}

	let traceStartTime = $derived(getTraceStartTime(spans));
	let flatSpans = $derived(flattenSpans(spans, traceStartTime));

	function getBarPosition(startOffset: number, duration: number) {
		const total = totalDurationMs || 1;
		const left = (startOffset / total) * 100;
		const width = (duration / total) * 100;

		return `left: ${Math.min(99, left)}%; width: ${Math.max(1, Math.min(100 - left, width))}%`;
	}
</script>

<div class="ql-card timeline-card">
	<div class="timeline-header">
		<h4 class="timeline-title">Trace Timeline Waterfall</h4>
		<span class="timeline-duration">Total Duration: {totalDurationMs}ms</span>
	</div>

	<div class="timeline-container">
		<!-- Grid lines -->
		<div class="grid-lines-bg">
			<div class="line"></div>
			<div class="line"></div>
			<div class="line"></div>
			<div class="line"></div>
			<div class="line"></div>
		</div>

		<!-- Rows -->
		<div class="timeline-rows">
			{#each flatSpans as span (span.id)}
				<div class="timeline-row">
					<div class="span-label" style="padding-left: {span.depth * 0.75}rem">
						<span class="span-label-name" title={span.name}>{span.name}</span>
						<span class="span-label-type">{span.type}</span>
					</div>
					<div class="bar-container">
						<div
							class="duration-bar span-type-{span.type.toLowerCase()}"
							style={getBarPosition(span.start_offset_ms, span.duration_ms)}
							title="{span.name}: {span.duration_ms}ms (offset {span.start_offset_ms}ms)"
						>
							<span class="bar-label">{span.duration_ms}ms</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<!-- Timeline legend/scale -->
	<div class="timeline-legend">
		<span>0%</span>
		<span>25%</span>
		<span>50%</span>
		<span>75%</span>
		<span>100%</span>
	</div>
</div>

<style>
	.timeline-card {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.timeline-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.timeline-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.timeline-duration {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
	}

	.timeline-container {
		position: relative;
		border: 1px solid var(--ql-border);
		border-radius: 0.375rem;
		background-color: var(--ql-bg);
		overflow: hidden;
		min-height: 100px;
	}

	.grid-lines-bg {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 180px; /* offset label width */
		right: 0;
		display: flex;
		justify-content: space-between;
		pointer-events: none;
		z-index: 1;
	}

	.grid-lines-bg .line {
		width: 1px;
		background-color: rgba(46, 51, 72, 0.5);
		height: 100%;
	}

	.timeline-rows {
		position: relative;
		z-index: 2;
		display: flex;
		flex-direction: column;
	}

	.timeline-row {
		display: flex;
		align-items: center;
		height: 2.25rem;
		border-bottom: 1px solid rgba(46, 51, 72, 0.3);
	}

	.timeline-row:last-child {
		border-bottom: none;
	}

	.span-label {
		width: 180px;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		justify-content: center;
		padding-right: 0.5rem;
		border-right: 1px solid var(--ql-border);
		height: 100%;
		background-color: rgba(26, 29, 46, 0.4);
		box-sizing: border-box;
	}

	.span-label-name {
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--ql-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.span-label-type {
		font-size: 0.6rem;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		font-weight: 600;
	}

	.bar-container {
		flex-grow: 1;
		position: relative;
		height: 100%;
		display: flex;
		align-items: center;
		padding: 0 0.5rem;
	}

	.duration-bar {
		position: absolute;
		height: 1rem;
		border-radius: 0.25rem;
		display: flex;
		align-items: center;
		padding: 0 0.25rem;
		min-width: 2rem;
		box-sizing: border-box;
		transition: all 0.2s;
	}

	.duration-bar:hover {
		filter: brightness(1.2);
		box-shadow: 0 0 8px rgba(255, 255, 255, 0.1);
	}

	.bar-label {
		font-size: 0.65rem;
		font-weight: 700;
		color: #0f1117;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.timeline-legend {
		display: flex;
		justify-content: space-between;
		padding-left: 180px; /* align with bars */
		font-size: 0.65rem;
		color: var(--ql-text-muted);
		font-weight: 500;
	}

	/* Color-coding bars */
	.span-type-llm {
		background-color: var(--ql-accent);
	}
	.span-type-retrieval {
		background-color: #3b82f6;
	}
	.span-type-embedding {
		background-color: #a855f7;
	}
	.span-type-tool {
		background-color: #f97316;
	}
	.span-type-agent {
		background-color: #ec4899;
	}
</style>
