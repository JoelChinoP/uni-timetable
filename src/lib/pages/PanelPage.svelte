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
		updateClassroom,
		updateCourse,
		updateGroup,
		updateTeacher,
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

	type PanelTab = 'courses' | 'classrooms' | 'teachers' | 'users';
	type DeleteTarget = { label: string; success: string; action: () => Promise<unknown> };

	let activeTab: PanelTab = 'courses';
	let data: PlannerData | null = null;
	let classrooms: ClassroomItem[] = [];
	let teachers: TeacherItem[] = [];
	let users: AuthUser[] = [];
	let loadError = '';
	let actionMessage = '';
	let busy = false;
	let modalError = '';

	let showCourseForm = false;
	let editingCourse: Course | null = null;
	let groupCourse: Course | null = null;
	let editingGroup: CourseGroup | null = null;
	let showClassroomForm = false;
	let editingClassroom: ClassroomItem | null = null;
	let showTeacherForm = false;
	let editingTeacher: TeacherItem | null = null;
	let deleteTarget: DeleteTarget | null = null;

	let classroomCode = '';
	let classroomType: 'THEORY' | 'LABORATORY' = 'THEORY';
	let classroomFloor: number | null = null;
	let classroomCapacity: number | null = null;
	let teacherName = '';
	let teacherLastName = '';
	let userEmail = '';
	let userName = '';

	$: isAdmin = user?.role === 'ADMIN';
	$: registered = !!user && user.id !== 0;
	$: tabs = [
		{ id: 'courses' as const, label: 'Cursos' },
		{ id: 'classrooms' as const, label: 'Aulas' },
		{ id: 'teachers' as const, label: 'Docentes' },
		...(isAdmin ? [{ id: 'users' as const, label: 'Usuarios' }] : []),
	];

	onMount(() => {
		if (registered) void refreshAll();
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
			if (isAdmin) users = await loadUsers();
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'No se pudo cargar el panel';
		}
	}

	async function run(action: () => Promise<unknown>, success: string) {
		busy = true;
		actionMessage = '';
		try {
			await action();
			await refreshAll();
			actionMessage = success;
			return true;
		} catch (error) {
			actionMessage = error instanceof Error ? error.message : 'La acción falló';
			return false;
		} finally {
			busy = false;
		}
	}

	async function saveCourse(payload: CoursePayload) {
		modalError = '';
		busy = true;
		try {
			if (editingCourse) await updateCourse(editingCourse.id, payload);
			else await createCourse(payload);
			showCourseForm = false;
			editingCourse = null;
			await refreshAll();
			actionMessage = 'Curso guardado.';
		} catch (error) {
			modalError = error instanceof Error ? error.message : 'No se pudo guardar el curso';
		} finally {
			busy = false;
		}
	}

	async function saveGroup(payload: GroupPayload) {
		modalError = '';
		busy = true;
		try {
			if (editingGroup) await updateGroup(editingGroup.id, payload);
			else await createGroup(payload);
			groupCourse = null;
			editingGroup = null;
			await refreshAll();
			actionMessage = 'Grupo y horarios guardados.';
		} catch (error) {
			modalError = error instanceof Error ? error.message : 'No se pudo guardar el grupo';
		} finally {
			busy = false;
		}
	}

	function openClassroom(classroom: ClassroomItem | null = null) {
		editingClassroom = classroom;
		classroomCode = classroom?.code ?? '';
		classroomType = classroom?.type ?? 'THEORY';
		classroomFloor = classroom?.floor ?? null;
		classroomCapacity = classroom?.capacity ?? null;
		showClassroomForm = true;
	}

	async function saveClassroom() {
		const payload = {
			code: classroomCode,
			type: classroomType,
			floor: classroomFloor,
			capacity: classroomCapacity,
		};
		const ok = await run(
			() =>
				editingClassroom ? updateClassroom(editingClassroom.id, payload) : createClassroom(payload),
			'Aula guardada.',
		);
		if (ok) {
			showClassroomForm = false;
			editingClassroom = null;
		}
	}

	function openTeacher(teacher: TeacherItem | null = null) {
		editingTeacher = teacher;
		teacherName = teacher?.name ?? '';
		teacherLastName = teacher?.lastName ?? '';
		showTeacherForm = true;
	}

	async function saveTeacher() {
		const payload = { name: teacherName.trim(), lastName: teacherLastName.trim() };
		const ok = await run(
			() => (editingTeacher ? updateTeacher(editingTeacher.id, payload) : createTeacher(payload)),
			'Docente guardado.',
		);
		if (ok) {
			showTeacherForm = false;
			editingTeacher = null;
		}
	}

	async function submitUser() {
		const ok = await run(
			() => createUser(userEmail.trim().toLowerCase(), userName.trim()),
			'Usuario registrado.',
		);
		if (ok) {
			userEmail = '';
			userName = '';
		}
	}

	function askDelete(label: string, action: () => Promise<unknown>, success: string) {
		deleteTarget = { label, action, success };
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		const target = deleteTarget;
		if (await run(target.action, target.success)) deleteTarget = null;
	}

	const fieldClass = 'neo-control min-h-11 w-full px-3 py-2 text-sm text-primary outline-none';
	const labelClass = 'text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase';
</script>

<svelte:head><title>Panel académico | Horarios</title></svelte:head>

<div class="min-h-dvh">
	<TopBar {user} {onNavigate} {onLogout} />
	<main id="main-content" class="mx-auto w-full max-w-7xl p-3 sm:p-5 lg:p-6">
		{#if !user}
			<section class="neo-panel p-8 text-center">
				<h1 class="text-3xl font-extrabold text-primary">Necesitas iniciar sesión</h1>
				<p class="mt-3 text-secondary">Entra con Google para gestionar los horarios.</p>
				<button
					class="mt-6 min-h-11 rounded-xl bg-accent-strong px-5 font-bold text-white"
					type="button"
					on:click={() => onNavigate('/login')}>Iniciar sesión</button
				>
			</section>
		{:else if !registered}
			<section class="neo-panel p-8 text-center">
				<h1 class="text-3xl font-extrabold text-warning">Cuenta no registrada</h1>
				<p class="mt-3 text-secondary">
					Un administrador debe registrar tu correo antes de editar el catálogo.
				</p>
				<button
					class="neo-button mt-6 min-h-11 px-5 font-bold text-primary"
					type="button"
					on:click={() => onNavigate('/')}>Volver al dashboard</button
				>
			</section>
		{:else}
			<header class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<p class="text-[10px] font-extrabold tracking-[0.2em] text-accent uppercase">
						{data?.termLabel ?? 'Periodo académico'}
					</p>
					<h1 class="mt-1 text-3xl font-extrabold tracking-tight text-primary">
						Gestión académica
					</h1>
				</div>
				<button
					class="neo-button min-h-11 px-4 text-sm font-bold text-primary"
					type="button"
					on:click={refreshAll}>Actualizar datos</button
				>
			</header>

			{#if loadError}<div
					class="mt-4 flex items-center justify-between gap-3 rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning"
					role="alert"
				>
					<span>{loadError}</span><button
						class="min-h-10 px-2 font-bold"
						type="button"
						on:click={refreshAll}>Reintentar</button
					>
				</div>{/if}
			{#if actionMessage}<p
					class="mt-4 rounded-2xl bg-success-soft px-4 py-3 text-sm font-semibold text-success"
					role="status"
					aria-live="polite"
				>
					{actionMessage}
				</p>{/if}

			<div
				class="neo-control mt-5 flex gap-1 overflow-x-auto p-1"
				role="tablist"
				aria-label="Áreas de gestión"
			>
				{#each tabs as tab (tab.id)}<button
						class="min-h-11 shrink-0 rounded-[10px] px-4 text-sm font-bold text-secondary"
						class:bg-surface={activeTab === tab.id}
						class:shadow-card={activeTab === tab.id}
						class:text-primary={activeTab === tab.id}
						type="button"
						role="tab"
						aria-selected={activeTab === tab.id}
						on:click={() => (activeTab = tab.id)}>{tab.label}</button
					>{/each}
			</div>

			<section class="neo-panel mt-4 overflow-hidden" aria-live="polite">
				{#if activeTab === 'courses'}
					<header
						class="flex items-center justify-between gap-3 border-b border-border-subtle px-4 py-3 sm:px-5"
					>
						<div>
							<h2 class="text-lg font-extrabold text-primary">Cursos, laboratorios y grupos</h2>
							<p class="mt-0.5 text-xs text-secondary">
								Selecciona un curso para administrar sus horarios.
							</p>
						</div>
						<button
							class="min-h-11 rounded-xl bg-accent-strong px-4 text-xs font-bold text-white disabled:opacity-50"
							type="button"
							disabled={busy}
							on:click={() => {
								modalError = '';
								editingCourse = null;
								showCourseForm = true;
							}}>Nuevo curso</button
						>
					</header>
					{#if data}<CourseCatalog
							courses={data.courses}
							academicHours={data.academicHours}
							onAddGroup={(course) => {
								modalError = '';
								groupCourse = course;
								editingGroup = null;
							}}
							onEditCourse={(course) => {
								modalError = '';
								editingCourse = course;
								showCourseForm = true;
							}}
							onEditGroup={(group, course) => {
								modalError = '';
								groupCourse = course;
								editingGroup = group;
							}}
							onDeleteGroup={(group: CourseGroup, course: Course) =>
								askDelete(
									`el grupo ${group.name} de ${course.name}`,
									() => deleteGroup(group.id),
									'Grupo eliminado.',
								)}
							onDeleteCourse={(course: Course) =>
								askDelete(course.name, () => deleteCourse(course.id), 'Curso eliminado.')}
						/>{:else}<p class="p-6 text-sm text-secondary">Cargando cursos…</p>{/if}
				{:else if activeTab === 'classrooms'}
					<header
						class="flex items-center justify-between border-b border-border-subtle px-4 py-3 sm:px-5"
					>
						<div>
							<h2 class="text-lg font-extrabold text-primary">Aulas y laboratorios</h2>
							<p class="mt-0.5 text-xs text-secondary">Capacidad, ubicación y modalidad.</p>
						</div>
						<button
							class="min-h-11 rounded-xl bg-accent-strong px-4 text-xs font-bold text-white"
							type="button"
							on:click={() => openClassroom()}>Nueva aula</button
						>
					</header>
					<div class="grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-3">
						{#each classrooms as classroom (classroom.id)}<article class="neo-card p-4">
								<div class="flex items-start justify-between gap-3">
									<div>
										<h3 class="font-extrabold text-primary">{classroom.code}</h3>
										<p class="mt-1 text-xs text-secondary">
											{classroom.type === 'THEORY' ? 'Aula de teoría' : 'Laboratorio'}
										</p>
									</div>
									<span
										class="rounded-lg bg-surface-muted px-2 py-1 text-[10px] font-bold text-secondary"
										>{classroom.capacity ? `${classroom.capacity} pers.` : 'Sin aforo'}</span
									>
								</div>
								<p class="mt-3 text-xs text-muted">
									{classroom.floor === null ? 'Piso no indicado' : `Piso ${classroom.floor}`}
								</p>
								<div class="mt-4 flex gap-2">
									<button
										class="neo-button min-h-10 flex-1 text-xs font-bold text-primary"
										type="button"
										on:click={() => openClassroom(classroom)}>Editar</button
									><button
										class="min-h-10 rounded-xl px-3 text-xs font-bold text-warning hover:bg-warning-soft"
										type="button"
										on:click={() =>
											askDelete(
												classroom.code,
												() => deleteClassroom(classroom.id),
												'Aula eliminada.',
											)}>Eliminar</button
									>
								</div>
							</article>{/each}
					</div>
				{:else if activeTab === 'teachers'}
					<header
						class="flex items-center justify-between border-b border-border-subtle px-4 py-3 sm:px-5"
					>
						<div>
							<h2 class="text-lg font-extrabold text-primary">Docentes</h2>
							<p class="mt-0.5 text-xs text-secondary">Personas vinculadas a un curso.</p>
						</div>
						<button
							class="min-h-11 rounded-xl bg-accent-strong px-4 text-xs font-bold text-white"
							type="button"
							on:click={() => openTeacher()}>Nuevo docente</button
						>
					</header>
					<div class="grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-3">
						{#each teachers as teacher (teacher.id)}<article
								class="neo-card flex items-center gap-3 p-4"
							>
								<span
									class="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-accent-soft text-sm font-extrabold text-accent"
									>{teacher.name.charAt(0)}{teacher.lastName.charAt(0)}</span
								>
								<div class="min-w-0 flex-1">
									<h3 class="truncate text-sm font-extrabold text-primary">{teacher.fullName}</h3>
									<div class="mt-2 flex gap-2">
										<button
											class="text-xs font-bold text-accent"
											type="button"
											on:click={() => openTeacher(teacher)}>Editar</button
										><button
											class="text-xs font-bold text-warning"
											type="button"
											on:click={() =>
												askDelete(
													teacher.fullName,
													() => deleteTeacher(teacher.id),
													'Docente eliminado.',
												)}>Eliminar</button
										>
									</div>
								</div>
							</article>{/each}
					</div>
				{:else if activeTab === 'users' && isAdmin}
					<header class="border-b border-border-subtle px-4 py-3 sm:px-5">
						<h2 class="text-lg font-extrabold text-primary">Usuarios autorizados</h2>
						<p class="mt-0.5 text-xs text-secondary">
							Registra las cuentas que pueden editar los horarios.
						</p>
					</header>
					<div class="p-4 sm:p-5">
						<form
							class="grid gap-3 md:grid-cols-[1fr_1fr_auto]"
							on:submit|preventDefault={submitUser}
						>
							<label class="flex flex-col gap-1.5"
								><span class={labelClass}>Correo</span><input
									class={fieldClass}
									bind:value={userEmail}
									autocomplete="email"
									inputmode="email"
									required
									type="email"
								/></label
							><label class="flex flex-col gap-1.5"
								><span class={labelClass}>Nombre</span><input
									class={fieldClass}
									bind:value={userName}
									autocomplete="name"
									required
									type="text"
								/></label
							><button
								class="min-h-11 self-end rounded-xl bg-accent-strong px-5 text-sm font-bold text-white disabled:opacity-50"
								type="submit"
								disabled={busy}>{busy ? 'Registrando…' : 'Registrar'}</button
							>
						</form>
						<div class="mt-5 overflow-x-auto">
							<table class="w-full min-w-152 text-left text-sm">
								<thead class="bg-surface-muted text-[10px] tracking-[0.14em] text-muted uppercase"
									><tr
										><th class="px-3 py-3">Correo</th><th class="px-3 py-3">Nombre</th><th
											class="px-3 py-3">Rol</th
										></tr
									></thead
								><tbody class="divide-y divide-grid"
									>{#each users as registeredUser (registeredUser.email)}<tr
											><td class="px-3 py-3 font-semibold text-primary">{registeredUser.email}</td
											><td class="px-3 py-3 text-secondary">{registeredUser.displayName}</td><td
												class="px-3 py-3 text-secondary">{registeredUser.role}</td
											></tr
										>{/each}</tbody
								>
							</table>
						</div>
					</div>
				{/if}
			</section>
		{/if}
	</main>

	{#if showCourseForm}<CourseFormModal
			courses={data?.courses ?? []}
			{teachers}
			{busy}
			initialCourse={editingCourse}
			errorMessage={modalError}
			onSave={saveCourse}
			onClose={() => {
				showCourseForm = false;
				editingCourse = null;
			}}
		/>{/if}
	{#if groupCourse}<GroupFormModal
			course={groupCourse}
			{classrooms}
			academicHours={data?.academicHours ?? []}
			{busy}
			initialGroup={editingGroup}
			errorMessage={modalError}
			onSave={saveGroup}
			onClose={() => {
				groupCourse = null;
				editingGroup = null;
			}}
		/>{/if}

	{#if showClassroomForm}<Modal
			title={editingClassroom ? 'Editar aula' : 'Nueva aula'}
			onClose={() => (showClassroomForm = false)}
			><form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={saveClassroom}>
				<label class="flex flex-col gap-1.5"
					><span class={labelClass}>Código</span><input
						class={fieldClass}
						bind:value={classroomCode}
						required
						type="text"
					/></label
				><label class="flex flex-col gap-1.5"
					><span class={labelClass}>Modalidad</span><select
						class={fieldClass}
						bind:value={classroomType}
						disabled={!!editingClassroom}
						><option value="THEORY">Teoría</option><option value="LABORATORY">Laboratorio</option
						></select
					></label
				><label class="flex flex-col gap-1.5"
					><span class={labelClass}>Piso</span><input
						class={fieldClass}
						bind:value={classroomFloor}
						inputmode="numeric"
						min="0"
						max="40"
						type="number"
					/></label
				><label class="flex flex-col gap-1.5"
					><span class={labelClass}>Capacidad</span><input
						class={fieldClass}
						bind:value={classroomCapacity}
						inputmode="numeric"
						min="1"
						type="number"
					/></label
				><button
					class="min-h-12 rounded-xl bg-accent-strong px-4 text-sm font-bold text-white disabled:opacity-50 sm:col-span-2"
					disabled={busy}
					type="submit">{busy ? 'Guardando…' : 'Guardar aula'}</button
				>
			</form></Modal
		>{/if}

	{#if showTeacherForm}<Modal
			title={editingTeacher ? 'Editar docente' : 'Nuevo docente'}
			onClose={() => (showTeacherForm = false)}
			><form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={saveTeacher}>
				<label class="flex flex-col gap-1.5"
					><span class={labelClass}>Nombres</span><input
						class={fieldClass}
						bind:value={teacherName}
						autocomplete="given-name"
						required
						type="text"
					/></label
				><label class="flex flex-col gap-1.5"
					><span class={labelClass}>Apellidos</span><input
						class={fieldClass}
						bind:value={teacherLastName}
						autocomplete="family-name"
						required
						type="text"
					/></label
				><button
					class="min-h-12 rounded-xl bg-accent-strong px-4 text-sm font-bold text-white disabled:opacity-50 sm:col-span-2"
					disabled={busy}
					type="submit">{busy ? 'Guardando…' : 'Guardar docente'}</button
				>
			</form></Modal
		>{/if}

	{#if deleteTarget}<Modal title="Confirmar eliminación" onClose={() => (deleteTarget = null)}
			><div>
				<p class="text-sm leading-6 text-secondary">
					Vas a eliminar <strong class="text-primary">{deleteTarget.label}</strong>. Esta acción no
					se puede deshacer.
				</p>
				<div class="mt-5 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<button
						class="neo-button min-h-11 px-4 text-sm font-bold text-primary"
						type="button"
						on:click={() => (deleteTarget = null)}>Cancelar</button
					><button
						class="min-h-11 rounded-xl bg-warning px-4 text-sm font-bold text-white disabled:opacity-50"
						type="button"
						disabled={busy}
						on:click={confirmDelete}>{busy ? 'Eliminando…' : 'Eliminar'}</button
					>
				</div>
			</div></Modal
		>{/if}
</div>
