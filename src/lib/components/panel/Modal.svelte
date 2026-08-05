<script lang="ts">
	import { fade, scale } from 'svelte/transition';

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
	class="fixed inset-0 z-40 grid place-items-center p-4 backdrop-blur-md"
	role="presentation"
	style="background:var(--ui-overlay);"
	on:click={(event) => event.target === event.currentTarget && onClose()}
	transition:fade
>
	<div
		class="flex max-h-[92vh] w-full max-w-2xl flex-col overflow-hidden rounded-[28px] border border-border-subtle bg-panel-strong shadow-panel"
		role="dialog"
		aria-modal="true"
		aria-label={title}
		transition:scale={{ start: 0.96, duration: 160 }}
	>
		<header class="flex items-center justify-between gap-4 border-b border-border-subtle px-5 py-4">
			<h2 class="font-display text-xl font-bold text-primary">{title}</h2>
			<button
				class="grid h-9 w-9 place-items-center rounded-full border border-border-subtle bg-surface text-secondary shadow-card transition hover:text-primary"
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
		<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
			<slot />
		</div>
	</div>
</div>
