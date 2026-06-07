<script lang="ts">
	import { onMount } from 'svelte';
	import { createIncident, getIncidentEvents, getIncidents, updateIncident, type Incident, type IncidentEvent } from '$api/operations';
	import IncidentTimeline from '$components/incidents/IncidentTimeline.svelte';
	import Badge from '$components/shared/Badge.svelte';
	import { addToast } from '$stores/ui';
	import { CheckCircle2, Flame, Plus, RefreshCw } from 'lucide-svelte';

	let incidents = $state([] as Incident[]);
	let selectedIncidentId = $state('');
	let events = $state([] as IncidentEvent[]);
	let loading = $state(true);
	let creating = $state(false);

	let title = $state('');
	let severity = $state('warning');
	let owner = $state('');
	let summary = $state('');

	let selectedIncident = $derived(incidents.find((incident) => incident.id === selectedIncidentId) || null);

	onMount(() => {
		loadIncidents();
	});

	async function loadIncidents() {
		loading = true;
		try {
			incidents = await getIncidents();
			if (!selectedIncidentId && incidents[0]) selectedIncidentId = incidents[0].id;
			if (selectedIncidentId) await loadEvents(selectedIncidentId);
		} catch (err) {
			console.error(err);
			addToast('Failed to load incidents.', 'error');
		} finally {
			loading = false;
		}
	}

	async function loadEvents(id: string) {
		selectedIncidentId = id;
		try {
			events = await getIncidentEvents(id);
		} catch (err) {
			console.error(err);
			events = [];
		}
	}

	async function handleCreateIncident(event: SubmitEvent) {
		event.preventDefault();
		creating = true;
		try {
			const incident = await createIncident({ title, severity, owner_id: owner, summary });
			addToast('Incident declared.', 'success');
			title = '';
			owner = '';
			summary = '';
			selectedIncidentId = incident.id;
			await loadIncidents();
		} catch (err) {
			console.error(err);
			addToast('Failed to declare incident.', 'error');
		} finally {
			creating = false;
		}
	}

	async function setStatus(incident: Incident, status: Incident['status']) {
		try {
			await updateIncident(incident.id, { status });
			addToast(`Incident moved to ${status}.`, 'success');
			await loadIncidents();
		} catch (err) {
			console.error(err);
			addToast('Failed to update incident.', 'error');
		}
	}

	function formatDateTime(value: string): string {
		try {
			return new Date(value).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return value;
		}
	}
</script>

<div class="incidents-page">
	<div class="page-header">
		<div>
			<h2>Incident Management</h2>
			<p>Declare, assign, investigate, mitigate, monitor, and resolve operational incidents.</p>
		</div>
		<button class="ql-btn-secondary" type="button" onclick={loadIncidents}>
			<RefreshCw size={15} /> Refresh
		</button>
	</div>

	<div class="incident-layout">
		<section class="incident-list-card">
			<h3>Response Queue</h3>
			{#if loading}
				<div class="empty-state">Loading incidents...</div>
			{:else}
				<div class="incident-list">
					{#each incidents as incident}
						<button
							type="button"
							class="incident-row"
							class:active={incident.id === selectedIncidentId}
							onclick={() => loadEvents(incident.id)}
						>
							<div>
								<strong>{incident.title}</strong>
								<span>{incident.status} · {formatDateTime(incident.created_at)}</span>
							</div>
							<Badge
								text={incident.severity}
								variant={incident.severity === 'critical' ? 'danger' : incident.severity === 'warning' ? 'warning' : 'info'}
							/>
						</button>
					{:else}
						<div class="empty-state">No active incidents. Declare one manually or from an alert.</div>
					{/each}
				</div>
			{/if}
		</section>

		<section class="incident-detail-card">
			{#if selectedIncident}
				<div class="detail-header">
					<div>
						<h3>{selectedIncident.title}</h3>
						<p>{selectedIncident.summary || 'No incident summary has been added.'}</p>
					</div>
					<Badge text={selectedIncident.status} variant={selectedIncident.status === 'resolved' ? 'success' : 'warning'} />
				</div>
				<div class="incident-actions">
					<button type="button" onclick={() => setStatus(selectedIncident, 'mitigating')}>Mitigating</button>
					<button type="button" onclick={() => setStatus(selectedIncident, 'monitoring')}>Monitoring</button>
					<button type="button" onclick={() => setStatus(selectedIncident, 'resolved')}>
						<CheckCircle2 size={14} /> Resolve
					</button>
				</div>
				<IncidentTimeline {events} />
			{:else}
				<div class="empty-state">
					<Flame size={22} />
					Select an incident to view its response timeline.
				</div>
			{/if}
		</section>

		<section class="declare-card">
			<h3><Plus size={16} /> Declare Incident</h3>
			<form onsubmit={handleCreateIncident}>
				<label>
					Title
					<input class="ql-input" bind:value={title} placeholder="e.g. Elevated LLM latency" required />
				</label>
				<label>
					Severity
					<select class="ql-input" bind:value={severity}>
						<option value="critical">Critical</option>
						<option value="warning">Warning</option>
						<option value="info">Info</option>
					</select>
				</label>
				<label>
					Owner
					<input class="ql-input" bind:value={owner} placeholder="Team or user ID" />
				</label>
				<label>
					Summary
					<textarea class="ql-input" bind:value={summary} placeholder="Current impact and investigation notes"></textarea>
				</label>
				<button class="ql-btn-primary" type="submit" disabled={creating}>
					<Plus size={15} /> Declare
				</button>
			</form>
		</section>
	</div>
</div>

<style>
	.incidents-page {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.page-header,
	.detail-header,
	.incident-actions,
	.declare-card h3 {
		display: flex;
		align-items: center;
	}

	.page-header {
		justify-content: space-between;
		gap: 1rem;
	}

	h2,
	h3,
	p {
		margin: 0;
	}

	.page-header p,
	.detail-header p,
	.incident-row span,
	.empty-state {
		color: var(--ql-text-muted);
		font-size: 0.875rem;
	}

	.incident-layout {
		display: grid;
		grid-template-columns: minmax(260px, 0.8fr) minmax(0, 1.25fr) minmax(280px, 0.8fr);
		gap: 1rem;
		align-items: start;
	}

	.incident-list-card,
	.incident-detail-card,
	.declare-card {
		padding: 1rem;
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius-sm);
		background: var(--ql-surface);
	}

	.incident-list-card,
	.incident-detail-card,
	.declare-card form {
		display: flex;
		flex-direction: column;
		gap: 0.875rem;
	}

	.incident-list {
		display: flex;
		flex-direction: column;
		gap: 0.625rem;
	}

	.incident-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-bg);
		color: var(--ql-text);
		cursor: pointer;
		text-align: left;
	}

	.incident-row.active,
	.incident-row:hover {
		border-color: var(--ql-accent);
	}

	.incident-row div,
	.declare-card label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		min-width: 0;
	}

	.detail-header {
		justify-content: space-between;
		gap: 1rem;
	}

	.incident-actions {
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.incident-actions button,
	.ql-btn-primary,
	.ql-btn-secondary {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}

	.incident-actions button {
		padding: 0.45rem 0.625rem;
		border: 1px solid var(--ql-border);
		border-radius: 6px;
		background: var(--ql-bg);
		color: var(--ql-text);
		cursor: pointer;
	}

	.declare-card h3 {
		gap: 0.5rem;
		margin-bottom: 0.875rem;
	}

	.declare-card label {
		color: var(--ql-text-muted);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.empty-state {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem;
		border: 1px dashed var(--ql-border);
		border-radius: 6px;
	}

	@media (max-width: 1180px) {
		.incident-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
