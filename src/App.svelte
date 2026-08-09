<script lang="ts">
	import { onDestroy, onMount } from 'svelte';

	import LoginPage from './lib/pages/LoginPage.svelte';
	import PanelPage from './lib/pages/PanelPage.svelte';
	import DashboardPage from './lib/pages/DashboardPage.svelte';
	import SharedPage from './lib/pages/SharedPage.svelte';
	import ClassroomsPage from './lib/pages/ClassroomsPage.svelte';
	import NotFoundPage from './lib/pages/NotFoundPage.svelte';
	import { loadSession, logoutSession } from './lib/api/auth';
	import type { AuthUser } from './lib/types/auth';

	let path = window.location.pathname;
	let user: AuthUser | null = null;
	let sessionKnown = false;

	function navigate(nextPath: string) {
		window.history.pushState({}, '', nextPath);
		path = new URL(nextPath, window.location.origin).pathname;
	}

	function onPopState() {
		path = window.location.pathname;
	}

	async function logoutUser() {
		user = null;
		try {
			await logoutSession();
		} catch {
			// ponytail: logout must still clear the UI when the API is unavailable.
		}
		navigate('/');
	}

	onMount(async () => {
		window.addEventListener('popstate', onPopState);
		try {
			user = await loadSession();
		} catch {
			user = null;
		} finally {
			sessionKnown = true;
		}
	});

	onDestroy(() => window.removeEventListener('popstate', onPopState));
</script>

{#if path === '/login'}
	<LoginPage {user} onSigned={(nextUser) => (user = nextUser)} onNavigate={navigate} />
{:else if path === '/panel'}
	{#if sessionKnown}
		<PanelPage {user} onNavigate={navigate} onLogout={logoutUser} />
	{:else}
		<div class="grid min-h-dvh place-items-center text-secondary">Comprobando sesión…</div>
	{/if}
{:else if path === '/aulas'}
	<ClassroomsPage {user} busy={!sessionKnown} onNavigate={navigate} onLogout={logoutUser} />
{:else if /^\/s\/[A-Za-z0-9]{10}$/.test(path)}
	<SharedPage
		shareId={path.slice(3)}
		{user}
		busy={!sessionKnown}
		onNavigate={navigate}
		onLogout={logoutUser}
	/>
{:else if path === '/'}
	<DashboardPage {user} busy={!sessionKnown} onNavigate={navigate} onLogout={logoutUser} />
{:else}
	<NotFoundPage onNavigate={navigate} />
{/if}
