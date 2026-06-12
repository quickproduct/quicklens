<script lang="ts">
	import { slide, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { AlertCircle, Check, ExternalLink } from 'lucide-svelte';
	import type { AlertSummary } from '$api/dashboard';
	import Badge from '$components/shared/Badge.svelte';

	let { alerts = [], onAcknowledge, onDeclareIncident, loading = false } = $props<{
		alerts?: AlertSummary[];
		onAcknowledge?: (id: string) => void;
		onDeclareIncident?: (alert: AlertSummary) => void;
		loading?: boolean;
	}>();

	let containerEl: HTMLElement | undefined = $state();

	function formatDateTime(value: string): string {
		try {
			return new Date(value).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' });
		} catch {
			return value;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.ctrlKey && e.altKey && e.key.toLowerCase() === 'a') {
			e.preventDefault();
			containerEl?.focus();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<section bind:this={containerEl} tabindex="-1" class="rounded-ql-sm border border-ql-border bg-ql-surface p-4 outline-none focus-visible:ring-2 focus-visible:ring-ql-accent" aria-labelledby="alert-inbox-title" aria-live="polite">
	<div class="mb-3 flex items-center justify-between gap-4">
		<div>
			<h3 id="alert-inbox-title" class="m-0 text-[0.95rem] font-semibold text-ql-text">Alert Center</h3>
			<p class="m-0 text-sm text-ql-text-muted">Critical issues first, with direct incident escalation.</p>
		</div>
		<a href="/alerts" class="text-sm font-bold text-ql-accent no-underline">View all</a>
	</div>

	<div class="flex flex-col gap-2.5">
		{#if loading}
			{#each Array(3) as _}
				<div class="flex items-center justify-between gap-3 rounded-md border border-ql-border border-l-4 border-l-ql-border bg-ql-bg p-3">
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2">
							<div class="h-5 w-16 animate-pulse rounded-full bg-ql-surface-2"></div>
							<div class="h-4 w-24 animate-pulse rounded bg-ql-surface-2"></div>
						</div>
						<div class="mt-2 h-4 w-3/4 animate-pulse rounded bg-ql-surface-2"></div>
						<div class="mt-2 flex items-center gap-2">
							<div class="h-5 w-12 animate-pulse rounded-full bg-ql-surface-2"></div>
							<div class="h-5 w-24 animate-pulse rounded-full bg-ql-surface-2"></div>
						</div>
					</div>
				</div>
			{/each}
		{:else}
			{#each alerts.slice(0, 5) as alert (alert.id)}
				{@const severityBorder = alert.severity === 'critical' ? 'border-l-ql-critical' : alert.severity === 'warning' ? 'border-l-ql-warning' : 'border-l-ql-info'}
				<article animate:flip={{ duration: 300 }} transition:slide={{ duration: 250 }} class="flex items-center justify-between gap-3 rounded-md border border-ql-border border-l-4 bg-ql-bg p-3 {severityBorder}">
					<div class="min-w-0">
						<div class="flex flex-wrap items-center gap-2 text-sm text-ql-text-muted">
							<Badge
								text={alert.severity}
								variant={alert.severity === 'critical' ? 'danger' : alert.severity === 'warning' ? 'warning' : 'info'}
							/>
							<span>{formatDateTime(alert.created_at)}</span>
						</div>
						<p class="mt-1.5 mb-0 text-sm leading-snug text-ql-text">{alert.message}</p>
						<div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-ql-text-muted">
							<span class="rounded-full border border-ql-border px-2 py-0.5">{alert.status || 'open'}</span>
							<span class="rounded-full border border-ql-border px-2 py-0.5">{alert.model_id || alert.service_id || 'Unassigned service'}</span>
						</div>
					</div>
					<div class="flex shrink-0 items-center gap-1.5">
						{#if onAcknowledge}
							<button 
								type="button" 
								class="inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-md border border-ql-border bg-ql-surface-2 text-ql-text-muted outline-none transition-colors hover:border-ql-accent hover:text-ql-text focus-visible:border-ql-accent focus-visible:text-ql-text focus-visible:ring-2 focus-visible:ring-ql-accent focus-visible:ring-offset-2 focus-visible:ring-offset-ql-bg" 
								onclick={() => onAcknowledge?.(alert.id)} 
								title="Acknowledge alert" 
								aria-label="Acknowledge alert"
							>
								<Check size={15} />
							</button>
						{/if}
						{#if onDeclareIncident}
							<button 
								type="button" 
								class="inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-md border border-ql-border bg-ql-surface-2 text-ql-accent outline-none transition-colors hover:border-ql-accent hover:text-ql-accent focus-visible:border-ql-accent focus-visible:text-ql-accent focus-visible:ring-2 focus-visible:ring-ql-accent focus-visible:ring-offset-2 focus-visible:ring-offset-ql-bg" 
								onclick={() => onDeclareIncident?.(alert)} 
								title="Declare incident" 
								aria-label="Declare incident"
							>
								<ExternalLink size={15} />
							</button>
						{/if}
					</div>
				</article>
			{:else}
				<div in:fade={{ duration: 300 }} class="flex items-center gap-2 rounded-md border border-dashed border-ql-border p-4 text-sm text-ql-text-muted" role="status" aria-live="polite">
					<AlertCircle size={18} />
					No active alerts. The current operating window is clear.
				</div>
			{/each}
		{/if}
	</div>
</section>
