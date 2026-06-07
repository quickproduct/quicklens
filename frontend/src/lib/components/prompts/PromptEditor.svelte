<script lang="ts">
	import { X } from 'lucide-svelte';

	let {
		isOpen = false,
		prompt = null as any,
		onClose = () => {},
		onSave = (data: any) => {}
	}: {
		isOpen: boolean;
		prompt?: any;
		onClose: () => void;
		onSave: (data: any) => void;
	} = $props();

	let name = $state('');
	let content = $state('');
	let modelID = $state('');
	let tagsInput = $state('');

	$effect(() => {
		if (isOpen) {
			if (prompt) {
				name = prompt.name || '';
				content = prompt.content || '';
				modelID = prompt.model_id || '';
				tagsInput = prompt.tags ? prompt.tags.join(', ') : '';
			} else {
				name = '';
				content = '';
				modelID = '';
				tagsInput = '';
			}
		}
	});

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!name || !content) return;

		const tags = tagsInput
			.split(',')
			.map((t) => t.trim())
			.filter((t) => t.length > 0);

		onSave({
			name,
			content,
			model_id: modelID,
			tags
		});
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" onclick={onClose} role="presentation">
		<div class="modal-container ql-card" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
			<!-- Header -->
			<div class="modal-header">
				<h3 class="modal-title">{prompt ? 'Edit / New Version' : 'Create New Prompt'}</h3>
				<button class="close-btn" onclick={onClose}>
					<X size={18} />
				</button>
			</div>

			<!-- Form -->
			<form onsubmit={handleSubmit} class="modal-form">
				<div class="form-group">
					<label for="prompt-name" class="form-label">Prompt Name</label>
					<input
						type="text"
						id="prompt-name"
						class="ql-input"
						placeholder="e.g. system_chat_agent"
						bind:value={name}
						required
						disabled={!!prompt}
					/>
					{#if prompt}
						<span class="input-tip">Prompt name cannot be modified. Saving will create a new version of this prompt.</span>
					{/if}
				</div>

				<div class="form-row">
					<div class="form-group">
						<label for="prompt-model" class="form-label">Associated Model (Optional)</label>
						<input
							type="text"
							id="prompt-model"
							class="ql-input"
							placeholder="e.g. gpt-4o"
							bind:value={modelID}
						/>
					</div>

					<div class="form-group">
						<label for="prompt-tags" class="form-label">Tags (Comma-separated)</label>
						<input
							type="text"
							id="prompt-tags"
							class="ql-input"
							placeholder="e.g. prod, chat, rag"
							bind:value={tagsInput}
						/>
					</div>
				</div>

				<div class="form-group">
					<label for="prompt-content" class="form-label">Content template</label>
					<textarea
						id="prompt-content"
						class="ql-input content-textarea"
						placeholder="You are a helpful assistant. User context: {{user_context}}"
						bind:value={content}
						required
					></textarea>
				</div>

				<div class="form-actions">
					<button type="button" class="ql-btn-secondary" onclick={onClose}>
						Cancel
					</button>
					<button type="submit" class="ql-btn-primary">
						{prompt ? 'Save New Version' : 'Create Prompt'}
					</button>
				</div>
			</form>
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
		width: 100%;
		max-width: 600px;
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
		margin-bottom: 1.25rem;
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

	.modal-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--ql-text-muted);
	}

	.input-tip {
		font-size: 0.75rem;
		color: var(--ql-accent);
	}

	.content-textarea {
		min-height: 12rem;
		font-family: var(--font-mono, 'JetBrains Mono', monospace);
		font-size: 0.85rem;
		line-height: 1.4;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		border-top: 1px solid var(--ql-border);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	@keyframes scale-up {
		from {
			transform: scale(0.95);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}
</style>
