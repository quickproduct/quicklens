<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { initAuth, isAuthenticated, authLoading } from '$stores/auth';
	import Sidebar from '$components/layout/Sidebar.svelte';
	import TopNav from '$components/layout/TopNav.svelte';
	import Toast from '$components/shared/Toast.svelte';
	import '../app.css';

	let { children } = $props();

	let authenticated = $state(false);
	let loading = $state(true);
	let currentPath = $state('/');

	$effect(() => {
		const unsub = isAuthenticated.subscribe((v) => {
			authenticated = v;
		});
		return unsub;
	});

	$effect(() => {
		const unsub = authLoading.subscribe((v) => {
			loading = v;
		});
		return unsub;
	});

	$effect(() => {
		const unsub = page.subscribe((p) => {
			currentPath = p.url.pathname;
		});
		return unsub;
	});

	onMount(async () => {
		// Initialize auth from localStorage
		await initAuth();

		// Handle routing guards
		if (!authenticated && currentPath !== '/login') {
			goto('/login');
		} else if (authenticated && currentPath === '/login') {
			goto('/');
		}
	});

	// Trigger route guard on auth change or path change
	$effect(() => {
		if (!loading) {
			if (!authenticated && currentPath !== '/login') {
				goto('/login');
			} else if (authenticated && currentPath === '/login') {
				goto('/');
			}
		}
	});
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet" />
</svelte:head>

<div class="app-layout">
	{#if loading}
		<div class="auth-loading-screen">
			<div class="loading-spinner-circle"></div>
			<span class="loading-text">Loading QuickLens...</span>
		</div>
	{:else if currentPath === '/login'}
		<main class="full-screen-main">
			{@render children()}
		</main>
	{:else}
		<div class="authenticated-layout">
			<Sidebar />
			<div class="main-content-area">
				<TopNav />
				<main class="page-viewport">
					{@render children()}
				</main>
			</div>
		</div>
	{/if}

	<Toast />
</div>

<style>
	.app-layout {
		width: 100vw;
		height: 100vh;
		background-color: var(--ql-bg);
		color: var(--ql-text);
		overflow: hidden;
	}

	.auth-loading-screen {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100vh;
		gap: 1rem;
		background-color: var(--ql-bg);
	}

	.loading-spinner-circle {
		width: 2.5rem;
		height: 2.5rem;
		border: 3px solid rgba(16, 185, 129, 0.1);
		border-top-color: var(--ql-accent);
		border-radius: 9999px;
		animation: spin 0.8s linear infinite;
	}

	.loading-text {
		font-size: 0.875rem;
		color: var(--ql-text-muted);
		font-weight: 500;
	}

	.full-screen-main {
		width: 100%;
		height: 100%;
		overflow-y: auto;
	}

	.authenticated-layout {
		display: flex;
		width: 100%;
		height: 100%;
		overflow: hidden;
	}

	.main-content-area {
		display: flex;
		flex-direction: column;
		flex-grow: 1;
		height: 100%;
		overflow: hidden;
		min-width: 0;
	}

	.page-viewport {
		flex-grow: 1;
		padding: 1.5rem;
		overflow-y: auto;
		min-height: 0;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
