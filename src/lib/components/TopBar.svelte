<script lang="ts">
	import ThemeToggle from './ThemeToggle.svelte';
	import { theme } from '../stores/theme';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;
	export let showShare = false;
	export let sharing = false;
	export let shareDisabled = false;
	export let onShare: () => void = () => {};

	let menuOpen = false;
	let menuRoot: HTMLDivElement;
	let menuButton: HTMLButtonElement;

	$: firstName = user?.displayName.trim().split(/\s+/)[0] || user?.email || 'Usuario';
	$: initial = firstName.charAt(0).toUpperCase();

	function navigate(path: string) {
		menuOpen = false;
		onNavigate(path);
	}

	function closeMenu() {
		if (!menuOpen) return;
		menuOpen = false;
		queueMicrotask(() => menuButton?.focus());
	}

	function handleWindowClick(event: MouseEvent) {
		if (menuOpen && !menuRoot?.contains(event.target as Node)) closeMenu();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') closeMenu();
	}
</script>

<svelte:window on:click={handleWindowClick} on:keydown={handleKeydown} />

<a class="skip-link" href="#main-content">Saltar al contenido</a>
<header
	class="glass-panel sticky top-0 z-30 flex h-16 items-center justify-between border-x-0 border-t-0 px-3 sm:px-5"
>
	<button
		class="flex min-w-0 items-center gap-2.5 rounded-xl p-1 text-primary"
		type="button"
		on:click={() => navigate('/')}
	>
		<img class="h-10 w-10 shrink-0" src="/logo.svg" alt="" />
		<span
			class="truncate font-display text-base font-extrabold tracking-tight sm:text-xl xl:text-2xl"
			>Horarios</span
		>
	</button>

	<div class="flex min-w-0 items-center gap-2">
		{#if showShare}
			<button
				class="neo-button inline-flex h-11 items-center gap-2 px-3 text-sm font-bold text-primary disabled:opacity-45"
				type="button"
				disabled={sharing || shareDisabled}
				title={shareDisabled ? 'Selecciona al menos un grupo' : 'Compartir horario'}
				on:click={onShare}
			>
				<svg
					class="h-4.5 w-4.5 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<circle cx="18" cy="5" r="2.5" /><circle cx="6" cy="12" r="2.5" /><circle
						cx="18"
						cy="19"
						r="2.5"
					/><path d="m8.3 10.8 7.4-4.4M8.3 13.2l7.4 4.4" />
				</svg>
				<span class="hidden sm:inline">{sharing ? 'Creando…' : 'Compartir'}</span>
			</button>
		{/if}

		<ThemeToggle theme={$theme} onToggle={theme.toggle} />

		{#if busy}
			<span class="hidden text-xs text-secondary sm:inline" role="status">Comprobando…</span>
		{:else if user}
			<div class="relative" bind:this={menuRoot}>
				<button
					class="neo-button flex h-11 min-w-11 items-center gap-2 p-1.5 pr-2.5 text-primary"
					type="button"
					aria-haspopup="menu"
					aria-expanded={menuOpen}
					bind:this={menuButton}
					on:click={(event) => {
						event.stopPropagation();
						menuOpen = !menuOpen;
					}}
				>
					{#if user.avatarUrl}
						<img
							class="h-8 w-8 rounded-full object-cover"
							src={user.avatarUrl}
							alt=""
							referrerpolicy="no-referrer"
						/>
					{:else}
						<span
							class="grid h-8 w-8 place-items-center rounded-full bg-accent-strong text-xs font-extrabold text-white"
							aria-hidden="true">{initial}</span
						>
					{/if}
					<span class="hidden max-w-28 truncate text-sm font-bold sm:block">{firstName}</span>
					<svg
						class="hidden h-4 w-4 stroke-current sm:block"
						viewBox="0 0 24 24"
						fill="none"
						stroke-width="2"
						aria-hidden="true"><path d="m7 9 5 5 5-5" /></svg
					>
				</button>

				{#if menuOpen}
					<div
						class="glass-panel absolute right-0 mt-2 w-52 rounded-2xl p-2"
						role="menu"
						aria-label="Cuenta"
					>
						<p class="truncate px-3 py-2 text-xs text-secondary" title={user.email}>{user.email}</p>
						<button
							class="flex h-11 w-full items-center rounded-xl px-3 text-left text-sm font-semibold text-primary hover:bg-surface-muted"
							type="button"
							role="menuitem"
							on:click={() => navigate('/')}>Dashboard</button
						>
						<button
							class="flex h-11 w-full items-center rounded-xl px-3 text-left text-sm font-semibold text-primary hover:bg-surface-muted"
							type="button"
							role="menuitem"
							on:click={() => navigate('/panel')}>Panel</button
						>
						<div class="my-1 border-t border-border-subtle"></div>
						<button
							class="flex h-11 w-full items-center rounded-xl px-3 text-left text-sm font-semibold text-warning hover:bg-warning-soft"
							type="button"
							role="menuitem"
							on:click={() => {
								menuOpen = false;
								onLogout();
							}}>Salir</button
						>
					</div>
				{/if}
			</div>
		{:else}
			<button
				class="h-11 rounded-xl bg-accent-strong px-3.5 text-sm font-bold text-white transition hover:bg-accent"
				type="button"
				on:click={() => navigate('/login')}>Entrar</button
			>
		{/if}
	</div>
</header>
