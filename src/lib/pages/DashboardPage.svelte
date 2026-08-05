<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import CourseExplorer from '../components/CourseExplorer.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import { getDashboard, createSharedTimetable } from '../api/planner';
	import { selection, toggleGroup, clearSelection, pruneSelection } from '../stores/selection';
	import {
		buildPlannerEvents,
		buildPlannerSummary,
		getSelectedCourseGroups,
		matchesCourseSearch,
	} from '../utils/planner';
	import type { Course, PlannerConflict, PlannerData, PlannerEvent } from '../types/planner';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let loadError = '';
	let searchQuery = '';
	let yearsSelected: Record<number, boolean> = {};
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;
	let sharing = false;
	let shareFeedback = '';

	onMount(async () => {
		try {
			data = await getDashboard();
			pruneSelection(data.courses);
		} catch {
			loadError = 'No se pudo cargar la oferta de cursos. Intenta de nuevo en un momento.';
		}
	});

	let events: PlannerEvent[] = [];
	let conflicts: PlannerConflict[] = [];

	$: allCourses = data?.courses ?? [];
	$: availableYears = [...new Set(allCourses.map((course) => course.academicYear))].sort();
	$: hasYearFilter = Object.values(yearsSelected).some(Boolean);
	$: filteredCourses = allCourses.filter(
		(course) =>
			(!hasYearFilter || yearsSelected[course.academicYear]) &&
			matchesCourseSearch(course, searchQuery),
	);
	$: selectedCourseGroups = getSelectedCourseGroups(allCourses, $selection);
	$: ({ events, conflicts } = buildPlannerEvents(selectedCourseGroups));
	$: summary = buildPlannerSummary(selectedCourseGroups, conflicts);

	function toggleYear(year: number) {
		yearsSelected = { ...yearsSelected, [year]: !yearsSelected[year] };
	}

	function openEvent(event: PlannerEvent) {
		const course = allCourses.find(({ id }) => id === event.courseId) ?? null;
		if (course) {
			detailCourse = course;
			focusedSessionId = event.sessionId;
		}
	}

	async function shareBoard() {
		if (sharing || summary.selectedCourses === 0) {
			return;
		}
		sharing = true;
		shareFeedback = '';
		try {
			const { id } = await createSharedTimetable($selection);
			await navigator.clipboard.writeText(`${window.location.origin}/s/${id}`);
			shareFeedback = 'Enlace copiado';
		} catch (error) {
			shareFeedback = error instanceof Error ? error.message : 'No se pudo crear el enlace';
		} finally {
			sharing = false;
		}
	}
</script>

<svelte:head>
	<title>Horarios | Uni Timetable</title>
</svelte:head>

<div class="flex min-h-dvh flex-col">
	<TopBar {user} {busy} {onNavigate} {onLogout} />

	<main
		class="mx-auto flex w-full max-w-[1880px] flex-1 flex-col gap-3 overflow-x-clip px-3 py-3 sm:px-4 sm:py-4 lg:h-[calc(100dvh-3.5rem)] lg:min-h-[32rem] lg:overflow-hidden xl:px-6"
	>
		{#if loadError}
			<div
				class="grid flex-1 place-items-center rounded-[24px] border border-dashed border-border-strong bg-panel p-8 text-center"
			>
				<p class="text-sm font-semibold text-warning">{loadError}</p>
			</div>
		{:else if !data}
			<div class="grid flex-1 place-items-center text-sm text-secondary">
				Cargando oferta académica…
			</div>
		{:else}
			<div
				class="grid min-h-0 w-full min-w-0 flex-1 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(300px,340px)]"
			>
				<div class="order-2 flex min-h-[26rem] min-w-0 flex-col lg:order-1 lg:min-h-0">
					<WeeklyPlanner
						days={data.days}
						academicHours={data.academicHours}
						{events}
						onOpenEvent={openEvent}
					/>
				</div>

				<div class="order-1 min-h-0 min-w-0 lg:order-2">
					<CourseExplorer
						termLabel={data.termLabel}
						{searchQuery}
						{yearsSelected}
						{availableYears}
						courses={filteredCourses}
						selectedGroups={$selection}
						{summary}
						{sharing}
						{shareFeedback}
						onSearchChange={(value) => (searchQuery = value)}
						onToggleYear={toggleYear}
						onToggleGroup={toggleGroup}
						onClearSelection={clearSelection}
						onOpenDetails={(course) => {
							detailCourse = course;
							focusedSessionId = null;
						}}
						onShare={shareBoard}
					/>
				</div>
			</div>
		{/if}
	</main>

	{#if detailCourse}
		{@const detailGroup =
			detailCourse.groups.find((group) => group.id === $selection[String(detailCourse?.id)]) ??
			null}
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
