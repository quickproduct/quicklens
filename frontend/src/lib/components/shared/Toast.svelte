<script lang="ts">
	import { toasts, removeToast } from '$stores/ui';
	import { X } from 'lucide-svelte';

	let activeToasts = $state([] as any[]);

	$effect(() => {
		const unsub = toasts.subscribe((t) => {
			activeToasts = t;
		});
		return unsub;
	});
</script>

<div class="toast-container">
	{#each activeToasts as toast (toast.id)}
		<div class="toast toast-{toast.type}" role="alert">
			<div class="toast-content">{toast.message}</div>
			<button class="toast-close" onclick={() => removeToast(toast.id)}>
				<X size={16} />
			</button>
		</div>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		bottom: 1.5rem;
		right: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		z-index: 9999;
		max-width: 24rem;
		width: calc(100% - 3rem);
	}

	.toast {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem;
		border-radius: 0.5rem;
		background: rgba(26, 29, 46, 0.9);
		backdrop-filter: blur(12px);
		border: 1px solid var(--ql-border);
		box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.5);
		animation: slide-in 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
	}

	.toast-content {
		font-size: 0.875rem;
		color: var(--ql-text);
		margin-right: 1rem;
		word-break: break-word;
	}

	.toast-close {
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

	.toast-close:hover {
		color: var(--ql-text);
		background-color: var(--ql-surface-2);
	}

	.toast-success {
		border-left: 4px solid var(--ql-accent);
	}

	.toast-error {
		border-left: 4px solid var(--ql-danger, #ef4444);
	}

	.toast-warning {
		border-left: 4px solid var(--ql-warning, #f59e0b);
	}

	.toast-info {
		border-left: 4px solid var(--ql-info, #3b82f6);
	}

	@keyframes slide-in {
		from {
			transform: translateY(1rem);
			opacity: 0;
		}
		to {
			transform: translateY(0);
			opacity: 1;
		}
	}
</style>
