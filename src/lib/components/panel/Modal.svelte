<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { trapFocus } from '../../utils/focus';

	export let title: string;
	export let onClose: () => void;

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onClose();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div
	class="fixed inset-0 z-40 grid place-items-center bg-[var(--ui-overlay)] p-3 backdrop-blur-sm sm:p-5"
	role="presentation"
	on:click={(event) => event.target === event.currentTarget && onClose()}
	transition:fade
>
	<div
		class="glass-panel flex max-h-[94dvh] w-full max-w-2xl flex-col overflow-hidden rounded-[24px]"
		role="dialog"
		aria-modal="true"
		aria-labelledby="panel-modal-title"
		tabindex="-1"
		use:trapFocus
		transition:scale={{ start: 0.96, duration: 160 }}
	>
		<header class="flex items-center justify-between gap-4 border-b border-border-subtle px-5 py-4">
			<h2 id="panel-modal-title" class="font-display text-xl font-extrabold text-primary">
				{title}
			</h2>
			<button
				class="neo-button grid h-11 w-11 place-items-center text-secondary"
				type="button"
				aria-label="Cerrar"
				on:click={onClose}
			>
				<svg
					class="h-4 w-4 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="2"
					stroke-linecap="round"
					aria-hidden="true"
				>
					<path d="m6 6 12 12M18 6 6 18" />
				</svg>
			</button>
		</header>
		<div class="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5">
			<slot />
		</div>
	</div>
</div>
