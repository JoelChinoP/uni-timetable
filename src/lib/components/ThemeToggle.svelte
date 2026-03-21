<script lang="ts">
	import type { ThemeMode } from '../stores/theme';

	export let theme: ThemeMode;
	export let onToggle: () => void;
</script>

<button
	class="group inline-flex items-center gap-2 rounded-full border border-border-subtle bg-surface px-2 py-1.5 text-xs font-semibold text-secondary shadow-card transition duration-200 hover:-translate-y-0.5 hover:border-accent/30 hover:text-primary"
	type="button"
	aria-label={`Cambiar a modo ${theme === 'dark' ? 'claro' : 'oscuro'}`}
	on:click={onToggle}
>
	<span
		class="grid h-7 w-7 place-items-center rounded-full bg-accent-soft text-accent transition duration-200 group-hover:scale-105"
	>
		{#if theme === 'dark'}
			<svg
				class="h-4 w-4 stroke-current"
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
				class="h-4 w-4 stroke-current"
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
</button>
