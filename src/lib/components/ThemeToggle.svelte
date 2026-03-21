<script lang="ts">
	import type { ThemeMode } from '../stores/theme';

	export let theme: ThemeMode;
	export let onToggle: () => void;
</script>

<button
	class="group inline-flex items-center gap-3 rounded-full border border-border-subtle bg-surface px-3 py-2 text-sm font-semibold text-secondary shadow-card transition duration-200 hover:-translate-y-0.5 hover:border-accent/30 hover:text-primary"
	type="button"
	aria-label={`Cambiar a modo ${theme === 'dark' ? 'claro' : 'oscuro'}`}
	on:click={onToggle}
>
	<span
		class="grid h-10 w-10 place-items-center rounded-full bg-accent-soft text-accent transition duration-200 group-hover:scale-105"
	>
		{#if theme === 'dark'}
			<svg
				class="h-5 w-5 stroke-current"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.9"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path
					d="M12 4.5V2m0 20v-2.5m5.303-12.803L19.5 4.5m-15 15 2.197-2.197M19.5 19.5l-2.197-2.197m-10.606 0L4.5 19.5m15-15-2.197 2.197M12 17.5a5.5 5.5 0 1 0 0-11a5.5 5.5 0 0 0 0 11Z"
				/>
			</svg>
		{:else}
			<svg
				class="h-5 w-5 stroke-current"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.9"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z" />
			</svg>
		{/if}
	</span>

	<span class="hidden min-w-0 text-left sm:block">
		<span class="block text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted">
			Tema
		</span>
		<span class="mt-0.5 block truncate text-sm font-bold text-primary">
			{theme === 'dark' ? 'Oscuro' : 'Claro'}
		</span>
	</span>

	<span
		class="hidden rounded-full bg-surface-muted px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted md:inline-flex"
	>
		{theme === 'dark' ? 'Cambiar a claro' : 'Cambiar a oscuro'}
	</span>
</button>
