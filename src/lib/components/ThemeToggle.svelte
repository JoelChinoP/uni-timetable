<script lang="ts">
	import type { ThemeMode } from '../stores/theme';

	export let theme: ThemeMode;
	export let onToggle: () => void;
</script>

<button
	class="neo-button grid h-11 w-11 shrink-0 place-items-center text-secondary"
	type="button"
	title={`Cambiar a modo ${theme === 'dark' ? 'claro' : 'oscuro'}`}
	aria-label={`Cambiar a modo ${theme === 'dark' ? 'claro' : 'oscuro'}`}
	on:click={onToggle}
>
	{#if theme === 'dark'}
		<svg
			class="h-5 w-5 stroke-current"
			viewBox="0 0 24 24"
			fill="none"
			stroke-width="1.8"
			stroke-linecap="round"
			aria-hidden="true"
		>
			<circle cx="12" cy="12" r="4" />
			<path
				d="M12 2v2M12 20v2M4.93 4.93l1.42 1.42M17.65 17.65l1.42 1.42M2 12h2M20 12h2M4.93 19.07l1.42-1.42M17.65 6.35l1.42-1.42"
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
			<path d="M20.5 14.1A8.5 8.5 0 0 1 9.9 3.5 8.5 8.5 0 1 0 20.5 14.1Z" />
		</svg>
	{/if}
</button>
