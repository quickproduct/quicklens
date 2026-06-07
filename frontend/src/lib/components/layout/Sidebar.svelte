<script lang="ts">
	import { page } from '$app/stores';
	import { sidebarCollapsed, toggleSidebar } from '$stores/ui';
	import { currentUser, logout } from '$stores/auth';
	import {
		LayoutDashboard,
		Cpu,
		GitBranch,
		MessageSquare,
		ScrollText,
		Star,
		Bell,
		Settings,
		ChevronLeft,
		ChevronRight,
		LogOut,
		Search as SearchIcon
	} from 'lucide-svelte';

	const navItems = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/models', label: 'Models', icon: Cpu },
		{ href: '/traces', label: 'Traces', icon: GitBranch },
		{ href: '/prompts', label: 'Prompts', icon: MessageSquare },
		{ href: '/logs', label: 'Logs', icon: ScrollText },
		{ href: '/evaluations', label: 'Evaluations', icon: Star },
		{ href: '/alerts', label: 'Alerts', icon: Bell },
		{ href: '/settings', label: 'Settings', icon: Settings }
	];

	let collapsed = $state(false);

	$effect(() => {
		const unsub = sidebarCollapsed.subscribe((v) => (collapsed = v));
		return unsub;
	});

	let userName = $state('');
	let userEmail = $state('');

	$effect(() => {
		const unsub = currentUser.subscribe((u) => {
			userName = u?.full_name || 'User';
			userEmail = u?.email || '';
		});
		return unsub;
	});

	let currentPath = $state('/');

	$effect(() => {
		const unsub = page.subscribe((p) => {
			currentPath = p.url.pathname;
		});
		return unsub;
	});

	function isActive(href: string): boolean {
		if (href === '/') return currentPath === '/';
		return currentPath.startsWith(href);
	}

	function handleLogout() {
		logout();
	}
</script>

<aside
	class="sidebar"
	class:collapsed
>
	<!-- Logo -->
	<div class="sidebar-logo">
		<div class="logo-icon">
			<svg width="28" height="28" viewBox="0 0 28 28" fill="none">
				<circle cx="14" cy="14" r="12" stroke="#10b981" stroke-width="2" fill="none" />
				<circle cx="14" cy="14" r="5" fill="#10b981" opacity="0.3" />
				<circle cx="14" cy="14" r="2.5" fill="#10b981" />
				<line x1="22" y1="22" x2="27" y2="27" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" />
			</svg>
		</div>
		{#if !collapsed}
			<span class="logo-text">QuickLens</span>
		{/if}
	</div>

	<!-- Navigation -->
	<nav class="sidebar-nav ql-scrollbar">
		{#each navItems as item}
			{@const Icon = item.icon}
			<a
				href={item.href}
				class="nav-item"
				class:active={isActive(item.href)}
				title={collapsed ? item.label : undefined}
			>
				<span class="nav-icon">
					<Icon size={20} />
				</span>
				{#if !collapsed}
					<span class="nav-label">{item.label}</span>
				{/if}
				{#if isActive(item.href)}
					<span class="nav-indicator"></span>
				{/if}
			</a>
		{/each}
	</nav>

	<!-- Bottom Section -->
	<div class="sidebar-bottom">
		<!-- Collapse Toggle -->
		<button class="collapse-btn" onclick={toggleSidebar}>
			{#if collapsed}
				<ChevronRight size={18} />
			{:else}
				<ChevronLeft size={18} />
				<span>Collapse</span>
			{/if}
		</button>

		<!-- User info -->
		{#if !collapsed}
			<div class="sidebar-user">
				<div class="user-avatar">
					{userName.charAt(0).toUpperCase()}
				</div>
				<div class="user-info">
					<span class="user-name">{userName}</span>
					<span class="user-email">{userEmail}</span>
				</div>
				<button class="logout-btn" onclick={handleLogout} title="Log out">
					<LogOut size={16} />
				</button>
			</div>
		{:else}
			<button class="nav-item" onclick={handleLogout} title="Log out">
				<span class="nav-icon">
					<LogOut size={20} />
				</span>
			</button>
		{/if}
	</div>
</aside>

<style>
	.sidebar {
		width: var(--ql-sidebar-width);
		height: 100vh;
		background: rgba(26, 29, 46, 0.95);
		backdrop-filter: blur(16px);
		border-right: 1px solid var(--ql-border);
		display: flex;
		flex-direction: column;
		transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
		position: fixed;
		left: 0;
		top: 0;
		z-index: 40;
		overflow: hidden;
	}

	.sidebar.collapsed {
		width: var(--ql-sidebar-collapsed);
	}

	.sidebar-logo {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 20px 20px 16px;
		border-bottom: 1px solid var(--ql-border);
		min-height: 64px;
	}

	.logo-icon {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.logo-text {
		font-size: 1.2rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		background: linear-gradient(135deg, var(--ql-accent) 0%, var(--ql-accent-hover) 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		white-space: nowrap;
	}

	.sidebar-nav {
		flex: 1;
		padding: 12px 8px;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: 8px;
		color: var(--ql-text-muted);
		text-decoration: none;
		font-size: 0.875rem;
		font-weight: 500;
		transition: all 0.15s ease;
		position: relative;
		cursor: pointer;
		border: none;
		background: none;
		width: 100%;
		text-align: left;
		font-family: var(--ql-font-ui);
	}

	.nav-item:hover {
		color: var(--ql-text);
		background: var(--ql-accent-subtle);
	}

	.nav-item.active {
		color: var(--ql-accent);
		background: var(--ql-accent-subtle);
	}

	.nav-icon {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
	}

	.nav-label {
		white-space: nowrap;
	}

	.nav-indicator {
		position: absolute;
		left: 0;
		top: 50%;
		transform: translateY(-50%);
		width: 3px;
		height: 20px;
		background: var(--ql-accent);
		border-radius: 0 3px 3px 0;
	}

	.sidebar-bottom {
		padding: 8px;
		border-top: 1px solid var(--ql-border);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.collapse-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		border: none;
		background: none;
		color: var(--ql-text-muted);
		font-size: 0.8rem;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s ease;
		font-family: var(--ql-font-ui);
	}

	.collapse-btn:hover {
		background: var(--ql-accent-subtle);
		color: var(--ql-text);
	}

	.sidebar-user {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 8px;
		background: rgba(36, 40, 66, 0.5);
	}

	.user-avatar {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: linear-gradient(135deg, var(--ql-accent) 0%, #0ea5e9 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.8rem;
		font-weight: 700;
		color: white;
		flex-shrink: 0;
	}

	.user-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.user-name {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--ql-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.user-email {
		font-size: 0.7rem;
		color: var(--ql-text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.logout-btn {
		padding: 6px;
		border: none;
		background: none;
		color: var(--ql-text-muted);
		cursor: pointer;
		border-radius: 6px;
		transition: all 0.15s ease;
		display: flex;
		align-items: center;
	}

	.logout-btn:hover {
		color: var(--ql-danger);
		background: rgba(239, 68, 68, 0.1);
	}
</style>
