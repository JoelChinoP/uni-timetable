<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import { getDashboard, getSharedTimetable } from '../api/planner';
	import { replaceSelection } from '../stores/selection';
	import {
		buildPlannerEvents,
		buildPlannerSummary,
		getSelectedCourseGroups,
	} from '../utils/planner';
	import type { Course, PlannerData, PlannerEvent } from '../types/planner';
	import type { AuthUser } from '../types/auth';

	export let shareId: string;
	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let sharedSelection: Record<string, number> = {};
	let loadError = '';
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;

	async function load() {
		loadError = '';
		try {
			const [dashboard, shared] = await Promise.all([getDashboard(), getSharedTimetable(shareId)]);
			data = dashboard;
			sharedSelection = {};
			for (const course of dashboard.courses) {
				const groupId = shared.selection[String(course.id)];
				if (groupId && course.groups.some((group) => group.id === groupId))
					sharedSelection[String(course.id)] = groupId;
			}
		} catch (error) {
			loadError =
				error instanceof Error ? error.message : 'No se pudo cargar el horario compartido.';
		}
	}

	onMount(() => void load());
	$: selectedCourseGroups = getSelectedCourseGroups(data?.courses ?? [], sharedSelection);
	$: ({ events, conflicts } = buildPlannerEvents(selectedCourseGroups));
	$: summary = buildPlannerSummary(selectedCourseGroups, conflicts);

	function openEvent(event: PlannerEvent) {
		const course = (data?.courses ?? []).find(({ id }) => id === event.courseId) ?? null;
		if (course) {
			detailCourse = course;
			focusedSessionId = event.sessionId;
		}
	}

	function useTimetable() {
		replaceSelection(sharedSelection);
		onNavigate('/');
	}
</script>

<svelte:head><title>Horario compartido | Horarios</title></svelte:head>

<div class="flex min-h-dvh flex-col lg:h-dvh lg:overflow-hidden">
	<TopBar {user} {busy} {onNavigate} {onLogout} />
	<main
		id="main-content"
		class="mx-auto flex min-h-0 w-full max-w-[1680px] flex-1 flex-col gap-3 px-3 py-3 sm:px-4 lg:overflow-hidden xl:px-5"
	>
		<section
			class="glass-panel flex flex-col gap-3 rounded-2xl px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
		>
			<div>
				<p class="text-[10px] font-extrabold tracking-[0.18em] text-accent uppercase">
					Horario compartido · {data?.termLabel ?? ''}
				</p>
				<p class="mt-1 text-sm text-secondary">
					<strong class="text-primary">{summary.selectedCourses}</strong> cursos ·
					<strong class={summary.conflictCount > 0 ? 'text-warning' : 'text-success'}
						>{summary.conflictCount > 0 ? `${summary.conflictCount} cruces` : 'sin cruces'}</strong
					>
				</p>
			</div>
			<div class="flex gap-2">
				<button
					class="neo-button min-h-11 px-3 text-xs font-bold text-primary"
					type="button"
					on:click={() => onNavigate('/')}>Crear uno nuevo</button
				><button
					class="min-h-11 rounded-xl bg-accent-strong px-3 text-xs font-bold text-white transition hover:bg-accent disabled:opacity-50"
					type="button"
					disabled={summary.selectedCourses === 0}
					on:click={useTimetable}>Usar y editar</button
				>
			</div>
		</section>

		{#if loadError}
			<section class="neo-panel grid flex-1 place-items-center p-8 text-center" role="alert">
				<div>
					<h1 class="text-2xl font-extrabold text-primary">Enlace no disponible</h1>
					<p class="mt-2 text-sm text-secondary">{loadError}</p>
					<button
						class="mt-5 h-11 rounded-xl bg-accent-strong px-4 text-sm font-bold text-white"
						type="button"
						on:click={load}>Reintentar</button
					>
				</div>
			</section>
		{:else if !data}
			<section
				class="neo-panel grid flex-1 place-items-center text-sm font-semibold text-secondary"
				aria-live="polite"
			>
				Cargando horario compartido…
			</section>
		{:else}
			<div class="flex min-h-[34rem] flex-1 flex-col lg:min-h-0">
				<WeeklyPlanner
					days={data.days}
					academicHours={data.academicHours}
					{events}
					onOpenEvent={openEvent}
				/>
			</div>
		{/if}
	</main>

	{#if detailCourse}
		<CourseDetailModal
			course={detailCourse}
			courses={data?.courses ?? []}
			selectedGroups={sharedSelection}
			academicHours={data?.academicHours ?? []}
			{events}
			{focusedSessionId}
			onClose={() => {
				detailCourse = null;
				focusedSessionId = null;
			}}
		/>
	{/if}
</div>
