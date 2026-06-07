<script lang="ts">
	let {
		title = '',
		value = '',
		icon: IconComponent = null,
		trend = '',
		trendUp = true
	}: {
		title: string;
		value: string | number;
		icon?: any;
		trend?: string;
		trendUp?: boolean;
	} = $props();
</script>

<div class="ql-card stat-card">
	<div class="stat-main">
		<div class="stat-info">
			<span class="stat-title">{title}</span>
			<span class="stat-value">{value}</span>
		</div>
		<div class="stat-icon-wrapper">
			{#if IconComponent}
				<IconComponent size={24} class="stat-icon" />
			{/if}
		</div>
	</div>
	{#if trend}
		<div class="stat-footer">
			<span class="stat-trend" class:trend-up={trendUp} class:trend-down={!trendUp}>
				{trend}
			</span>
			<span class="stat-trend-label">vs yesterday</span>
		</div>
	{/if}
</div>

<style>
	.stat-card {
		display: flex;
		flex-direction: column;
		padding: 1.25rem;
		position: relative;
		overflow: hidden;
	}

	.stat-card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: linear-gradient(135deg, rgba(16, 185, 129, 0.03) 0%, transparent 100%);
		pointer-events: none;
	}

	.stat-main {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 0.5rem;
	}

	.stat-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.stat-title {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--ql-text-muted);
	}

	.stat-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--ql-text);
		letter-spacing: -0.025em;
	}

	.stat-icon-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2.75rem;
		height: 2.75rem;
		border-radius: 0.5rem;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		color: var(--ql-accent);
	}

	.stat-footer {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
	}

	.stat-trend {
		font-weight: 600;
	}

	.stat-trend.trend-up {
		color: var(--ql-success, #22c55e);
	}

	.stat-trend.trend-down {
		color: var(--ql-danger, #ef4444);
	}

	.stat-trend-label {
		color: var(--ql-text-muted);
	}
</style>
