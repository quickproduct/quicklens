<script lang="ts">
	let {
		title = 'No data available',
		description = 'There is nothing to display here yet.',
		icon: IconComponent = null,
		actionText = '',
		onAction = null
	}: {
		title?: string;
		description?: string;
		icon?: any;
		actionText?: string;
		onAction?: (() => void) | null;
	} = $props();
</script>

<div class="empty-state">
	<div class="empty-icon-wrapper">
		{#if IconComponent}
			<IconComponent size={32} class="empty-icon" />
		{/if}
	</div>
	<h3 class="empty-title">{title}</h3>
	<p class="empty-desc">{description}</p>
	{#if actionText && onAction}
		<button class="ql-btn-primary empty-btn" onclick={onAction}>
			{actionText}
		</button>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 3rem 1.5rem;
		background: rgba(26, 29, 46, 0.4);
		border: 1px dashed var(--ql-border);
		border-radius: 0.5rem;
		margin: 1rem 0;
	}

	.empty-icon-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 4rem;
		height: 4rem;
		border-radius: 9999px;
		background-color: var(--ql-surface-2);
		margin-bottom: 1rem;
		color: var(--ql-text-muted);
	}

	.empty-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--ql-text);
		margin-bottom: 0.5rem;
	}

	.empty-desc {
		font-size: 0.875rem;
		color: var(--ql-text-muted);
		max-width: 20rem;
		margin-bottom: 1.5rem;
	}

	.empty-btn {
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
	}
</style>
