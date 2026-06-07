<script lang="ts">
	import { onMount } from 'svelte';
	import { apiPut } from '$lib/api/client';
	import { addToast } from '$stores/ui';
	import { Settings, Shield, Server, Coins, Database, RefreshCw, Key } from 'lucide-svelte';

	// Account Change Password Form State
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordLoading = $state(false);

	// General settings (saved locally in browser or loaded/mocked)
	let ollamaHost = $state('http://localhost:11434');
	let testingOllama = $state(false);

	// API Keys State
	let openaiKey = $state('');
	let anthropicKey = $state('');
	let googleKey = $state('');

	async function handlePasswordChange(e: SubmitEvent) {
		e.preventDefault();
		if (newPassword !== confirmPassword) {
			addToast('Passwords do not match.', 'error');
			return;
		}

		passwordLoading = true;
		try {
			await apiPut('/api/v1/auth/password', {
				current_password: currentPassword,
				new_password: newPassword
			});
			addToast('Password updated successfully.', 'success');
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (err) {
			console.error(err);
			addToast('Failed to change password. Make sure current password is correct.', 'error');
		} finally {
			passwordLoading = false;
		}
	}

	async function testOllamaConnection() {
		testingOllama = true;
		addToast('Testing connection to Ollama...', 'info');
		try {
			// Proxy tag check via backend proxy tags endpoint (which checks Ollama)
			const res = await fetch('/ollama/api/tags');
			if (res.ok) {
				addToast('Connection to Ollama successful!', 'success');
			} else {
				addToast('Ollama returned an error status.', 'warning');
			}
		} catch (err) {
			console.error(err);
			addToast('Failed to connect to Ollama. Make sure it is running locally.', 'error');
		} finally {
			testingOllama = false;
		}
	}

	function saveApiKeys() {
		localStorage.setItem('openai_key', openaiKey);
		localStorage.setItem('anthropic_key', anthropicKey);
		localStorage.setItem('google_key', googleKey);
		addToast('API keys saved locally in browser!', 'success');
	}

	onMount(() => {
		openaiKey = localStorage.getItem('openai_key') || '';
		anthropicKey = localStorage.getItem('anthropic_key') || '';
		googleKey = localStorage.getItem('google_key') || '';
	});
</script>

<div class="settings-page">
	<!-- Header -->
	<div class="page-header">
		<h2 class="page-title">Settings</h2>
		<p class="page-subtitle">Configure cloud integrations, local model connections, model cost mappings, and security parameters.</p>
	</div>

	<!-- Main Settings Grid -->
	<div class="settings-layout">
		<!-- Left side: Setup and Connections -->
		<div class="settings-column">
			<!-- Ollama -->
			<div class="ql-card settings-card">
				<div class="card-header">
					<Server size={18} class="card-icon" />
					<h3 class="card-title">Ollama Configuration</h3>
				</div>
				<p class="card-desc">Configure the local Ollama instance path for automatic model tag discovery and transparent request proxying.</p>

				<div class="form-group">
					<label for="ollama-host" class="form-label">Ollama Host Address</label>
					<div class="input-action-row">
						<input
							type="text"
							id="ollama-host"
							class="ql-input"
							placeholder="http://localhost:11434"
							bind:value={ollamaHost}
						/>
						<button class="ql-btn-secondary test-btn" onclick={testOllamaConnection} disabled={testingOllama}>
							{#if testingOllama}
								<RefreshCw size={14} class="spin" />
							{/if}
							Test Connection
						</button>
					</div>
				</div>
			</div>

			<!-- API Keys -->
			<div class="ql-card settings-card">
				<div class="card-header">
					<Key size={18} class="card-icon" />
					<h3 class="card-title">Cloud Provider API Keys</h3>
				</div>
				<p class="card-desc">Specify credentials to query external services. Keys are used locally for transparent API completions forwarding.</p>

				<div class="form-group">
					<label for="openai-key" class="form-label">OpenAI API Key</label>
					<input
						type="password"
						id="openai-key"
						class="ql-input"
						placeholder="sk-proj-..."
						bind:value={openaiKey}
					/>
				</div>

				<div class="form-group">
					<label for="anthropic-key" class="form-label">Anthropic API Key</label>
					<input
						type="password"
						id="anthropic-key"
						class="ql-input"
						placeholder="sk-ant-..."
						bind:value={anthropicKey}
					/>
				</div>

				<div class="form-group">
					<label for="google-key" class="form-label">Google Gemini API Key</label>
					<input
						type="password"
						id="google-key"
						class="ql-input"
						placeholder="AIzaSy..."
						bind:value={googleKey}
					/>
				</div>

				<button class="ql-btn-primary save-keys-btn" onclick={saveApiKeys}>
					Save Credentials
				</button>
			</div>
		</div>

		<!-- Right side: Security & Data -->
		<div class="settings-column">
			<!-- Change Password -->
			<div class="ql-card settings-card">
				<div class="card-header">
					<Shield size={18} class="card-icon" />
					<h3 class="card-title">Security & Account</h3>
				</div>
				<p class="card-desc">Update password settings for the default seed admin or register credentials.</p>

				<form onsubmit={handlePasswordChange} class="settings-form">
					<div class="form-group">
						<label for="current-pwd" class="form-label">Current Password</label>
						<input
							type="password"
							id="current-pwd"
							class="ql-input"
							bind:value={currentPassword}
							required
							autocomplete="current-password"
						/>
					</div>

					<div class="form-group">
						<label for="new-pwd" class="form-label">New Password</label>
						<input
							type="password"
							id="new-pwd"
							class="ql-input"
							bind:value={newPassword}
							required
							autocomplete="new-password"
						/>
					</div>

					<div class="form-group">
						<label for="confirm-pwd" class="form-label">Confirm New Password</label>
						<input
							type="password"
							id="confirm-pwd"
							class="ql-input"
							bind:value={confirmPassword}
							required
							autocomplete="new-password"
						/>
					</div>

					<button type="submit" class="ql-btn-primary save-keys-btn" disabled={passwordLoading}>
						{passwordLoading ? 'Updating...' : 'Change Password'}
					</button>
				</form>
			</div>

			<!-- Retention and Storage -->
			<div class="ql-card settings-card">
				<div class="card-header">
					<Database size={18} class="card-icon" />
					<h3 class="card-title">Database Storage & Retention</h3>
				</div>
				<p class="card-desc">Current database is SQLite (`quicklens.db`). Storage cleanups are evaluated every hour by the Janitor background worker.</p>
				<div class="info-row">
					<span class="info-lbl">Data Retention Days</span>
					<span class="info-val">30 Days (Default)</span>
				</div>
				<div class="info-row">
					<span class="info-lbl">Database Journaling Mode</span>
					<span class="info-val font-mono">WAL (Write-Ahead Logging)</span>
				</div>
				<div class="info-row">
					<span class="info-lbl">Integrations Mode</span>
					<span class="info-val">Transparent API Proxy Gateway</span>
				</div>
			</div>
		</div>
	</div>
</div>


<style>
	.settings-page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.page-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--ql-text);
		letter-spacing: -0.025em;
	}

	.page-subtitle {
		font-size: 0.875rem;
		color: var(--ql-text-muted);
	}

	.settings-layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	@media (max-width: 768px) {
		.settings-layout {
			grid-template-columns: 1fr;
		}
	}

	.settings-column {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.settings-card {
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		border-bottom: 1px solid var(--ql-border);
		padding-bottom: 0.5rem;
	}

	.card-icon {
		color: var(--ql-accent);
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.card-desc {
		font-size: 0.825rem;
		color: var(--ql-text-muted);
		line-height: 1.4;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.form-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ql-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.input-action-row {
		display: flex;
		gap: 0.5rem;
	}

	.input-action-row input {
		flex-grow: 1;
	}

	.test-btn {
		font-size: 0.75rem;
		padding: 0.5rem 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
		white-space: nowrap;
	}

	.save-keys-btn {
		align-self: flex-start;
		font-size: 0.85rem;
		padding: 0.5rem 1rem;
		margin-top: 0.5rem;
		background-color: var(--ql-accent);
	}

	.settings-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.info-row {
		display: flex;
		justify-content: space-between;
		font-size: 0.85rem;
		border-bottom: 1px solid rgba(46, 51, 72, 0.4);
		padding-bottom: 0.5rem;
	}

	.info-row:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	.info-lbl {
		color: var(--ql-text-muted);
	}

	.info-val {
		color: var(--ql-text);
		font-weight: 500;
	}

	.spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
