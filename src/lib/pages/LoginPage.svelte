<script lang="ts">
	import { onMount } from 'svelte';

	import { loginWithGoogle } from '../api/auth';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let onSigned: (user: AuthUser) => void;
	export let onNavigate: (path: string) => void;

	const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID?.trim() ?? '';

	let buttonElement: HTMLDivElement;
	let errorMessage = '';
	let isSubmitting = false;

	async function handleCredential(response: { credential?: string }) {
		if (!response.credential || isSubmitting) {
			return;
		}

		isSubmitting = true;
		errorMessage = '';

		try {
			onSigned(await loginWithGoogle(response.credential));
			onNavigate('/');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'No se pudo iniciar sesión';
		} finally {
			isSubmitting = false;
		}
	}

	onMount(() => {
		if (user) {
			return;
		}
		if (!clientId) {
			errorMessage = 'Falta configurar VITE_GOOGLE_CLIENT_ID.';
			return;
		}

		let attempts = 0;
		const timer = window.setInterval(() => {
			attempts += 1;
			const googleId = window.google?.accounts?.id;

			if (googleId) {
				window.clearInterval(timer);
				googleId.initialize({ client_id: clientId, callback: handleCredential });
				googleId.renderButton(buttonElement, {
					theme: 'outline',
					size: 'large',
					shape: 'pill',
					text: 'signin_with',
					locale: 'es',
				});
			} else if (attempts >= 50) {
				window.clearInterval(timer);
				errorMessage = 'No se pudo cargar Google Identity Services.';
			}
		}, 100);

		return () => window.clearInterval(timer);
	});
</script>

<svelte:head>
	<title>Iniciar sesión | Horarios</title>
</svelte:head>

<main id="main-content" class="grid min-h-dvh place-items-center px-4 py-10">
	<section class="neo-panel w-full max-w-md p-6 text-center sm:p-8">
		<img class="mx-auto h-14 w-14" src="/logo.svg" alt="" />
		<p class="mt-4 text-xs font-extrabold tracking-[0.24em] text-accent uppercase">Horarios</p>
		<h1 class="mt-3 font-display text-3xl font-extrabold text-primary">Inicia sesión</h1>
		<p class="mt-3 text-sm leading-6 text-secondary">
			Usa únicamente tu cuenta de Google para entrar al planificador.
		</p>

		{#if user}
			<div class="neo-control mt-7 p-5 text-left">
				<p class="text-xs font-extrabold tracking-[0.2em] text-muted uppercase">Sesión activa</p>
				<h2 class="mt-2 text-lg font-bold text-primary">{user.displayName}</h2>
				<p class="truncate text-sm text-secondary">{user.email}</p>
			</div>
			<div class="mt-6 flex justify-center gap-3">
				<button
					class="min-h-11 rounded-xl bg-accent-strong px-5 py-2.5 text-sm font-bold text-white"
					type="button"
					on:click={() => onNavigate('/')}
				>
					Ir al inicio
				</button>
				{#if user.role === 'ADMIN'}
					<button
						class="neo-button min-h-11 px-5 py-2.5 text-sm font-bold text-primary"
						type="button"
						on:click={() => onNavigate('/panel')}
					>
						Ir al panel
					</button>
				{/if}
			</div>
		{:else}
			<div class="mt-8 flex min-h-11 justify-center" bind:this={buttonElement}></div>
			{#if isSubmitting}
				<p class="mt-4 text-sm text-secondary">Verificando con el servidor…</p>
			{/if}
			<button
				class="mt-6 min-h-11 rounded-xl px-3 text-sm font-bold text-secondary transition hover:bg-surface-muted hover:text-primary"
				type="button"
				on:click={() => onNavigate('/')}
			>
				Volver al inicio
			</button>
		{/if}

		{#if errorMessage}
			<p
				class="mt-5 rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning"
				role="alert"
			>
				{errorMessage}
			</p>
		{/if}
	</section>
</main>
