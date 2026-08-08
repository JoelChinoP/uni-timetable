<script lang="ts">
	export let onNavigate: (path: string) => void;
</script>

<svelte:head><title>Página no encontrada | Horarios</title></svelte:head>

<main id="main-content" class="grid min-h-dvh place-items-center px-4 py-10">
	<section class="neo-panel w-full max-w-2xl overflow-hidden p-6 text-center sm:p-10">
		<svg
			class="mx-auto h-auto w-full max-w-sm"
			viewBox="0 0 440 260"
			fill="none"
			role="img"
			aria-labelledby="not-found-illustration"
		>
			<title id="not-found-illustration">Calendario sin una ruta asignada</title>
			<path
				d="M75 220h290"
				stroke="var(--ui-border-strong)"
				stroke-width="2"
				stroke-linecap="round"
			/>
			<g class="float-calendar">
				<rect
					x="105"
					y="35"
					width="230"
					height="168"
					rx="28"
					fill="var(--ui-surface-muted)"
					stroke="var(--ui-border-strong)"
					stroke-width="3"
				/>
				<path
					d="M106 84h228M155 51v22M285 51v22"
					stroke="var(--ui-accent)"
					stroke-width="9"
					stroke-linecap="round"
				/>
				<path
					d="M145 116h42v28h-42zM199 116h42v28h-42zM253 116h42v28h-42zM145 156h42v20h-42zM253 156h42v20h-42z"
					fill="var(--ui-accent-soft)"
				/>
				<path
					d="m204 168 16-16 16 16"
					stroke="var(--ui-warning)"
					stroke-width="5"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</g>
		</svg>
		<p class="text-xs font-extrabold tracking-[0.22em] text-accent uppercase">Error 404</p>
		<h1 class="mt-3 font-display text-3xl font-extrabold text-primary sm:text-4xl">
			Esta ruta no tiene horario
		</h1>
		<p class="mx-auto mt-3 max-w-md text-sm leading-6 text-secondary">
			La página que buscas cambió de aula o nunca estuvo registrada.
		</p>
		<button
			class="mt-6 h-12 rounded-xl bg-accent-strong px-5 text-sm font-bold text-white transition hover:bg-accent"
			type="button"
			on:click={() => onNavigate('/')}>Volver al dashboard</button
		>
	</section>
</main>

<style>
	.float-calendar {
		animation: float 4s ease-in-out infinite;
		transform-origin: center;
	}
	@keyframes float {
		50% {
			transform: translateY(-7px);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.float-calendar {
			animation: none;
		}
	}
</style>
