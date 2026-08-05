<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import { createUser, loadUsers } from '../api/auth';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let users: AuthUser[] = [];
	let email = '';
	let displayName = '';
	let isLoading = true;
	let isSubmitting = false;
	let errorMessage = '';
	let successMessage = '';

	async function refreshUsers() {
		isLoading = true;
		errorMessage = '';
		try {
			users = await loadUsers();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'No se pudieron cargar los usuarios';
		} finally {
			isLoading = false;
		}
	}

	async function submitUser() {
		errorMessage = '';
		successMessage = '';
		isSubmitting = true;

		try {
			await createUser(email.trim().toLowerCase(), displayName.trim());
			email = '';
			displayName = '';
			successMessage = 'Usuario registrado correctamente.';
			await refreshUsers();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'No se pudo registrar el usuario';
		} finally {
			isSubmitting = false;
		}
	}

	onMount(() => {
		if (user?.role === 'ADMIN') {
			void refreshUsers();
		}
	});
</script>

<svelte:head>
	<title>Panel | Uni Timetable</title>
</svelte:head>

<div class="min-h-dvh">
	<TopBar {user} {onNavigate} {onLogout} />

	<main class="mx-auto w-full max-w-6xl p-4 sm:p-6">
		{#if !user}
			<section
				class="rounded-[28px] border border-border-subtle bg-panel p-8 text-center shadow-panel"
			>
				<h1 class="font-display text-3xl font-bold text-primary">Necesitas iniciar sesión</h1>
				<p class="mt-3 text-secondary">Entra con Google para continuar.</p>
				<button
					class="mt-6 rounded-full bg-accent-strong px-5 py-2.5 font-bold text-white"
					type="button"
					on:click={() => onNavigate('/login')}
				>
					Ir a login
				</button>
			</section>
		{:else if user.role !== 'ADMIN'}
			<section
				class="rounded-[28px] border border-warning/25 bg-warning-soft p-8 text-center shadow-panel"
			>
				<h1 class="font-display text-3xl font-bold text-warning">Acceso restringido</h1>
				<p class="mt-3 text-secondary">Tu cuenta tiene rol normal y no puede ver este panel.</p>
				<button
					class="mt-6 rounded-full bg-accent-strong px-5 py-2.5 font-bold text-white"
					type="button"
					on:click={() => onNavigate('/')}
				>
					Volver al inicio
				</button>
			</section>
		{:else}
			<div class="flex flex-wrap items-end justify-between gap-4">
				<div>
					<p class="text-xs font-extrabold tracking-[0.26em] text-accent uppercase">
						Administración
					</p>
					<h1 class="mt-2 font-display text-3xl font-bold text-primary sm:text-4xl">Usuarios</h1>
					<p class="mt-2 text-sm text-secondary">Registra únicamente correo y nombre visible.</p>
				</div>
				<button
					class="rounded-full border border-border-subtle bg-panel px-4 py-2 text-sm font-bold text-primary shadow-card"
					type="button"
					on:click={() => onNavigate('/')}
				>
					Volver al inicio
				</button>
			</div>

			<form
				class="mt-6 grid gap-4 rounded-[28px] border border-border-subtle bg-panel p-4 shadow-card sm:p-5 lg:grid-cols-[1fr_1fr_auto]"
				on:submit|preventDefault={submitUser}
			>
				<label class="flex flex-col gap-2">
					<span class="text-xs font-extrabold tracking-[0.18em] text-muted uppercase">Correo</span>
					<input
						class="rounded-[16px] border border-border-subtle bg-surface px-4 py-3 text-sm text-primary transition outline-none focus:border-accent focus:ring-4 focus:ring-accent-soft"
						bind:value={email}
						autocomplete="email"
						inputmode="email"
						placeholder="persona@correo.com"
						required
						type="email"
					/>
				</label>

				<label class="flex flex-col gap-2">
					<span class="text-xs font-extrabold tracking-[0.18em] text-muted uppercase">Nombre</span>
					<input
						class="rounded-[16px] border border-border-subtle bg-surface px-4 py-3 text-sm text-primary transition outline-none focus:border-accent focus:ring-4 focus:ring-accent-soft"
						bind:value={displayName}
						autocomplete="name"
						placeholder="Nombre Apellido"
						required
						type="text"
					/>
				</label>

				<button
					class="self-end rounded-[16px] bg-accent-strong px-5 py-3 text-sm font-bold text-white transition hover:bg-accent disabled:cursor-not-allowed disabled:opacity-60"
					disabled={isSubmitting}
					type="submit"
				>
					{isSubmitting ? 'Registrando…' : 'Registrar'}
				</button>
			</form>

			{#if errorMessage}
				<p class="mt-4 rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning">
					{errorMessage}
				</p>
			{/if}
			{#if successMessage}
				<p class="mt-4 rounded-2xl bg-success/10 px-4 py-3 text-sm font-semibold text-success">
					{successMessage}
				</p>
			{/if}

			<section
				class="mt-6 overflow-hidden rounded-[28px] border border-border-subtle bg-panel shadow-panel"
			>
				<div class="flex items-center justify-between border-b border-border-subtle px-5 py-4">
					<h2 class="font-display text-xl font-bold text-primary">Registrados</h2>
					<span class="text-sm font-bold text-secondary">{users.length}</span>
				</div>

				{#if isLoading}
					<p class="p-6 text-secondary">Cargando usuarios…</p>
				{:else if users.length === 0}
					<p class="p-6 text-secondary">Aún no hay usuarios registrados.</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full min-w-152 text-left text-sm">
							<thead class="bg-surface-muted text-xs tracking-[0.16em] text-muted uppercase">
								<tr>
									<th class="px-5 py-3">Correo</th>
									<th class="px-5 py-3">Nombre</th>
									<th class="px-5 py-3">Rol</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-grid">
								{#each users as registeredUser (registeredUser.email)}
									<tr>
										<td class="px-5 py-4 font-semibold text-primary">
											{registeredUser.email}
										</td>
										<td class="px-5 py-4 text-secondary">
											{registeredUser.displayName}
										</td>
										<td class="px-5 py-4">
											<span
												class={`rounded-full px-2.5 py-1 text-xs font-extrabold ${
													registeredUser.role === 'ADMIN'
														? 'bg-accent-soft text-accent'
														: 'bg-surface-muted text-secondary'
												}`}
											>
												{registeredUser.role}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</section>
		{/if}
	</main>
</div>
