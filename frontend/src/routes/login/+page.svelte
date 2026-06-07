<script lang="ts">
	import { login } from '$stores/auth';
	import { addToast } from '$stores/ui';
	import { Cpu } from 'lucide-svelte';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMsg = $state('');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!email || !password) return;

		loading = true;
		errorMsg = '';

		try {
			const success = await login(email, password);
			if (success) {
				addToast('Successfully logged in!', 'success');
			} else {
				errorMsg = 'Invalid email or password';
			}
		} catch (err) {
			console.error(err);
			errorMsg = 'Authentication failed. Please try again.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="login-page">
	<div class="background-radial-glow"></div>

	<div class="login-card-container">
		<div class="ql-card login-card">
			<!-- Logo Header -->
			<div class="login-header">
				<div class="login-logo-wrapper">
					<svg width="40" height="40" viewBox="0 0 28 28" fill="none">
						<circle cx="14" cy="14" r="12" stroke="#10b981" stroke-width="2" fill="none" />
						<circle cx="14" cy="14" r="6" fill="#10b981" />
					</svg>
				</div>
				<h2 class="login-title">Welcome to QuickLens</h2>
				<p class="login-subtitle">Self-hosted LLM Observability & Monitoring</p>
			</div>

			<!-- Error display -->
			{#if errorMsg}
				<div class="error-banner">
					{errorMsg}
				</div>
			{/if}

			<!-- Form -->
			<form onsubmit={handleSubmit} class="login-form">
				<div class="form-group">
					<label for="email-input" class="form-label">Email Address</label>
					<input
						type="email"
						id="email-input"
						class="ql-input"
						placeholder="admin@quicklens.local"
						bind:value={email}
						required
						autocomplete="username"
					/>
				</div>

				<div class="form-group">
					<label for="password-input" class="form-label">Password</label>
					<input
						type="password"
						id="password-input"
						class="ql-input"
						placeholder="••••••••"
						bind:value={password}
						required
						autocomplete="current-password"
					/>
				</div>

				<button type="submit" class="ql-btn-primary login-submit-btn" disabled={loading}>
					{#if loading}
						<div class="spinner"></div> Authenticating...
					{:else}
						Sign In
					{/if}
				</button>
			</form>
		</div>

		<!-- Footer -->
		<div class="login-footer">
			<span>QuickLens is open source under MIT License</span>
		</div>
	</div>
</div>

<style>
	.login-page {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100vh;
		background-color: var(--ql-bg);
		overflow: hidden;
	}

	.background-radial-glow {
		position: absolute;
		top: -20%;
		left: -20%;
		width: 140%;
		height: 140%;
		background: radial-gradient(circle at center, rgba(16, 185, 129, 0.08) 0%, transparent 60%);
		pointer-events: none;
		z-index: 1;
	}

	.login-card-container {
		position: relative;
		z-index: 10;
		width: 100%;
		max-width: 420px;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.login-card {
		padding: 2.25rem 2rem;
		background: rgba(26, 29, 46, 0.75);
		backdrop-filter: blur(20px);
		border-color: var(--ql-border);
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.6);
	}

	.login-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		margin-bottom: 2rem;
	}

	.login-logo-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 1rem;
		filter: drop-shadow(0 0 10px rgba(16, 185, 129, 0.3));
	}

	.login-title {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--ql-text);
		margin-bottom: 0.35rem;
		letter-spacing: -0.025em;
	}

	.login-subtitle {
		font-size: 0.85rem;
		color: var(--ql-text-muted);
	}

	.error-banner {
		padding: 0.75rem 1rem;
		border-radius: 0.375rem;
		background-color: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.2);
		color: #fca5a5;
		font-size: 0.825rem;
		text-align: center;
		margin-bottom: 1.25rem;
	}

	.login-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
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

	.login-submit-btn {
		width: 100%;
		padding: 0.75rem;
		font-weight: 600;
		font-size: 0.9rem;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid rgba(15, 17, 23, 0.2);
		border-top-color: var(--ql-text);
		border-radius: 9999px;
		animation: spin 0.8s linear infinite;
	}

	.login-footer {
		text-align: center;
		font-size: 0.75rem;
		color: var(--ql-text-muted);
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
