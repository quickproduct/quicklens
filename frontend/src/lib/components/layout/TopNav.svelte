<script lang="ts">
	import { page } from '$app/stores';
	import { currentUser, logout } from '$stores/auth';
	import { wsConnected } from '$stores/ui';
	import { User, LogOut, ChevronDown } from 'lucide-svelte';

	let currentPath = $state('/');
	let connected = $state(false);
	let userName = $state('User');
	let showDropdown = $state(false);

	$effect(() => {
		const unsub = page.subscribe((p) => {
			currentPath = p.url.pathname;
		});
		return unsub;
	});

	$effect(() => {
		const unsub = wsConnected.subscribe((v) => (connected = v));
		return unsub;
	});

	$effect(() => {
		const unsub = currentUser.subscribe((u) => {
			userName = u?.full_name || 'User';
		});
		return unsub;
	});

	function getPageTitle(path: string): string {
		if (path === '/') return 'Dashboard';
		if (path.startsWith('/models')) return 'Models';
		if (path.startsWith('/traces/')) return 'Trace Detail';
		if (path.startsWith('/traces')) return 'Traces';
		if (path.startsWith('/prompts')) return 'Prompts';
		if (path.startsWith('/logs')) return 'Live Logs';
		if (path.startsWith('/evaluations')) return 'Evaluations';
		if (path.startsWith('/alerts')) return 'Alerts';
		if (path.startsWith('/settings')) return 'Settings';
		return 'QuickLens';
	}

	function handleLogout() {
		showDropdown = false;
		logout();
	}

	function handleClickOutside(event: MouseEvent) {
		if (showDropdown) {
			showDropdown = false;
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<header class="topnav">
	<div class="topnav-left">
		<h1 class="page-title">{getPageTitle(currentPath)}</h1>
	</div>

	<div class="topnav-right">
		<!-- Connection Status -->
		<div class="connection-status" title={connected ? 'Connected' : 'Disconnected'}>
			<span class="status-dot" class:connected></span>
			<span class="status-text">{connected ? 'Live' : 'Offline'}</span>
		</div>

		<!-- User Dropdown -->
		<div class="user-dropdown-wrapper">
			<button
				class="user-dropdown-trigger"
				onclick={(e) => { e.stopPropagation(); showDropdown = !showDropdown; }}
			>
				<div class="user-avatar-sm">
					{userName.charAt(0).toUpperCase()}
				</div>
				<span class="user-dropdown-name">{userName}</span>
				<ChevronDown size={14} />
			</button>

			{#if showDropdown}
				<div class="user-dropdown-menu animate-fade-in" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="menu" tabindex="-1">
					<a href="/settings" class="dropdown-item" onclick={() => showDropdown = false}>
						<User size={16} />
						<span>Profile & Settings</span>
					</a>
					<div class="dropdown-divider"></div>
					<button class="dropdown-item danger" onclick={handleLogout}>
						<LogOut size={16} />
						<span>Log out</span>
					</button>
				</div>
			{/if}
		</div>
	</div>
</header>

<style>
	.topnav {
		height: var(--ql-topnav-height);
		background: rgba(26, 29, 46, 0.6);
		backdrop-filter: blur(12px);
		border-bottom: 1px solid var(--ql-border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 24px;
		position: sticky;
		top: 0;
		z-index: 30;
	}

	.topnav-left {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.page-title {
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--ql-text);
		margin: 0;
	}

	.topnav-right {
		display: flex;
		align-items: center;
		gap: 20px;
	}

	.connection-status {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 12px;
		border-radius: 9999px;
		background: var(--ql-surface);
		border: 1px solid var(--ql-border);
		font-size: 0.75rem;
	}

	.status-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: var(--ql-danger);
		transition: background 0.3s ease;
	}

	.status-dot.connected {
		background: var(--ql-success);
		animation: ql-pulse-dot 2s ease-in-out infinite;
	}

	.status-text {
		color: var(--ql-text-muted);
		font-weight: 500;
	}

	.user-dropdown-wrapper {
		position: relative;
	}

	.user-dropdown-trigger {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px 12px 4px 4px;
		border: 1px solid var(--ql-border);
		border-radius: 9999px;
		background: var(--ql-surface);
		color: var(--ql-text);
		cursor: pointer;
		font-family: var(--ql-font-ui);
		font-size: 0.85rem;
		transition: all 0.15s ease;
	}

	.user-dropdown-trigger:hover {
		border-color: var(--ql-text-muted);
	}

	.user-avatar-sm {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--ql-accent) 0%, #0ea5e9 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 700;
		color: white;
	}

	.user-dropdown-name {
		font-weight: 500;
	}

	.user-dropdown-menu {
		position: absolute;
		top: calc(100% + 8px);
		right: 0;
		min-width: 200px;
		background: var(--ql-surface);
		border: 1px solid var(--ql-border);
		border-radius: var(--ql-radius);
		box-shadow: var(--ql-shadow-lg);
		overflow: hidden;
		z-index: 50;
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		color: var(--ql-text);
		text-decoration: none;
		font-size: 0.85rem;
		transition: background 0.15s ease;
		border: none;
		background: none;
		width: 100%;
		cursor: pointer;
		font-family: var(--ql-font-ui);
		text-align: left;
	}

	.dropdown-item:hover {
		background: var(--ql-accent-subtle);
	}

	.dropdown-item.danger {
		color: #f87171;
	}

	.dropdown-item.danger:hover {
		background: rgba(239, 68, 68, 0.1);
	}

	.dropdown-divider {
		height: 1px;
		background: var(--ql-border);
	}
</style>
