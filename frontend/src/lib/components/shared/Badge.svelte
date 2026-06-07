<script lang="ts">
	let { text = '', variant = 'neutral' }: {
		text: string;
		variant?: 'success' | 'danger' | 'warning' | 'info' | 'neutral';
	} = $props();
</script>

<span class="ql-badge ql-badge-{variant}">
	{text}
</span>

<style>
	.ql-badge {
		display: inline-flex;
		align-items: center;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
		line-height: 1rem;
		border: 1px solid transparent;
	}

	.ql-badge-neutral {
		background-color: var(--ql-surface-2);
		color: var(--ql-text);
		border-color: var(--ql-border);
	}

	.ql-badge-success {
		background-color: rgba(34, 197, 94, 0.1);
		color: #4ade80;
		border-color: rgba(34, 197, 94, 0.2);
	}

	.ql-badge-danger {
		background-color: rgba(239, 68, 68, 0.1);
		color: #fca5a5;
		border-color: rgba(239, 68, 68, 0.2);
	}

	.ql-badge-warning {
		background-color: rgba(245, 158, 11, 0.1);
		color: #fcd34d;
		border-color: rgba(245, 158, 11, 0.2);
	}

	.ql-badge-info {
		background-color: rgba(59, 130, 246, 0.1);
		color: #93c5fd;
		border-color: rgba(59, 130, 246, 0.2);
	}
</style>
