<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import { getDashboard, getSharedTimetable } from '../api/planner';
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
	let selection: Record<string, number> = {};
	let loadError = '';
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;

	onMount(async () => {
		try {
			const [dashboard, shared] = await Promise.all([getDashboard(), getSharedTimetable(shareId)]);
			data = dashboard;
			// ponytail: ids inexistentes se ignoran; la fuente de verdad es la oferta vigente.
			selection = {};
			for (const course of dashboard.courses) {
				const groupId = shared.selection[String(course.id)];
				if (groupId && course.groups.some((group) => group.id === groupId)) {
					selection[String(course.id)] = groupId;
				}
			}
		} catch (error) {
			loadError =
				error instanceof Error ? error.message : 'No se pudo cargar el horario compartido.';
		}
	});

	$: selectedCourseGroups = getSelectedCourseGroups(data?.courses ?? [], selection);
	$: ({ events, conflicts } = buildPlannerEvents(selectedCourseGroups));
	$: summary = buildPlannerSummary(selectedCourseGroups, conflicts);

	function openEvent(event: PlannerEvent) {
		const course = (data?.courses ?? []).find(({ id }) => id === event.courseId) ?? null;
		if (course) {
			detailCourse = course;
			focusedSessionId = event.sessionId;
		}
	}
</script>

<div class="flex min-h-dvh flex-col">
	<TopBar {user} {busy} {onNavigate} {onLogout} />

	<main
		class="mx-auto flex w-full max-w-[1880px] flex-1 flex-col gap-3 px-3 py-3 sm:px-4 sm:py-4 lg:h-[calc(100dvh-3.5rem)] lg:min-h-[32rem] lg:overflow-hidden xl:px-6"
	>
		<div
			class="flex flex-wrap items-center justify-between gap-2 rounded-[16px] border border-border-subtle bg-panel px-4 py-2 shadow-card backdrop-blur-xl"
		>
			<p class="text-[11px] text-secondary">
				<span class="font-bold text-primary">Horario compartido</span> · {data?.termLabel ?? ''}
				· <strong class="font-bold text-primary">{summary.selectedCourses}</strong> cursos ·
				<strong class={`font-bold ${summary.conflictCount > 0 ? 'text-warning' : 'text-primary'}`}
					>{summary.conflictCount}</strong
				>
				cruces
			</p>
			<button
				class="rounded-[10px] bg-accent-strong px-3 py-1.5 text-xs font-bold text-white transition hover:bg-accent"
				type="button"
				on:click={() => onNavigate('/')}
			>
				Crear mi horario
			</button>
		</div>

		{#if loadError}
			<div
				class="grid flex-1 place-items-center rounded-[24px] border border-dashed border-border-strong bg-panel p-8 text-center"
			>
				<div>
					<h1 class="font-display text-2xl font-bold text-primary">Enlace no disponible</h1>
					<p class="mt-2 text-sm text-secondary">{loadError}</p>
				</div>
			</div>
		{:else if !data}
			<div class="grid flex-1 place-items-center text-sm text-secondary">
				Cargando horario compartido…
			</div>
		{:else}
			<div class="flex min-h-[26rem] flex-1 flex-col lg:min-h-0">
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
		{@const detailGroup =
			detailCourse.groups.find((group) => group.id === selection[String(detailCourse?.id)]) ?? null}
		<CourseDetailModal
			course={detailCourse}
			selectedGroup={detailGroup}
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
