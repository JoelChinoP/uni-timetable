<script lang="ts">
	import type { NavigationItem, PlannerUser } from '../types/planner';
	import type { ThemeMode } from '../stores/theme';
	import ThemeToggle from './ThemeToggle.svelte';

	export let navigation: NavigationItem[] = [];
	export let user: PlannerUser;
	export let themeMode: ThemeMode;
	export let onToggleTheme: () => void;

	const quickActions = [
		{ label: 'Notificaciones', icon: 'bell' },
		{ label: 'Ajustes', icon: 'settings' },
	] as const;
</script>

<header class="sticky top-0 z-30 border-b border-border-subtle bg-panel backdrop-blur-xl">
	<div
		class="mx-auto flex w-full max-w-[1880px] flex-wrap items-center justify-between gap-4 px-4 py-4 sm:px-6 xl:px-8"
	>
		<div class="flex min-w-0 items-center gap-4">
			<div
				class="grid h-14 w-14 shrink-0 place-items-center rounded-[20px] border border-border-subtle bg-surface shadow-card"
			>
				<svg
					class="h-7 w-7 stroke-accent"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<path d="M4 6.5 12 3l8 3.5v11L12 21l-8-3.5v-11Z" />
					<path d="M4 6.5 12 10l8-3.5" />
					<path d="M12 10v11" />
				</svg>
			</div>

			<div class="min-w-0">
				<p class="text-[11px] font-extrabold uppercase tracking-[0.24em] text-accent">
					Uni Timetable
				</p>
				<h1 class="truncate font-display text-xl leading-tight text-primary">
					Planner Dashboard
				</h1>
			</div>
		</div>

		<nav class="hidden flex-1 justify-center xl:flex" aria-label="Primary">
			<div
				class="flex items-center gap-2 rounded-full border border-border-subtle bg-surface p-1.5 shadow-card"
			>
				{#each navigation as item (item.id)}
					<button
						class={`rounded-full px-4 py-2.5 text-sm font-bold transition duration-200 ${
							item.active
								? 'bg-accent-soft text-accent'
								: 'text-secondary hover:bg-surface-muted hover:text-primary'
						}`}
						type="button"
					>
						{item.label}
					</button>
				{/each}
			</div>
		</nav>

		<div class="flex items-center gap-2 sm:gap-3">
			<ThemeToggle theme={themeMode} onToggle={onToggleTheme} />

			{#each quickActions as action (action.label)}
				<button
					class="hidden h-11 w-11 items-center justify-center rounded-full border border-border-subtle bg-surface text-secondary shadow-card transition duration-200 hover:-translate-y-0.5 hover:border-accent/30 hover:text-primary sm:inline-flex"
					type="button"
					aria-label={action.label}
				>
					{#if action.icon === 'bell'}
						<svg
							class="h-5 w-5 stroke-current"
							viewBox="0 0 24 24"
							fill="none"
							stroke-width="1.8"
							stroke-linecap="round"
							stroke-linejoin="round"
							aria-hidden="true"
						>
							<path
								d="M15 17H5.8A1.8 1.8 0 0 1 4 15.2c0-.48.19-.94.53-1.28L6 12.46V9a6 6 0 1 1 12 0v3.46l1.47 1.46c.34.34.53.8.53 1.28A1.8 1.8 0 0 1 18.2 17H15m0 0a3 3 0 0 1-6 0"
							/>
						</svg>
					{:else}
						<svg
							class="h-5 w-5 stroke-current"
							viewBox="0 0 24 24"
							fill="none"
							stroke-width="1.8"
							stroke-linecap="round"
							stroke-linejoin="round"
							aria-hidden="true"
						>
							<path
								d="m12 3 1.2 2.76 3 .27-2.25 2.03.64 2.94L12 9.5 9.4 11l.64-2.94L7.8 6.03l3-.27L12 3Zm0 8a4 4 0 1 0 0 8a4 4 0 0 0 0-8Z"
							/>
						</svg>
					{/if}
				</button>
			{/each}

			<button
				class="grid h-11 w-11 place-items-center rounded-full border border-border-subtle text-sm font-extrabold text-primary shadow-card transition duration-200 hover:-translate-y-0.5"
				type="button"
				aria-label={user.avatarLabel}
				style="background:linear-gradient(135deg,var(--ui-accent-soft),var(--ui-surface));"
			>
				{user.initials}
			</button>
		</div>
	</div>
</header>
