<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getAlerts, acknowledgeAlert, getRules, createRule, deleteRule, updateRule, type Alert, type AlertRule } from '$api/alerts';
	import { createIncident } from '$api/operations';
	import { addToast } from '$stores/ui';
	import Badge from '$components/shared/Badge.svelte';
	import EmptyState from '$components/shared/EmptyState.svelte';
	import { createWebSocket } from '$lib/websocket/client';
	import { Bell, ShieldAlert, Plus, Trash2, Check, X, ShieldCheck, ExternalLink } from 'lucide-svelte';

	let alerts = $state([] as Alert[]);
	let rules = $state([] as AlertRule[]);
	let loading = $state(true);
	let wsClient: any = null;

	// Modal State
	let isModalOpen = $state(false);
	let metricType = $state('cost');
	let threshold = $state(0.0);
	let operator = $state('>');
	let windowSeconds = $state(300);
	let ruleEnabled = $state(true);
	let submitting = $state(false);

	async function loadAlertsAndRules() {
		try {
			const [alertsRes, rulesRes] = await Promise.all([getAlerts(), getRules()]);
			alerts = alertsRes || [];
			rules = rulesRes || [];
		} catch (err) {
			console.error(err);
			addToast('Failed to load alert rules or notifications.', 'error');
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadAlertsAndRules();

		// Subscribe to WebSocket alerts channel for live alert notifications
		wsClient = createWebSocket('alerts', (msg: any) => {
			loadAlertsAndRules();
		});
	});

	onDestroy(() => {
		if (wsClient) wsClient.close();
	});

	async function handleAcknowledge(id: string) {
		try {
			await acknowledgeAlert(id);
			addToast('Alert acknowledged.', 'success');
			loadAlertsAndRules();
		} catch (err) {
			console.error(err);
			addToast('Failed to acknowledge alert.', 'error');
		}
	}

	async function handleDeclareIncident(alert: Alert) {
		try {
			await createIncident({
				title: alert.message,
				severity: alert.severity,
				alert_id: alert.id,
				model_id: alert.model_id,
				service_id: alert.service_id,
				owner_id: alert.owner_id,
				runbook_url: alert.runbook_url,
				summary: 'Declared from Alert Center.'
			});
			addToast('Incident declared.', 'success');
			loadAlertsAndRules();
		} catch (err) {
			console.error(err);
			addToast('Failed to declare incident.', 'error');
		}
	}

	async function handleDeleteRule(id: string) {
		if (!confirm('Are you sure you want to delete this alert rule?')) return;
		try {
			await deleteRule(id);
			addToast('Alert rule deleted.', 'success');
			loadAlertsAndRules();
		} catch (err) {
			console.error(err);
			addToast('Failed to delete alert rule.', 'error');
		}
	}

	async function handleToggleRule(rule: AlertRule) {
		try {
			await updateRule(rule.id, { enabled: !rule.enabled });
			addToast(`Rule ${!rule.enabled ? 'enabled' : 'disabled'}.`, 'success');
			loadAlertsAndRules();
		} catch (err) {
			console.error(err);
			addToast('Failed to toggle rule state.', 'error');
		}
	}

	async function handleCreateRule(e: SubmitEvent) {
		e.preventDefault();
		submitting = true;
		try {
			await createRule({
				metric_type: metricType,
				operator,
				threshold,
				window_seconds: windowSeconds,
				enabled: ruleEnabled
			});
			addToast('Alert rule created successfully!', 'success');
			isModalOpen = false;
			loadAlertsAndRules();
		} catch (err) {
			console.error(err);
			addToast('Failed to create alert rule.', 'error');
		} finally {
			submitting = false;
		}
	}

	function formatDateTime(isoString: string): string {
		try {
			const d = new Date(isoString);
			return d.toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return isoString;
		}
	}

	function formatOperator(op: string): string {
		switch (op) {
			case 'gt': return '>';
			case 'lt': return '<';
			case 'gte': return '>=';
			case 'lte': return '<=';
			default: return op;
		}
	}
</script>

<div class="alerts-page">
	<!-- Header -->
	<div class="page-header-row">
		<div class="header-titles">
			<h2 class="page-title">Alert Center</h2>
			<p class="page-subtitle">Configure threshold rules for cost, latency, and error rates, and respond to incidents.</p>
		</div>

		<button class="ql-btn-primary add-btn" onclick={() => (isModalOpen = true)}>
			<Plus size={16} /> Create Rule
		</button>
	</div>

	<!-- Content split layout -->
	{#if loading}
		<div class="loading-container">
			<div class="spinner"></div>
			<span>Loading alerts...</span>
		</div>
	{:else}
		<div class="alerts-grid-layout">
			<!-- Left side: Rules -->
			<div class="rules-section">
				<div class="ql-card section-card">
					<h3 class="card-title">Alert Rules</h3>
					{#if rules.length === 0}
						<div class="empty-list-state">No alert rules configured yet.</div>
					{:else}
						<div class="table-wrapper">
							<table class="alerts-table">
								<thead>
									<tr>
										<th>Metric</th>
										<th>Condition</th>
										<th>Time Window</th>
										<th>State</th>
										<th class="text-center">Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each rules as rule (rule.id)}
										<tr>
											<td class="font-semibold text-uppercase">{rule.metric_type}</td>
											<td class="font-mono">{formatOperator(rule.operator)} {rule.threshold}</td>
											<td>{rule.window_seconds / 60} min ({rule.window_seconds}s)</td>
											<td>
												<button
													class="toggle-switch"
													class:active={rule.enabled}
													onclick={() => handleToggleRule(rule)}
													title={rule.enabled ? 'Click to Disable' : 'Click to Enable'}
												>
													<span class="switch-handle"></span>
												</button>
											</td>
											<td class="text-center">
												<button class="delete-btn" onclick={() => handleDeleteRule(rule.id)} title="Delete Rule">
													<Trash2 size={15} />
												</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			</div>

			<!-- Right side: Notifications -->
			<div class="notifications-section">
				<div class="ql-card section-card">
					<h3 class="card-title">Triggered Alerts</h3>
					{#if alerts.length === 0}
						<div class="empty-alerts">
							<ShieldCheck size={36} class="all-clear-icon" />
							<h4>All Clear</h4>
							<p>No alerts triggered. System operating within nominal parameters.</p>
						</div>
					{:else}
						<div class="alerts-scroller">
							{#each alerts as alert (alert.id)}
								<div class="alert-item" class:acked={alert.acknowledged}>
									<div class="alert-header">
										<Badge
											text={alert.severity}
											variant={alert.severity === 'critical' ? 'danger' : alert.severity === 'warning' ? 'warning' : 'info'}
										/>
										<span class="alert-time text-xs text-muted">{formatDateTime(alert.created_at)}</span>
									</div>
									<p class="alert-msg">{alert.message}</p>
									<div class="alert-context-row">
										<span>Status: {alert.status || (alert.acknowledged ? 'acknowledged' : 'open')}</span>
										<span>Owner: {alert.owner_id || 'Unassigned'}</span>
										<span>Service: {alert.service_id || alert.model_id || 'Unmapped'}</span>
									</div>
									<div class="alert-actions">
										{#if !alert.acknowledged}
											<button class="ql-btn-secondary ack-btn" onclick={() => handleAcknowledge(alert.id)}>
												<Check size={14} /> Acknowledge
											</button>
										{:else}
											<span class="status-label-acked">Acknowledged</span>
										{/if}
										{#if !alert.incident_id}
											<button class="ql-btn-secondary ack-btn" onclick={() => handleDeclareIncident(alert)}>
												<ExternalLink size={14} /> Declare Incident
											</button>
										{:else}
											<a class="incident-link" href="/incidents">Incident linked</a>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Create Rule Modal -->
	{#if isModalOpen}
		<div class="modal-backdrop" onclick={() => (isModalOpen = false)} role="presentation">
			<div class="modal-container ql-card" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
				<div class="modal-header">
					<h3 class="modal-title">Create Alert Rule</h3>
					<button class="close-btn" onclick={() => (isModalOpen = false)}>
						<X size={18} />
					</button>
				</div>

				<form onsubmit={handleCreateRule} class="modal-form">
					<div class="form-group">
						<label for="rule-metric" class="form-label">Metric Type</label>
						<select id="rule-metric" class="ql-input" bind:value={metricType}>
							<option value="cost">Total Cost ($)</option>
							<option value="error_rate">Error Rate (%)</option>
							<option value="latency">Average Latency (ms)</option>
							<option value="tokens">Total Token Throughput</option>
						</select>
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="rule-op" class="form-label">Operator</label>
							<select id="rule-op" class="ql-input" bind:value={operator}>
								<option value=">">&gt; (Greater Than)</option>
								<option value=">=">&gt;= (Greater Than or Equal)</option>
								<option value="<">&lt; (Less Than)</option>
								<option value="<=">&lt;= (Less Than or Equal)</option>
							</select>
						</div>

						<div class="form-group">
							<label for="rule-thresh" class="form-label">Threshold Value</label>
							<input
								type="number"
								step="0.0001"
								id="rule-thresh"
								class="ql-input"
								placeholder="e.g. 0.05 or 1000"
								bind:value={threshold}
								required
							/>
						</div>
					</div>

					<div class="form-group">
						<label for="rule-window" class="form-label">Evaluation Time Window (Seconds)</label>
						<input
							type="number"
							id="rule-window"
							class="ql-input"
							placeholder="e.g. 300 for 5 minutes"
							bind:value={windowSeconds}
							required
						/>
					</div>

					<div class="form-actions">
						<button type="button" class="ql-btn-secondary" onclick={() => (isModalOpen = false)}>
							Cancel
						</button>
						<button type="submit" class="ql-btn-primary" disabled={submitting}>
							{submitting ? 'Creating...' : 'Create Rule'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>

<style>
	.alerts-page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-titles {
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

	.add-btn {
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		display: flex;
		align-items: center;
		gap: 0.35rem;
		background-color: var(--ql-accent);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
		color: var(--ql-text-muted);
	}

	.spinner {
		width: 2rem;
		height: 2rem;
		border: 3px solid rgba(16, 185, 129, 0.1);
		border-top-color: var(--ql-accent);
		border-radius: 9999px;
		animation: spin 0.8s linear infinite;
	}

	.alerts-grid-layout {
		display: grid;
		grid-template-columns: 1.2fr 1fr;
		gap: 1.5rem;
		min-height: 0;
	}

	@media (max-width: 900px) {
		.alerts-grid-layout {
			grid-template-columns: 1fr;
		}
	}

	.section-card {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		height: 100%;
		box-sizing: border-box;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
	}

	.empty-list-state {
		text-align: center;
		padding: 3rem 1.5rem;
		color: var(--ql-text-muted);
		font-size: 0.875rem;
		border: 1px dashed var(--ql-border);
		border-radius: 0.375rem;
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.alerts-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.alerts-table th,
	.alerts-table td {
		padding: 0.625rem 0.75rem;
		text-align: left;
		border-bottom: 1px solid var(--ql-border);
	}

	.alerts-table th {
		color: var(--ql-text-muted);
		font-weight: 600;
		text-transform: uppercase;
		font-size: 0.7rem;
		letter-spacing: 0.05em;
		background-color: rgba(26, 29, 46, 0.4);
	}

	.alerts-table tr:last-child td {
		border-bottom: none;
	}

	.toggle-switch {
		position: relative;
		width: 2.5rem;
		height: 1.25rem;
		border-radius: 9999px;
		background-color: var(--ql-surface-2);
		border: 1px solid var(--ql-border);
		cursor: pointer;
		display: flex;
		align-items: center;
		transition: background-color 0.2s;
		padding: 0;
	}

	.toggle-switch.active {
		background-color: var(--ql-accent);
		border-color: var(--ql-accent);
	}

	.switch-handle {
		width: 1rem;
		height: 1rem;
		border-radius: 9999px;
		background-color: var(--ql-text);
		transition: transform 0.2s;
		transform: translateX(0.125rem);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}

	.toggle-switch.active .switch-handle {
		transform: translateX(1.25rem);
	}

	.delete-btn {
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
		margin: 0 auto;
	}

	.delete-btn:hover {
		color: var(--ql-danger, #ef4444);
		background-color: var(--ql-surface-2);
	}

	/* Notifications list */
	.empty-alerts {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 4rem 1.5rem;
		color: var(--ql-text-muted);
	}

	.all-clear-icon {
		color: var(--ql-success, #22c55e);
		margin-bottom: 1rem;
		opacity: 0.8;
	}

	.empty-alerts h4 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--ql-text);
		margin-bottom: 0.25rem;
	}

	.empty-alerts p {
		font-size: 0.825rem;
		max-width: 15rem;
	}

	.alerts-scroller {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		overflow-y: auto;
		max-height: 450px;
	}

	.alert-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 0.875rem;
		background-color: rgba(239, 68, 68, 0.03);
		border: 1px solid rgba(239, 68, 68, 0.15);
		border-radius: 0.375rem;
	}

	.alert-item.acked {
		background-color: rgba(255, 255, 255, 0.01);
		border-color: var(--ql-border);
		opacity: 0.7;
	}

	.alert-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.alert-msg {
		font-size: 0.85rem;
		color: var(--ql-text);
		word-break: break-word;
	}

	.alert-context-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}

	.alert-context-row span {
		padding: 0.2rem 0.45rem;
		border: 1px solid var(--ql-border);
		border-radius: 999px;
		color: var(--ql-text-muted);
		font-size: 0.72rem;
	}

	.alert-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.ack-btn {
		font-size: 0.75rem;
		padding: 0.25rem 0.625rem;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.status-label-acked {
		font-size: 0.75rem;
		color: var(--ql-text-muted);
		font-weight: 500;
	}

	.incident-link {
		color: var(--ql-accent);
		font-size: 0.75rem;
		text-decoration: none;
		font-weight: 700;
	}

	.text-center { text-align: center; }
	.text-muted { color: var(--ql-text-muted); }
	.text-xs { font-size: 0.75rem; }
	.text-uppercase { text-transform: uppercase; font-size: 0.75rem; }

	/* Modal */
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
		max-width: 480px;
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

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		border-top: 1px solid var(--ql-border);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
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
