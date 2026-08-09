<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import { getClassrooms, type ClassroomItem } from '../api/catalog';
	import { getDashboard } from '../api/planner';
	import type { AuthUser } from '../types/auth';
	import type { Course, PlannerData, PlannerEvent } from '../types/planner';
	import {
		buildPlannerEvents,
		getClassroomCourseGroups,
		getCourseDisplayCode,
	} from '../utils/planner';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let classrooms: ClassroomItem[] = [];
	let selectedClassroomId: number | null = null;
	let selectedYear: number | null = null;
	let loadError = '';
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;
	let events: PlannerEvent[] = [];

	onMount(async () => {
		loadError = '';
		try {
			const [dashboard, nextClassrooms] = await Promise.all([getDashboard(), getClassrooms()]);
			data = dashboard;
			classrooms = nextClassrooms;
			const requestedId = Number(new URLSearchParams(window.location.search).get('classroom'));
			selectedClassroomId = nextClassrooms.some(({ id }) => id === requestedId)
				? requestedId
				: (nextClassrooms[0]?.id ?? null);
		} catch (error) {
			loadError = error instanceof Error ? error.message : 'No se pudieron cargar las aulas';
		}
	});

	$: allCourses = data?.courses ?? [];
	$: availableYears = [...new Set(allCourses.map((course) => course.academicYear))].sort();
	$: filteredCourses = selectedYear
		? allCourses.filter((course) => course.academicYear === selectedYear)
		: allCourses;
	$: selectedClassroom = classrooms.find(({ id }) => id === selectedClassroomId) ?? null;
	$: classroomGroups = selectedClassroomId
		? getClassroomCourseGroups(filteredCourses, selectedClassroomId)
		: [];
	$: classroomCourseCount = new Set(classroomGroups.map(({ course }) => course.id)).size;
	$: ({ events } = buildPlannerEvents(classroomGroups));
	$: selectedGroups = Object.fromEntries(
		classroomGroups.map(({ course, group }) => [String(course.id), group.id]),
	);

	function coursesInClassroom(classroomId: number) {
		return [
			...new Map(
				getClassroomCourseGroups(filteredCourses, classroomId).map(({ course }) => [
					course.id,
					course,
				]),
			).values(),
		];
	}

	function selectClassroom(classroomId: number) {
		selectedClassroomId = classroomId;
		window.history.replaceState({}, '', `/aulas?classroom=${classroomId}`);
	}

	function openEvent(event: PlannerEvent) {
		detailCourse = allCourses.find(({ id }) => id === event.courseId) ?? null;
		focusedSessionId = event.sessionId;
	}
</script>

<svelte:head><title>Aulas | Horarios</title></svelte:head>

<div class="flex min-h-dvh flex-col">
	<TopBar {user} {busy} {onNavigate} {onLogout} />
	<main
		id="main-content"
		class="mx-auto flex w-full max-w-[2400px] flex-1 flex-col gap-3 p-3 sm:p-4"
	>
		<header class="neo-panel flex flex-wrap items-end justify-between gap-3 px-4 py-3">
			<div>
				<p class="text-[10px] font-extrabold tracking-[0.18em] text-accent uppercase">
					{data?.termLabel ?? 'Periodo académico'}
				</p>
				<h1 class="mt-1 text-2xl font-extrabold text-primary">Horarios por aula</h1>
			</div>
			<label class="neo-control grid min-w-44"
				><span class="sr-only">Filtrar por año</span><select
					class="h-11 bg-transparent px-3 text-sm font-bold text-primary outline-none"
					bind:value={selectedYear}
					><option value={null}>Todos los años</option>{#each availableYears as year (year)}<option
							value={year}>Año {year}</option
						>{/each}</select
				></label
			>
		</header>

		{#if loadError}<section class="neo-panel p-6 text-sm font-semibold text-warning" role="alert">
				{loadError}
			</section>
		{:else if !data}<section
				class="neo-panel grid min-h-80 place-items-center text-sm font-semibold text-secondary"
			>
				Cargando aulas…
			</section>
		{:else}<div class="grid min-w-0 flex-1 gap-3 lg:grid-cols-[320px_minmax(0,1fr)]">
				<aside
					class="neo-panel course-island flex max-h-[720px] min-h-0 flex-col overflow-hidden p-3"
				>
					<div class="mb-2 flex items-center justify-between px-1">
						<h2 class="text-base font-extrabold text-primary">Aulas</h2>
						<span class="text-xs text-secondary">{classrooms.length}</span>
					</div>
					<div class="min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
						{#each classrooms as classroom (classroom.id)}
							{@const assignedCourses = coursesInClassroom(classroom.id)}
							<button
								class={`neo-card w-full border-l-[3px] p-3 text-left transition ${selectedClassroomId === classroom.id ? 'border-l-accent bg-accent-soft shadow-[var(--ui-shadow-pressed)]' : 'border-l-transparent'}`}
								type="button"
								on:click={() => selectClassroom(classroom.id)}
							>
								<span class="flex items-center justify-between gap-2"
									><strong class="text-sm text-primary">{classroom.code}</strong><span
										class="text-[10px] font-bold text-muted"
										>{classroom.type === 'THEORY' ? 'Teoría' : 'Lab'}</span
									></span
								>
								<span
									class="mt-2 block min-h-4 truncate text-xs text-secondary"
									title={assignedCourses.map(({ name }) => name).join(', ')}
									>{assignedCourses.length
										? assignedCourses.map(getCourseDisplayCode).join(' · ')
										: 'Sin cursos asignados'}</span
								>
							</button>
						{/each}
					</div>
				</aside>
				<section class="flex min-w-0 flex-col overflow-auto">
					<div class="mb-2 flex items-center justify-between gap-3 px-1">
						<div>
							<p class="text-[10px] font-bold tracking-[0.16em] text-muted uppercase">
								Aula seleccionada
							</p>
							<h2 class="text-lg font-extrabold text-primary">
								{selectedClassroom?.code ?? 'Sin aula'}
							</h2>
						</div>
						<span class="rounded-xl bg-accent-soft px-3 py-2 text-xs font-bold text-accent"
							>{classroomCourseCount} cursos</span
						>
					</div>
					<WeeklyPlanner
						days={data.days}
						academicHours={data.academicHours}
						{events}
						onOpenEvent={openEvent}
					/>
				</section>
			</div>{/if}
	</main>

	{#if detailCourse}<CourseDetailModal
			course={detailCourse}
			courses={allCourses}
			{selectedGroups}
			academicHours={data?.academicHours ?? []}
			{events}
			{focusedSessionId}
			onClose={() => {
				detailCourse = null;
				focusedSessionId = null;
			}}
		/>{/if}
</div>
