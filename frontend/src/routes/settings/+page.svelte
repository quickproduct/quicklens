<script lang="ts">
	import { onMount } from 'svelte';
	import { apiPut } from '$lib/api/client';
	import { addToast } from '$stores/ui';
	import { Settings, Shield, Server, Database, RefreshCw, Key, Code, HelpCircle } from 'lucide-svelte';

	// Account Change Password Form State
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordLoading = $state(false);

	// General settings
	let ollamaHost = $state('http://localhost:11434');
	let testingOllama = $state(false);

	// API Keys State
	let openaiKey = $state('');
	let anthropicKey = $state('');
	let googleKey = $state('');

	// Active tab in settings
	let activeTab = $state('credentials'); // 'credentials', 'security', 'integration'
	// Integration sub-tab
	let integrationSubTab = $state('proxy'); // 'proxy', 'python', 'typescript'

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

<div class="settings-page animate-fade-in">
	<!-- Header -->
	<div class="page-header">
		<h2 class="page-title text-gradient">Settings & Setup</h2>
		<p class="page-subtitle">Configure proxy credentials, local server parameters, database settings, and view SDK integration guides.</p>
	</div>

	<!-- Tab Headers -->
	<div class="tabs-header mb-5 border-b border-ql-border flex gap-4">
		<button class="tab-btn pb-2 font-semibold text-sm transition-colors" class:active={activeTab === 'credentials'} onclick={() => (activeTab = 'credentials')}>
			<div class="flex items-center gap-1.5"><Key size={16} /> Credentials & Connectors</div>
		</button>
		<button class="tab-btn pb-2 font-semibold text-sm transition-colors" class:active={activeTab === 'integration'} onclick={() => (activeTab = 'integration')}>
			<div class="flex items-center gap-1.5"><Code size={16} /> Integration Guide</div>
		</button>
		<button class="tab-btn pb-2 font-semibold text-sm transition-colors" class:active={activeTab === 'security'} onclick={() => (activeTab = 'security')}>
			<div class="flex items-center gap-1.5"><Shield size={16} /> Security & Database</div>
		</button>
	</div>

	<!-- Main Settings Grid -->
	<div class="settings-content">
		{#if activeTab === 'credentials'}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-5">
				<!-- Ollama -->
				<div class="ql-card p-5 flex flex-col gap-4">
					<div class="card-header flex items-center gap-2 border-b border-ql-border pb-3">
						<Server size={18} class="text-ql-accent" />
						<h3 class="card-title text-md font-semibold m-0">Ollama Setup</h3>
					</div>
					<p class="text-xs text-ql-text-muted leading-relaxed m-0">Specify the location of your local Ollama engine. When configured, QuickLens will automatically sync available model tags and enable completions tracking.</p>

					<div class="form-group flex flex-col gap-1.5">
						<label for="ollama-host" class="form-label text-xs font-semibold text-ql-text-muted uppercase tracking-wider">Ollama Host URI</label>
						<div class="input-action-row flex gap-2">
							<input
								type="text"
								id="ollama-host"
								class="ql-input"
								placeholder="http://localhost:11434"
								bind:value={ollamaHost}
							/>
							<button class="ql-btn-secondary test-btn flex items-center gap-1.5 text-xs py-1" onclick={testOllamaConnection} disabled={testingOllama}>
								{#if testingOllama}
									<RefreshCw size={14} class="spin animate-spin" />
								{/if}
								Test Connection
							</button>
						</div>
					</div>
				</div>

				<!-- API Keys -->
				<div class="ql-card p-5 flex flex-col gap-4">
					<div class="card-header flex items-center gap-2 border-b border-ql-border pb-3">
						<Key size={18} class="text-ql-accent" />
						<h3 class="card-title text-md font-semibold m-0">Cloud Provider API Keys</h3>
					</div>
					<p class="text-xs text-ql-text-muted leading-relaxed m-0">Specify your API keys for transparent completions proxying. Keys are saved securely in your browser's local storage and used locally.</p>

					<div class="form-group flex flex-col gap-1.5">
						<label for="openai-key" class="form-label text-xs font-semibold text-ql-text-muted uppercase tracking-wider">OpenAI API Key</label>
						<input
							type="password"
							id="openai-key"
							class="ql-input text-sm"
							placeholder="sk-proj-..."
							bind:value={openaiKey}
						/>
					</div>

					<div class="form-group flex flex-col gap-1.5">
						<label for="anthropic-key" class="form-label text-xs font-semibold text-ql-text-muted uppercase tracking-wider">Anthropic API Key</label>
						<input
							type="password"
							id="anthropic-key"
							class="ql-input text-sm"
							placeholder="sk-ant-..."
							bind:value={anthropicKey}
						/>
					</div>

					<div class="form-group flex flex-col gap-1.5">
						<label for="google-key" class="form-label text-xs font-semibold text-ql-text-muted uppercase tracking-wider">Google Gemini API Key</label>
						<input
							type="password"
							id="google-key"
							class="ql-input text-sm"
							placeholder="AIzaSy..."
							bind:value={googleKey}
						/>
					</div>

					<button class="ql-btn-primary save-keys-btn self-start text-xs py-2 px-4 bg-ql-accent font-semibold" onclick={saveApiKeys}>
						Save API Credentials
					</button>
				</div>
			</div>
		{:else if activeTab === 'integration'}
			<div class="ql-card p-5 flex flex-col gap-4">
				<div class="card-header flex items-center gap-2 border-b border-ql-border pb-3">
					<Code size={18} class="text-ql-accent" />
					<h3 class="card-title text-md font-semibold m-0">Quick Start Integration Guide</h3>
				</div>
				<p class="text-xs text-ql-text-muted m-0">Connect your AI applications directly to the QuickLens collector. Choose an integration mode below.</p>

				<!-- Integration sub tabs -->
				<div class="flex gap-2 bg-ql-surface-2 p-1.5 rounded-lg self-start">
					<button class="px-3 py-1 text-xs rounded font-medium" class:bg-ql-surface={integrationSubTab === 'proxy'} class:text-ql-accent={integrationSubTab === 'proxy'} onclick={() => (integrationSubTab = 'proxy')}>Zero-Code Proxy</button>
					<button class="px-3 py-1 text-xs rounded font-medium" class:bg-ql-surface={integrationSubTab === 'python'} class:text-ql-accent={integrationSubTab === 'python'} onclick={() => (integrationSubTab = 'python')}>Python SDK</button>
					<button class="px-3 py-1 text-xs rounded font-medium" class:bg-ql-surface={integrationSubTab === 'typescript'} class:text-ql-accent={integrationSubTab === 'typescript'} onclick={() => (integrationSubTab = 'typescript')}>TypeScript SDK</button>
				</div>

				<div class="code-instructions bg-ql-surface-2 p-4 rounded-lg border border-ql-border flex flex-col gap-3">
					{#if integrationSubTab === 'proxy'}
						<div>
							<h4 class="text-sm font-semibold text-ql-text m-0 mb-1">Method 1: Zero-Code OpenAI Redirection</h4>
							<p class="text-xs text-ql-text-muted m-0">Route requests through the QuickLens gateway. It intercepts details (prompt, responses, costs, token counters) and passes them to OpenAI transparently.</p>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Option A: Environment Variable</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">export OPENAI_BASE_URL="http://localhost:8000/v1"</pre>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Option B: Python client code integration</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">import openai

openai.api_base = "http://localhost:8000/v1"
response = openai.chat.completions.create(
    model="gpt-4o",
    messages=[&#123;"role": "user", "content": "Hello world"&#125;]
)</pre>
						</div>
					{:else if integrationSubTab === 'python'}
						<div>
							<h4 class="text-sm font-semibold text-ql-text m-0 mb-1">Method 2: Python SDK Integration</h4>
							<p class="text-xs text-ql-text-muted m-0">Use the official Python library to trace application chains, tools, and retrievals with decorator methods.</p>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Install library</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">pip install quicklens-sdk</pre>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Code setup</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">from quicklens import QuickLensClient, quicklens_trace

client = QuickLensClient("http://localhost:8000")

# Instrument a function to record traces automatically
@quicklens_trace(client, name="document_rag_query")
def summarize_doc(text):
    # Your LLM/RAG call goes here
    pass</pre>
						</div>
					{:else if integrationSubTab === 'typescript'}
						<div>
							<h4 class="text-sm font-semibold text-ql-text m-0 mb-1">Method 3: Node/TypeScript SDK Integration</h4>
							<p class="text-xs text-ql-text-muted m-0">Send nested spans and traces using the standard JS SDK collector wrapper.</p>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Install package</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">npm install quicklens-sdk</pre>
						</div>

						<div class="flex flex-col gap-1">
							<span class="text-xs font-mono text-ql-text-muted">Code setup</span>
							<pre class="bg-ql-bg p-3 rounded font-mono text-xs overflow-x-auto text-ql-text border border-ql-border">import &#123; QuickLensClient &#125; from 'quicklens-sdk';

const ql = new QuickLensClient(&#123; baseUrl: 'http://localhost:8000' &#125;);

const trace = ql.createTrace(&#123; name: 'translation-service' &#125;);
trace.addSpan(&#123;
    name: 'claude-translate',
    model: 'claude-3-5-sonnet',
    tokens: &#123; input: 120, output: 250 &#125;
&#125;);
await trace.end();</pre>
						</div>
					{/if}
				</div>
			</div>
		{:else if activeTab === 'security'}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-5">
				<!-- Change Password -->
				<div class="ql-card p-5 flex flex-col gap-4">
					<div class="card-header flex items-center gap-2 border-b border-ql-border pb-3">
						<Shield size={18} class="text-ql-accent" />
						<h3 class="card-title text-md font-semibold m-0">Security Parameters</h3>
					</div>
					<p class="text-xs text-ql-text-muted leading-relaxed m-0">Update account credentials for the default seed administrator user.</p>

					<form onsubmit={handlePasswordChange} class="flex flex-col gap-3">
						<div class="form-group flex flex-col gap-1">
							<label for="current-pwd" class="form-label text-xs font-semibold text-ql-text-muted uppercase">Current Password</label>
							<input
								type="password"
								id="current-pwd"
								class="ql-input text-sm"
								bind:value={currentPassword}
								required
								autocomplete="current-password"
							/>
						</div>

						<div class="form-group flex flex-col gap-1">
							<label for="new-pwd" class="form-label text-xs font-semibold text-ql-text-muted uppercase">New Password</label>
							<input
								type="password"
								id="new-pwd"
								class="ql-input text-sm"
								bind:value={newPassword}
								required
								autocomplete="new-password"
							/>
						</div>

						<div class="form-group flex flex-col gap-1">
							<label for="confirm-pwd" class="form-label text-xs font-semibold text-ql-text-muted uppercase">Confirm New Password</label>
							<input
								type="password"
								id="confirm-pwd"
								class="ql-input text-sm"
								bind:value={confirmPassword}
								required
								autocomplete="new-password"
							/>
						</div>

						<button type="submit" class="ql-btn-primary self-start text-xs py-2 px-4 bg-ql-accent font-semibold" disabled={passwordLoading}>
							{passwordLoading ? 'Updating...' : 'Change Password'}
						</button>
					</form>
				</div>

				<!-- Storage -->
				<div class="ql-card p-5 flex flex-col gap-4">
					<div class="card-header flex items-center gap-2 border-b border-ql-border pb-3">
						<Database size={18} class="text-ql-accent" />
						<h3 class="card-title text-md font-semibold m-0">Database & Retention</h3>
					</div>
					<p class="text-xs text-ql-text-muted leading-relaxed m-0">QuickLens uses SQLite for self-hosted, CGO-free, high-performance tracing metadata storage.</p>

					<div class="flex flex-col gap-2.5 mt-2">
						<div class="info-row flex justify-between border-b border-ql-border pb-2 text-sm">
							<span class="text-ql-text-muted">Data Retention Period</span>
							<span class="text-ql-text font-semibold">30 Days</span>
						</div>
						<div class="info-row flex justify-between border-b border-ql-border pb-2 text-sm">
							<span class="text-ql-text-muted">SQLite Journaling Mode</span>
							<span class="text-ql-text font-mono font-semibold">WAL (Write-Ahead Log)</span>
						</div>
						<div class="info-row flex justify-between pb-1 text-sm">
							<span class="text-ql-text-muted">Autovacuum Mode</span>
							<span class="text-ql-text font-semibold">Incremental vacuum</span>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.page-title {
		font-size: 1.5rem;
		font-weight: 800;
		letter-spacing: -0.025em;
		margin: 0;
	}

	.tab-btn {
		background: none;
		border: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		border-bottom: 2px solid transparent;
	}

	.tab-btn:hover {
		color: var(--ql-text);
	}

	.tab-btn.active {
		color: var(--ql-accent);
		border-bottom-color: var(--ql-accent);
	}
</style>
