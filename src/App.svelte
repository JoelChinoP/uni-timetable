<script lang="ts">
	import { onDestroy, onMount } from 'svelte';

	import TopBar from './lib/components/TopBar.svelte';
	import LoginPage from './lib/pages/LoginPage.svelte';
	import AdminPanel from './lib/pages/AdminPanel.svelte';
	import { loadSession, logoutSession } from './lib/api/auth';
	import type { AuthUser } from './lib/types/auth';

	let path = window.location.pathname;
	let user: AuthUser | null = null;
	let sessionKnown = false;

	function navigate(nextPath: string) {
		window.history.pushState({}, '', nextPath);
		path = nextPath;
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
		<AdminPanel {user} onNavigate={navigate} onLogout={logoutUser} />
	{:else}
		<div class="grid min-h-dvh place-items-center text-secondary">Comprobando sesión…</div>
	{/if}
{:else}
	<div class="min-h-dvh">
		<TopBar {user} busy={!sessionKnown} onNavigate={navigate} onLogout={logoutUser} />
		<main class="mx-auto grid w-full max-w-7xl flex-1 gap-4 p-4 sm:p-6 lg:grid-cols-[3fr_1fr]">
			<section
				class="min-h-[28rem] rounded-[28px] border border-border-subtle bg-panel p-5 shadow-panel backdrop-blur-xl lg:min-h-[calc(100dvh-7rem)]"
			>
				<p class="text-xs font-extrabold tracking-[0.26em] text-accent uppercase">
					Tablero principal
				</p>
				<h1 class="mt-3 font-display text-4xl font-bold text-primary sm:text-5xl">
					Prepara tu horario
				</h1>
				<p class="mt-4 max-w-lg text-sm leading-6 text-secondary sm:text-base">
					Este espacio queda listo para integrar el planificador después de la autenticación.
				</p>
				<div
					class="mt-6 grid min-h-72 place-items-center rounded-[24px] border border-dashed border-border-strong bg-surface-muted text-sm font-bold text-secondary"
				>
					Tablero de cursos
				</div>
			</section>

			<aside
				class="min-h-80 rounded-[28px] border border-border-subtle bg-panel p-5 shadow-panel backdrop-blur-xl lg:min-h-[calc(100dvh-7rem)]"
			>
				<p class="text-xs font-extrabold tracking-[0.26em] text-accent uppercase">Resumen</p>
				<h2 class="mt-3 font-display text-2xl font-bold text-primary">Estado</h2>
				{#if user}
					<div class="mt-5 rounded-[22px] bg-surface-muted p-4">
						<p class="truncate font-bold text-primary">{user.displayName}</p>
						<p class="mt-1 truncate text-sm text-secondary">{user.email}</p>
						<span
							class="mt-3 inline-flex rounded-full bg-accent-soft px-3 py-1 text-xs font-extrabold text-accent"
						>
							{user.role}
						</span>
					</div>
					{#if user.role === 'ADMIN'}
						<button
							class="mt-4 w-full rounded-[18px] bg-accent-strong px-4 py-3 text-sm font-bold text-white"
							type="button"
							on:click={() => navigate('/panel')}
						>
							Gestionar usuarios
						</button>
					{/if}
				{:else}
					<p class="mt-5 text-sm leading-6 text-secondary">
						No has iniciado sesión. Usa Google para conectar tu cuenta cuando necesites administrar
						o guardar información.
					</p>
				{/if}
			</aside>
		</main>
	</div>
{/if}
