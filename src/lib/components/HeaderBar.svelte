<script lang="ts">
	import type { NavigationItem, PlannerUser } from '../types/planner';
	import type { ThemeMode } from '../stores/theme';
	import ThemeToggle from './ThemeToggle.svelte';

	export let navigation: NavigationItem[] = [];
	export let user: PlannerUser;
	export let themeMode: ThemeMode;
	export let onToggleTheme: () => void;
</script>

<header class="sticky top-0 z-30 border-b border-border-subtle bg-panel backdrop-blur-xl">
	<div
		class="mx-auto flex w-full max-w-[1880px] flex-wrap items-center justify-between gap-3 px-3 py-2.5 sm:px-4 xl:px-6"
	>
		<div class="flex min-w-0 items-center gap-3">
			<div
				class="grid h-10 w-10 shrink-0 place-items-center rounded-2xl border border-border-subtle bg-surface shadow-card"
			>
				<svg
					class="h-5 w-5 stroke-accent"
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
				<p class="text-[9px] font-extrabold tracking-[0.26em] text-accent uppercase">
					Uni Timetable
				</p>
				<h1 class="truncate font-display text-base leading-tight text-primary sm:text-lg">
					Horario
				</h1>
			</div>
		</div>

		<nav class="hidden flex-1 justify-center lg:flex" aria-label="Primary">
			<div
				class="flex items-center gap-1 rounded-full border border-border-subtle bg-surface p-1 shadow-card"
			>
				{#each navigation as item (item.id)}
					<button
						class={`rounded-full px-3 py-1.5 text-xs font-bold transition duration-200 ${
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

			<button
				class="grid h-9 w-9 place-items-center rounded-full border border-border-subtle text-xs font-extrabold text-primary shadow-card transition duration-200 hover:-translate-y-0.5"
				type="button"
				aria-label={user.avatarLabel}
				style="background:linear-gradient(135deg,var(--ui-accent-soft),var(--ui-surface));"
			>
				{user.initials}
			</button>
		</div>
	</div>
</header>
