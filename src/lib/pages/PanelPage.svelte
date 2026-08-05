<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import Modal from '../components/panel/Modal.svelte';
	import CourseCatalog from '../components/panel/CourseCatalog.svelte';
	import GroupFormModal from '../components/panel/GroupFormModal.svelte';
	import CourseFormModal from '../components/panel/CourseFormModal.svelte';
	import { createUser, loadUsers } from '../api/auth';
	import { getDashboard } from '../api/planner';
	import {
		createClassroom,
		createCourse,
		createGroup,
		createTeacher,
		deleteClassroom,
		deleteCourse,
		deleteGroup,
		deleteTeacher,
		getClassrooms,
		getTeachers,
		type ClassroomItem,
		type CoursePayload,
		type GroupPayload,
		type TeacherItem,
	} from '../api/catalog';
	import type { Course, CourseGroup, PlannerData } from '../types/planner';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let classrooms: ClassroomItem[] = [];
	let teachers: TeacherItem[] = [];
	let users: AuthUser[] = [];
	let loadError = '';
	let actionMessage = '';
	let busy = false;
	let modalError = '';

	// Estado de modales; uno a la vez basta en un panel simple.
	let groupCourse: Course | null = null;
	let showCourseForm = false;
	let showClassroomForm = false;
	let showTeacherForm = false;

	// Formulario de usuarios (ADMIN)
	let userEmail = '';
	let userName = '';
	let userBusy = false;
	let userError = '';
	let userSuccess = '';

	// Formularios simples
	let classroomCode = '';
	let classroomType: 'THEORY' | 'LABORATORY' = 'THEORY';
	let classroomFloor: number | null = null;
	let classroomCapacity: number | null = null;
	let teacherName = '';
	let teacherLastName = '';

	$: isAdmin = user?.role === 'ADMIN';
	$: registered = !!user && user.id !== 0;

	onMount(() => {
		if (registered) {
			void refreshAll();
		}
	});

	async function refreshAll() {
		loadError = '';
		try {
			const [dashboard, nextClassrooms, nextTeachers] = await Promise.all([
				getDashboard(),
				getClassrooms(),
				getTeachers(),
			]);
			data = dashboard;
			classrooms = nextClassrooms;
			teachers = nextTeachers;
			if (isAdmin) {
				users = await loadUsers();
			}
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'No se pudo cargar el panel';
		}
	}

	function confirmDelete(message: string) {
		return window.confirm(message);
	}

	async function run(action: () => Promise<unknown>, success: string) {
		busy = true;
		actionMessage = '';
		try {
			await action();
			actionMessage = success;
			await refreshAll();
			return true;
		} catch (error) {
			actionMessage = error instanceof Error ? error.message : 'La acción falló';
			return false;
		} finally {
			busy = false;
		}
	}

	async function saveGroup(payload: GroupPayload) {
		modalError = '';
		busy = true;
		try {
			await createGroup(payload);
			groupCourse = null;
			await run(async () => {}, 'Grupo registrado.');
		} catch (error) {
			modalError = error instanceof Error ? error.message : 'No se pudo registrar el grupo';
		} finally {
			busy = false;
		}
	}

	async function saveCourse(payload: CoursePayload) {
		modalError = '';
		busy = true;
		try {
			await createCourse(payload);
			showCourseForm = false;
			await run(async () => {}, 'Curso registrado.');
		} catch (error) {
			modalError = error instanceof Error ? error.message : 'No se pudo registrar el curso';
		} finally {
			busy = false;
		}
	}

	function askDeleteGroup(group: CourseGroup, course: Course) {
		if (!confirmDelete(`¿Eliminar el grupo ${group.name} de ${course.name} con sus horarios?`)) {
			return;
		}
		void run(() => deleteGroup(group.id), 'Grupo eliminado.');
	}

	function askDeleteCourse(course: Course) {
		if (!confirmDelete(`¿Eliminar ${course.name} con todos sus grupos?`)) {
			return;
		}
		void run(() => deleteCourse(course.id), 'Curso eliminado.');
	}

	async function saveClassroom() {
		const ok = await run(
			() =>
				createClassroom({
					code: classroomCode,
					type: classroomType,
					floor: classroomFloor,
					capacity: classroomCapacity,
				}),
			'Aula registrada.',
		);
		if (ok) {
			showClassroomForm = false;
			classroomCode = '';
			classroomFloor = null;
			classroomCapacity = null;
		}
	}

	async function saveTeacher() {
		const ok = await run(
			() => createTeacher({ name: teacherName.trim(), lastName: teacherLastName.trim() }),
			'Docente registrado.',
		);
		if (ok) {
			showTeacherForm = false;
			teacherName = '';
			teacherLastName = '';
		}
	}

	async function submitUser() {
		userError = '';
		userSuccess = '';
		userBusy = true;
		try {
			await createUser(userEmail.trim().toLowerCase(), userName.trim());
			userEmail = '';
			userName = '';
			userSuccess = 'Usuario registrado correctamente.';
			users = await loadUsers();
		} catch (error) {
			userError = error instanceof Error ? error.message : 'No se pudo registrar el usuario';
		} finally {
			userBusy = false;
		}
	}

	const fieldClass =
		'w-full rounded-[12px] border border-border-subtle bg-surface px-3 py-2 text-sm text-primary outline-none transition focus:border-accent focus:ring-4 focus:ring-accent-soft';
	const labelClass = 'text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase';
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
		{:else if !registered}
			<section
				class="rounded-[28px] border border-warning/25 bg-warning-soft p-8 text-center shadow-panel"
			>
				<h1 class="font-display text-3xl font-bold text-warning">Cuenta no registrada</h1>
				<p class="mt-3 text-secondary">
					Un administrador debe registrar tu correo para que puedas gestionar el catálogo.
				</p>
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
					<p class="text-xs font-extrabold tracking-[0.26em] text-accent uppercase">Panel</p>
					<h1 class="mt-2 font-display text-3xl font-bold text-primary sm:text-4xl">
						Gestión académica
					</h1>
					<p class="mt-2 text-sm text-secondary">
						Cursos, grupos, aulas y docentes de {data?.termLabel ?? 'este periodo'}.
					</p>
				</div>
				<button
					class="rounded-full border border-border-subtle bg-panel px-4 py-2 text-sm font-bold text-primary shadow-card"
					type="button"
					on:click={() => onNavigate('/')}
				>
					Volver al tablero
				</button>
			</div>

			{#if loadError}
				<p class="mt-4 rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning">
					{loadError}
				</p>
			{/if}
			{#if actionMessage}
				<p class="mt-4 rounded-2xl bg-accent-soft px-4 py-3 text-sm font-semibold text-accent">
					{actionMessage}
				</p>
			{/if}

			<div class="mt-6 grid items-start gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
				<section
					class="overflow-hidden rounded-[24px] border border-border-subtle bg-panel shadow-panel"
				>
					<div class="flex items-center justify-between border-b border-border-subtle px-5 py-4">
						<h2 class="font-display text-xl font-bold text-primary">Cursos</h2>
						<button
							class="rounded-full bg-accent-strong px-4 py-2 text-xs font-bold text-white transition hover:bg-accent disabled:opacity-60"
							type="button"
							disabled={busy}
							on:click={() => {
								modalError = '';
								showCourseForm = true;
							}}
						>
							+ Curso
						</button>
					</div>
					{#if !data}
						<p class="p-6 text-secondary">Cargando cursos…</p>
					{:else}
						<CourseCatalog
							courses={data.courses}
							academicHours={data.academicHours}
							onAddGroup={(course) => {
								modalError = '';
								groupCourse = course;
							}}
							onDeleteGroup={askDeleteGroup}
							onDeleteCourse={askDeleteCourse}
						/>
					{/if}
				</section>

				<div class="grid gap-4">
					<section
						class="overflow-hidden rounded-[24px] border border-border-subtle bg-panel shadow-panel"
					>
						<div class="flex items-center justify-between border-b border-border-subtle px-5 py-4">
							<h2 class="font-display text-lg font-bold text-primary">Aulas</h2>
							<button
								class="rounded-full bg-accent-strong px-3 py-1.5 text-xs font-bold text-white transition hover:bg-accent"
								type="button"
								on:click={() => (showClassroomForm = true)}
							>
								+ Aula
							</button>
						</div>
						<div class="max-h-72 overflow-y-auto p-3">
							{#each classrooms as classroom (classroom.id)}
								<div
									class="flex items-center justify-between gap-2 rounded-[12px] px-2 py-1.5 text-sm hover:bg-surface-muted"
								>
									<span class="text-primary">
										{classroom.code}
										<span class="text-[10px] font-bold text-muted uppercase">
											· {classroom.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}
										</span>
									</span>
									<button
										class="rounded-lg px-2 py-1 text-[11px] font-bold text-warning transition hover:bg-warning-soft"
										type="button"
										on:click={() => {
											if (confirmDelete(`¿Eliminar ${classroom.code}?`)) {
												void run(() => deleteClassroom(classroom.id), 'Aula eliminada.');
											}
										}}
									>
										Eliminar
									</button>
								</div>
							{/each}
						</div>
					</section>

					<section
						class="overflow-hidden rounded-[24px] border border-border-subtle bg-panel shadow-panel"
					>
						<div class="flex items-center justify-between border-b border-border-subtle px-5 py-4">
							<h2 class="font-display text-lg font-bold text-primary">Docentes</h2>
							<button
								class="rounded-full bg-accent-strong px-3 py-1.5 text-xs font-bold text-white transition hover:bg-accent"
								type="button"
								on:click={() => (showTeacherForm = true)}
							>
								+ Docente
							</button>
						</div>
						<div class="max-h-72 overflow-y-auto p-3">
							{#if teachers.length === 0}
								<p class="px-2 py-1.5 text-xs text-secondary">Sin docentes registrados.</p>
							{/if}
							{#each teachers as teacher (teacher.id)}
								<div
									class="flex items-center justify-between gap-2 rounded-[12px] px-2 py-1.5 text-sm hover:bg-surface-muted"
								>
									<span class="text-primary">{teacher.fullName}</span>
									<button
										class="rounded-lg px-2 py-1 text-[11px] font-bold text-warning transition hover:bg-warning-soft"
										type="button"
										on:click={() => {
											if (confirmDelete(`¿Eliminar a ${teacher.fullName}?`)) {
												void run(() => deleteTeacher(teacher.id), 'Docente eliminado.');
											}
										}}
									>
										Eliminar
									</button>
								</div>
							{/each}
						</div>
					</section>
				</div>
			</div>

			{#if isAdmin}
				<section
					class="mt-4 overflow-hidden rounded-[24px] border border-border-subtle bg-panel shadow-panel"
				>
					<div class="border-b border-border-subtle px-5 py-4">
						<h2 class="font-display text-xl font-bold text-primary">Usuarios</h2>
						<p class="mt-1 text-sm text-secondary">
							Registra únicamente correo y nombre visible. Solo administradores.
						</p>
					</div>
					<div class="p-5">
						<form
							class="grid gap-4 lg:grid-cols-[1fr_1fr_auto]"
							on:submit|preventDefault={submitUser}
						>
							<label class="flex flex-col gap-2">
								<span class={labelClass}>Correo</span>
								<input
									class={fieldClass}
									bind:value={userEmail}
									autocomplete="email"
									inputmode="email"
									placeholder="persona@correo.com"
									required
									type="email"
								/>
							</label>
							<label class="flex flex-col gap-2">
								<span class={labelClass}>Nombre</span>
								<input
									class={fieldClass}
									bind:value={userName}
									autocomplete="name"
									placeholder="Nombre Apellido"
									required
									type="text"
								/>
							</label>
							<button
								class="self-end rounded-[12px] bg-accent-strong px-5 py-2.5 text-sm font-bold text-white transition hover:bg-accent disabled:cursor-not-allowed disabled:opacity-60"
								disabled={userBusy}
								type="submit"
							>
								{userBusy ? 'Registrando…' : 'Registrar'}
							</button>
						</form>

						{#if userError}
							<p
								class="mt-4 rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning"
							>
								{userError}
							</p>
						{/if}
						{#if userSuccess}
							<p
								class="mt-4 rounded-2xl bg-success/10 px-4 py-3 text-sm font-semibold text-success"
							>
								{userSuccess}
							</p>
						{/if}

						<div class="mt-5 overflow-x-auto">
							<table class="w-full min-w-152 text-left text-sm">
								<thead class="bg-surface-muted text-xs tracking-[0.16em] text-muted uppercase">
									<tr>
										<th class="px-4 py-2.5">Correo</th>
										<th class="px-4 py-2.5">Nombre</th>
										<th class="px-4 py-2.5">Rol</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-grid">
									{#each users as registeredUser (registeredUser.email)}
										<tr>
											<td class="px-4 py-3 font-semibold text-primary">
												{registeredUser.email}
											</td>
											<td class="px-4 py-3 text-secondary">{registeredUser.displayName}</td>
											<td class="px-4 py-3">
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
					</div>
				</section>
			{/if}
		{/if}
	</main>

	{#if groupCourse}
		<GroupFormModal
			course={groupCourse}
			{classrooms}
			academicHours={data?.academicHours ?? []}
			{busy}
			errorMessage={modalError}
			onSave={saveGroup}
			onClose={() => (groupCourse = null)}
		/>
	{/if}

	{#if showCourseForm}
		<CourseFormModal
			courses={data?.courses ?? []}
			{teachers}
			{busy}
			errorMessage={modalError}
			onSave={saveCourse}
			onClose={() => (showCourseForm = false)}
		/>
	{/if}

	{#if showClassroomForm}
		<Modal title="Nueva aula" onClose={() => (showClassroomForm = false)}>
			<form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={saveClassroom}>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Código</span>
					<input
						class={fieldClass}
						bind:value={classroomCode}
						placeholder="Aula 306"
						required
						type="text"
					/>
				</label>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Tipo</span>
					<select class={fieldClass} bind:value={classroomType}>
						<option value="THEORY">Teoría</option>
						<option value="LABORATORY">Laboratorio</option>
					</select>
				</label>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Piso (opcional)</span>
					<input class={fieldClass} bind:value={classroomFloor} min="0" max="40" type="number" />
				</label>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Capacidad (opcional)</span>
					<input class={fieldClass} bind:value={classroomCapacity} min="1" type="number" />
				</label>
				<button
					class="rounded-[14px] bg-accent-strong px-4 py-3 text-sm font-bold text-white transition hover:bg-accent disabled:opacity-60 sm:col-span-2"
					disabled={busy}
					type="submit"
				>
					{busy ? 'Guardando…' : 'Guardar aula'}
				</button>
			</form>
		</Modal>
	{/if}

	{#if showTeacherForm}
		<Modal title="Nuevo docente" onClose={() => (showTeacherForm = false)}>
			<form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={saveTeacher}>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Nombres</span>
					<input class={fieldClass} bind:value={teacherName} required type="text" />
				</label>
				<label class="flex flex-col gap-1.5">
					<span class={labelClass}>Apellidos</span>
					<input class={fieldClass} bind:value={teacherLastName} required type="text" />
				</label>
				<button
					class="rounded-[14px] bg-accent-strong px-4 py-3 text-sm font-bold text-white transition hover:bg-accent disabled:opacity-60 sm:col-span-2"
					disabled={busy}
					type="submit"
				>
					{busy ? 'Guardando…' : 'Guardar docente'}
				</button>
			</form>
		</Modal>
	{/if}
</div>
