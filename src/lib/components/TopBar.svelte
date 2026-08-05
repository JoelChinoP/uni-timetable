<script lang="ts">
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	$: initial = user?.displayName.charAt(0).toUpperCase() ?? 'U';
</script>

<header
	class="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-border-subtle bg-panel px-4 backdrop-blur-xl sm:px-6"
>
	<button
		class="flex items-center gap-3 font-display text-base font-bold text-primary"
		type="button"
		on:click={() => onNavigate('/')}
	>
		<span class="grid h-9 w-9 place-items-center rounded-2xl bg-accent-soft text-sm text-accent">
			UT
		</span>
		<span>Uni Timetable</span>
	</button>

	<div class="flex min-w-0 items-center gap-3">
		{#if busy}
			<span class="text-sm text-secondary">Comprobando sesión…</span>
		{:else if user}
			<span
				class="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-accent-strong text-sm font-extrabold text-white"
				aria-hidden="true"
			>
				{initial}
			</span>
			<span class="hidden max-w-56 min-w-0 flex-col text-right sm:flex">
				<span class="truncate text-sm font-bold text-primary">{user.displayName}</span>
				<span class="truncate text-xs text-secondary">{user.email}</span>
			</span>
			<span
				class={`rounded-full px-2.5 py-1 text-[11px] font-extrabold uppercase ${
					user.role === 'ADMIN' ? 'bg-accent-soft text-accent' : 'bg-surface-muted text-secondary'
				}`}
			>
				{user.role}
			</span>
			{#if user.role === 'ADMIN'}
				<button
					class="rounded-full border border-border-subtle px-3 py-2 text-sm font-bold text-primary transition hover:bg-surface-muted"
					type="button"
					on:click={() => onNavigate('/panel')}
				>
					Panel
				</button>
			{/if}
			<button
				class="rounded-full border border-border-subtle px-3 py-2 text-sm font-bold text-secondary transition hover:bg-surface-muted hover:text-primary"
				type="button"
				on:click={onLogout}
			>
				Salir
			</button>
		{:else}
			<button
				class="rounded-full bg-accent-strong px-4 py-2 text-sm font-bold text-white transition hover:bg-accent"
				type="button"
				on:click={() => onNavigate('/login')}
			>
				Iniciar sesión
			</button>
		{/if}
	</div>
</header>
